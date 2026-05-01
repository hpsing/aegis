#!/usr/bin/env bash
# Aegis end-to-end demo: publisher posts a job → 3 verifiers commit + reveal
# through KeeperHub workflows → executor submits claim → settle.
#
# Total wall-clock: ~10 min (5 min commit window + 5 min reveal window).
#
# Required env:
#   TREASURY_PK         private key of 0x7C9DcA…
#   V1_PK / V2_PK / V3_PK   verifier private keys (must match KH wallets)
#   EXECUTOR_PK         private key of a fresh executor wallet
#   ORG1_API_KEY / ORG2_API_KEY / ORG3_API_KEY   per-org KH keys
#
# Optional env:
#   RPC_URL          default https://evmrpc-testnet.0g.ai
#   AEGIS_ADDR       default 0xa89833fBD1844763cc77C0a3aFaE32697A2F990f
#   USDC_ADDR        default 0xe9dA98EB0AF68cC48be7F71C29A7Bc5bA7fB45Eb
#   REGISTRY_ADDR    default 0x50ce23AE35bbe43fFAd0B36FD3F567560b8EfB18
#   SWAP_RPC         default same as RPC_URL (synthetic-claim mode)
#   SKIP_PREFLIGHT   set to 1 to skip USDC mint / executor funding
#   SETTLE_BY        treasury (default) | manual
#
# Run:
#   export TREASURY_PK=0x... V1_PK=0x... V2_PK=0x... V3_PK=0x... EXECUTOR_PK=0x...
#   export ORG1_API_KEY=... ORG2_API_KEY=... ORG3_API_KEY=...
#   bash scripts/e2e-demo.sh

set -uo pipefail

: "${TREASURY_PK:?set TREASURY_PK}"
: "${V1_PK:?set V1_PK}"
: "${V2_PK:?set V2_PK}"
: "${V3_PK:?set V3_PK}"
: "${EXECUTOR_PK:?set EXECUTOR_PK}"
: "${ORG1_API_KEY:?set ORG1_API_KEY}"
: "${ORG2_API_KEY:?set ORG2_API_KEY}"
: "${ORG3_API_KEY:?set ORG3_API_KEY}"

RPC_URL="${RPC_URL:-https://evmrpc-testnet.0g.ai}"
AEGIS_ADDR="${AEGIS_ADDR:-0xa89833fBD1844763cc77C0a3aFaE32697A2F990f}"
USDC_ADDR="${USDC_ADDR:-0xe9dA98EB0AF68cC48be7F71C29A7Bc5bA7fB45Eb}"
REGISTRY_ADDR="${REGISTRY_ADDR:-0x50ce23AE35bbe43fFAd0B36FD3F567560b8EfB18}"
SWAP_RPC="${SWAP_RPC:-$RPC_URL}"
SKIP_PREFLIGHT="${SKIP_PREFLIGHT:-0}"

LOGDIR="logs/e2e-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$LOGDIR"

TREASURY_ADDR=$(cast wallet address --private-key "$TREASURY_PK")
EXECUTOR_ADDR=$(cast wallet address --private-key "$EXECUTOR_PK")

echo "============================================================"
echo "Aegis E2E Demo"
echo "============================================================"
echo "  treasury:  $TREASURY_ADDR"
echo "  executor:  $EXECUTOR_ADDR"
echo "  aegis:     $AEGIS_ADDR"
echo "  rpc:       $RPC_URL"
echo "  logs:      $LOGDIR/"
echo

echo "=== killing stale aegis processes ==="
# Zombie verifiers from a prior run (Ctrl-C / terminal close skipped the
# cleanup trap) keep polling /recv on the AXL daemons we're about to
# restart and steal spec_publish messages from the new verifiers,
# causing them to abstain. The on-chain commits/reveals from zombies
# also confuse the e2e because they're done with stale code paths.
for pat in "exe/verifier --label" "exe/publisher" "exe/executor"; do
  pids=$(pgrep -f "$pat" 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "  killing: $pat → $pids"
    kill $pids 2>/dev/null || true
  fi
done
sleep 2
echo

echo "=== AXL swarm ==="
bash scripts/run-axl-swarm.sh
AXL_PUB_URL="http://127.0.0.1:9002"
AXL_V1_URL="http://127.0.0.1:9012"
AXL_V2_URL="http://127.0.0.1:9022"
AXL_V3_URL="http://127.0.0.1:9032"
if [[ ! -f .axl/run/peer-v1 || ! -f .axl/run/peer-v2 || ! -f .axl/run/peer-v3 ]]; then
  echo "ERROR: AXL swarm did not produce peer ids for v1/v2/v3" >&2
  exit 1
fi
AXL_VERIFIER_PEERS="$(cat .axl/run/peer-v1),$(cat .axl/run/peer-v2),$(cat .axl/run/peer-v3)"
echo "  publisher → ${AXL_VERIFIER_PEERS:0:30}…"

# ---------- 1. preflight ----------
if [[ "$SKIP_PREFLIGHT" != "1" ]]; then
  echo "=== preflight ==="

  treasury_usdc=$(cast call "$USDC_ADDR" "balanceOf(address)(uint256)" "$TREASURY_ADDR" --rpc-url "$RPC_URL" 2>/dev/null | head -1 | awk '{print $1}')
  treasury_usdc="${treasury_usdc:-0}"
  echo "  treasury USDC: $treasury_usdc"
  if (( treasury_usdc < 200000000 )); then
    echo "    minting 200 mUSDC to treasury..."
    cast send "$USDC_ADDR" "mint(address,uint256)" "$TREASURY_ADDR" 200000000 \
      --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
  fi

  echo "    approving aegis to pull 200 mUSDC from treasury..."
  cast send "$USDC_ADDR" "approve(address,uint256)" "$AEGIS_ADDR" 200000000 \
    --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true

  exec_bal_hex=$(cast balance "$EXECUTOR_ADDR" --rpc-url "$RPC_URL" 2>/dev/null)
  echo "  executor 0G:   $exec_bal_hex"
  if [[ "$exec_bal_hex" == "0" ]]; then
    echo "    funding executor with 0.05 0G..."
    cast send "$EXECUTOR_ADDR" --value 0.05ether \
      --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
  fi
  echo
fi

# ---------- 2. publisher posts job ----------
echo "=== publisher posts job ==="
TX_HASH=$(AEGIS_RPC="$RPC_URL" \
  AEGIS_CONTRACT="$AEGIS_ADDR" \
  USDC_CONTRACT="$USDC_ADDR" \
  CLIENT_PRIVATE_KEY="$TREASURY_PK" \
  EXECUTOR_ADDRESS="$EXECUTOR_ADDR" \
  AXL_NODE_URL="$AXL_PUB_URL" \
  AXL_VERIFIER_PEERS="$AXL_VERIFIER_PEERS" \
  go run ./cmd/publisher 2>&1 | tee "$LOGDIR/publisher.log" | grep -oE '0x[a-f0-9]{64}' | head -1)

if [[ -z "$TX_HASH" ]]; then
  echo "  ERROR: no tx hash from publisher. Inspect $LOGDIR/publisher.log"
  exit 1
fi
echo "  tx_hash=$TX_HASH"

echo "  waiting 8s for receipt..."
sleep 8

JOB_POSTED_SIG=$(cast keccak "JobPosted(uint256,address,address,bytes32,uint256,uint256,uint256)")
RECEIPT_JSON=$(cast receipt "$TX_HASH" --rpc-url "$RPC_URL" --json 2>/dev/null)
JOB_ID_HEX=$(echo "$RECEIPT_JSON" \
  | jq -r --arg sig "$JOB_POSTED_SIG" '.logs[]? | select(.topics[0] == $sig) | .topics[1]' \
  | head -1)
# Capture postJob block — used later as FROM_BLOCK for the commit/reveal log
# scan so we always cover this job's full lifecycle regardless of demo length.
POST_BLOCK_HEX=$(echo "$RECEIPT_JSON" | jq -r '.blockNumber // empty')
POST_BLOCK_DEC=$((POST_BLOCK_HEX))

if [[ -z "$JOB_ID_HEX" ]]; then
  echo "  ERROR: JobPosted event not found in receipt. Inspect via:"
  echo "    cast receipt $TX_HASH --rpc-url $RPC_URL"
  exit 1
fi
JOB_ID=$(printf '%d' "$JOB_ID_HEX")
echo "  JOB_ID=$JOB_ID  postJob_block=$POST_BLOCK_DEC"
echo

# ---------- 3. spawn 3 verifiers (KH-routed) ----------
echo "=== starting 3 verifiers (KH-routed) ==="
declare -a VPIDS=()
for i in 1 2 3; do
  vpk_var="V${i}_PK"; vpk="$(eval echo \$$vpk_var)"
  vkey_var="ORG${i}_API_KEY"; vkey="$(eval echo \$$vkey_var)"

  # Each verifier writes its own ProofBundle to 0G Storage; OG_PRIVATE_KEY
  # = the verifier's own key so uploads are signed by the same identity
  # the verifier reveals as. (Each verifier wallet was funded with 0.05 0G
  # — covers a few uploads.)
    axl_url=""
  case "$i" in
    1) axl_url="$AXL_V1_URL" ;;
    2) axl_url="$AXL_V2_URL" ;;
    3) axl_url="$AXL_V3_URL" ;;
  esac
  # AXL_REPLICATE_PEERS lets each verifier replicate its vote envelope
  # to its peers (the OTHER verifier nodes). We send to the union of
  # peer ids; each verifier ignores its own.
  axl_replicate="$AXL_VERIFIER_PEERS"
  KEEPERHUB_API_KEY="$vkey" \
  KEEPERHUB_VERIFIER_INDEX="$i" \
  AEGIS_RPC="$RPC_URL" \
  AEGIS_CONTRACT="$AEGIS_ADDR" \
  REGISTRY_CONTRACT="$REGISTRY_ADDR" \
  VERIFIER_PRIVATE_KEY="$vpk" \
  SWAP_RPC="$SWAP_RPC" \
  OG_PRIVATE_KEY="$vpk" \
  AXL_NODE_URL="$axl_url" \
  AXL_REPLICATE_PEERS="$axl_replicate" \
  go run ./cmd/verifier --label "v$i" >"$LOGDIR/verifier-$i.log" 2>&1 &
  pid=$!
  VPIDS+=("$pid")
  echo "  v$i pid=$pid log=$LOGDIR/verifier-$i.log"
done

cleanup() {
  echo "stopping verifiers..."
  for pid in "${VPIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true
}
trap cleanup EXIT

echo "  waiting 12s for verifiers to subscribe..."
sleep 12
echo

# ---------- 4. executor submits claim ----------
echo "=== executor submits claim (mode=honest) ==="
AEGIS_RPC="$RPC_URL" \
AEGIS_CONTRACT="$AEGIS_ADDR" \
EXECUTOR_PRIVATE_KEY="$EXECUTOR_PK" \
JOB_ID="$JOB_ID" \
CLIENT_ADDRESS="$TREASURY_ADDR" \
go run ./cmd/executor --mode=honest 2>&1 | tee "$LOGDIR/executor.log"
echo

# ---------- 5. watch the flow ----------
status_label() {
  case "$1" in
    0) echo "None";;
    1) echo "Posted";;
    2) echo "ClaimSubmitted";;
    3) echo "Settled";;
    4) echo "Cancelled";;
    *) echo "Unknown($1)";;
  esac
}

# Job struct: client, executor, reimbursement, fee, bounty, specHash, txHash,
# reportedOutcomeHash, claimDeadline, commitDeadline, revealDeadline, status
read_status() {
  local raw
  raw=$(cast call "$AEGIS_ADDR" \
    "jobs(uint256)(address,address,uint256,uint256,uint256,bytes32,bytes32,bytes32,uint64,uint64,uint64,uint8)" \
    "$JOB_ID" --rpc-url "$RPC_URL" 2>/dev/null)
  echo "$raw" | tail -1 | tr -d ' '
}

read_reveals() {
  cast call "$AEGIS_ADDR" "revealCount(uint256)(uint256)" "$JOB_ID" --rpc-url "$RPC_URL" 2>/dev/null | head -1 | awk '{print $1}'
}

# read_reveal_deadline returns the job's revealDeadline (uint64 unix
# seconds). Reads the full Job struct and pulls the 11th field.
read_reveal_deadline() {
  cast call "$AEGIS_ADDR" \
    "jobs(uint256)(address,address,uint256,uint256,uint256,bytes32,bytes32,bytes32,uint64,uint64,uint64,uint8)" \
    "$JOB_ID" --rpc-url "$RPC_URL" 2>/dev/null \
    | sed -n '11p' | awk '{print $1}'
}

# read_chain_now returns the latest block's timestamp (Unix s). Settle
# checks block.timestamp, NOT wall-clock — chain may lag wall. cast
# block (text mode) prints "timestamp  <decimal> (...)"; we take col 2.
read_chain_now() {
  cast block latest --rpc-url "$RPC_URL" 2>/dev/null \
    | awk '$1=="timestamp"{print $2; exit}'
}

show_progress() {
  local s; s=$(read_status)
  local label; label=$(status_label "$s")
  local nrev; nrev=$(read_reveals)
  echo "  job status=$label  reveals=${nrev:-0}"
}

# Wait for all 3 reveals to land before calling settle. Short-circuits
# once revealCount==3 instead of grinding through the full 70s window
# every run. Caps at ~10 min as a safety net for chain stalls.
echo "=== watching commit + reveal phases (until reveals=3, max 600s) ==="
echo "  tail logs in another terminal: tail -f $LOGDIR/verifier-*.log"
START_TS=$(date +%s)
elapsed=0
while (( elapsed < 600 )); do
  sleep 10
  elapsed=$(( $(date +%s) - START_TS ))
  printf -- "--- t+%ds ---\n" "$elapsed"
  show_progress
  nrev=$(read_reveals)
  if [[ "${nrev:-0}" == "3" ]]; then
    echo "  all 3 reveals landed — waiting for revealDeadline before settle"
    break
  fi
done


# ---------- 6. settle ----------
# settle() requires block.timestamp > revealDeadline. Even with all 3
# reveals in, calling settle before the deadline reverts with
# DeadlineNotPassed. Wait until chain time is past revealDeadline
# (+5s buffer for clock skew) before dispatching.
RD=$(read_reveal_deadline)
if [[ -n "$RD" ]]; then
  while :; do
    NOW=$(read_chain_now)
    if [[ -n "$NOW" ]] && (( NOW > RD )); then
      break
    fi
    if [[ -n "$NOW" ]]; then
      printf -- "  chain time %s, revealDeadline %s, waiting %ds...\n" "$NOW" "$RD" $(( RD - NOW + 5 ))
    fi
    sleep 5
  done
fi

echo
echo "=== calling settle ==="
# 0G Galileo's RPC frequently returns partial/null responses for cast
# send (no transactionHash field). Don't trust its stdout; treat the
# call as best-effort and recover the real tx hash from the JobSettled
# event log below.
cast send "$AEGIS_ADDR" "settle(uint256)" "$JOB_ID" \
  --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
echo "  settle dispatched — waiting for JobSettled event..."

# ---------- 7. event extraction (commits, reveals, settle) ----------
EXPLORER="${EXPLORER:-https://chainscan-galileo.0g.ai}"

# event signatures
COMMIT_SIG=$(cast keccak "VoteCommitted(uint256,address,bytes32)")
REVEAL_SIG=$(cast keccak "VoteRevealed(uint256,address,bool)")
SETTLE_SIG=$(cast keccak "JobSettled(uint256,bool,uint256,uint256)")
JOB_TOPIC=$(printf '0x%064x' "$JOB_ID")

# logs_for_range scans a fixed range. 0G's eth_getLogs is finicky with the
# (--address + topics) combo: filtering by address sometimes returns empty
# even when the events exist. Topic-only filtering (sig + jobId) is
# reliable; we re-check the address client-side.
logs_for_range() {
  local sig="$1" from="$2" to="$3"
  cast logs --rpc-url "$RPC_URL" \
    --from-block "$from" --to-block "$to" \
    "$sig" "$JOB_TOPIC" --json 2>/dev/null \
    | jq --arg a "$AEGIS_ADDR" '[.[] | select((.address // "") | ascii_downcase == ($a | ascii_downcase))]'
}

# Poll for JobSettled up to ~90s. Once it lands, anchor the scan window
# to its block so the upper bound covers settle and we know exactly when
# to read commits/reveals.
SETTLE_TX=""
SETTLE_BLOCK=""
for _ in $(seq 1 18); do
  sleep 5
  HEAD_HEX=$(cast block-number --rpc-url "$RPC_URL" 2>/dev/null)
  HEAD_DEC=$((HEAD_HEX))
  FROM_BLOCK=${POST_BLOCK_DEC:-$((HEAD_DEC - 600))}
  [[ $FROM_BLOCK -lt 0 ]] && FROM_BLOCK=0
  SETTLE_LOGS=$(logs_for_range "$SETTLE_SIG" "$FROM_BLOCK" "$HEAD_DEC")
  SETTLE_TX=$(echo "$SETTLE_LOGS" | jq -r '.[0].transactionHash // empty')
  SETTLE_BLOCK=$(echo "$SETTLE_LOGS" | jq -r '.[0].blockNumber // empty')
  if [[ -n "$SETTLE_TX" ]]; then
    break
  fi
done

if [[ -z "$SETTLE_TX" ]]; then
  echo "  WARN: JobSettled not found within 90s. Falling back to head."
  HEAD_HEX=$(cast block-number --rpc-url "$RPC_URL" 2>/dev/null)
  HEAD_DEC=$((HEAD_HEX))
  FROM_BLOCK=${POST_BLOCK_DEC:-$((HEAD_DEC - 600))}
  [[ $FROM_BLOCK -lt 0 ]] && FROM_BLOCK=0
fi

echo
echo "scanning blocks $FROM_BLOCK..$HEAD_DEC for VoteCommitted/VoteRevealed (jobId=$JOB_ID)..."

COMMIT_LOGS=$(logs_for_range "$COMMIT_SIG" "$FROM_BLOCK" "$HEAD_DEC")
REVEAL_LOGS=$(logs_for_range "$REVEAL_SIG" "$FROM_BLOCK" "$HEAD_DEC")

# pull out (verifier, txHash) pairs
parse_logs() {
  echo "$1" | jq -r '.[]? | "\(.topics[2] | sub("^0x000000000000000000000000"; "0x"))  \(.transactionHash)"'
}

COMMITS=$(parse_logs "$COMMIT_LOGS")
REVEALS=$(parse_logs "$REVEAL_LOGS")

# Proof-bundle roots scraped from verifier logs. The verifier logs each
# upload as `proof bundle <jobID> -> 0g://0x<64-hex>` (see
# internal/verifier/loop.go and internal/ogstorage/og_service.go). The
# root is a 0G Storage merkle root, NOT an EVM tx hash — it lives on
# the storage layer, viewable via the 0G Storage gateway.
STORAGE_GATEWAY="${STORAGE_GATEWAY:-https://storagescan-galileo.0g.ai}"
proofbundle_roots() {
  for i in 1 2 3; do
    local addr root
    addr=$(grep -oE 'addr=0x[a-fA-F0-9]{40}' "$LOGDIR/verifier-$i.log" | head -1 | cut -d= -f2)
    root=$(grep -oE '0g://0x[a-fA-F0-9]{64}' "$LOGDIR/verifier-$i.log" | head -1 | sed 's|^0g://||')
    if [[ -n "$root" ]]; then
      echo "  v$i ${addr:-?}"
      echo "      root: $root"
      echo "      view: $STORAGE_GATEWAY/tx/$root"
    else
      echo "  v$i ${addr:-?}  (no proofbundle root found in log)"
    fi
  done
}

# ---------- 8. final state ----------
echo
echo "============================================================"
echo "Final state — Job $JOB_ID"
echo "============================================================"
show_progress
echo

echo "On-chain transactions"
echo "---------------------"
echo "  postJob:      $EXPLORER/tx/$TX_HASH"
EXEC_TX=$(grep -oE 'submitClaim tx=0x[a-f0-9]{64}' "$LOGDIR/executor.log" | head -1 | cut -d= -f2)
[[ -n "$EXEC_TX" ]]   && echo "  submitClaim:  $EXPLORER/tx/$EXEC_TX"

echo
echo "Verifier commits (msg.sender = each verifier's own wallet)"
echo "---------------------"
if [[ -n "$COMMITS" ]]; then
  while IFS=' ' read -r addr txh; do
    [[ -z "$addr" ]] && continue
    echo "  $addr  $EXPLORER/tx/$txh"
  done <<< "$COMMITS"
else
  echo "  (no logs found in scanned range)"
fi

echo
echo "Verifier reveals"
echo "---------------------"
if [[ -n "$REVEALS" ]]; then
  while IFS=' ' read -r addr txh; do
    [[ -z "$addr" ]] && continue
    echo "  $addr  $EXPLORER/tx/$txh"
  done <<< "$REVEALS"
else
  echo "  (no logs found in scanned range)"
fi

echo
echo "Settle"
echo "---------------------"
if [[ -n "$SETTLE_TX" ]]; then
  echo "  settle:       $EXPLORER/tx/$SETTLE_TX"
else
  echo "  (settle tx hash not captured)"
fi

echo
echo "0G Storage — ProofBundles"
echo "-------------------------"
echo "  (gateway: $STORAGE_GATEWAY — override with STORAGE_GATEWAY=...)"
proofbundle_roots

echo
echo "0G Storage — vote-history append entries"
echo "----------------------------------------"
for i in 1 2 3; do
  addr=$(grep -oE 'addr=0x[a-fA-F0-9]{40}' "$LOGDIR/verifier-$i.log" | head -1 | cut -d= -f2)
  vote_root=$(grep -oE 'storage append job=[0-9]+ entry=0g://0x[a-fA-F0-9]{64}' "$LOGDIR/verifier-$i.log" | head -1 | grep -oE '0x[a-fA-F0-9]{64}')
  if [[ -n "$vote_root" ]]; then
    echo "  v$i ${addr:-?}"
    echo "      root: $vote_root"
    echo "      view: $STORAGE_GATEWAY/tx/$vote_root"
  else
    echo "  v$i ${addr:-?}  (no vote-record append found in log)"
  fi
done

echo
echo "KeeperHub audit trail"
echo "---------------------"
echo "  Each commit + reveal fires execute_workflow asynchronously into"
echo "  KH for audit. Sign in to each org's dashboard to inspect:"
echo "    org-1 (verifier-1):  https://app.keeperhub.com/  → workflow 'aegis-commit-vote', 'aegis-reveal-vote'"
echo "    org-2 (verifier-2):  https://app.keeperhub.com/  → same"
echo "    org-3 (verifier-3):  https://app.keeperhub.com/  → same"
echo "  Each successful job emits 6 KH workflow invocations (3 commit + 3 reveal)."

echo
echo "Logs: $LOGDIR/"
echo "  publisher.log  executor.log  verifier-{1,2,3}.log"
