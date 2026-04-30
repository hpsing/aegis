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
  KEEPERHUB_API_KEY="$vkey" \
  KEEPERHUB_VERIFIER_INDEX="$i" \
  AEGIS_RPC="$RPC_URL" \
  AEGIS_CONTRACT="$AEGIS_ADDR" \
  REGISTRY_CONTRACT="$REGISTRY_ADDR" \
  VERIFIER_PRIVATE_KEY="$vpk" \
  SWAP_RPC="$SWAP_RPC" \
  OG_PRIVATE_KEY="$vpk" \
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

show_progress() {
  local s; s=$(read_status)
  local label; label=$(status_label "$s")
  local nrev; nrev=$(cast call "$AEGIS_ADDR" "revealCount(uint256)(uint256)" "$JOB_ID" --rpc-url "$RPC_URL" 2>/dev/null | head -1 | awk '{print $1}')
  echo "  job status=$label  reveals=${nrev:-0}"
}

echo "=== watching commit + reveal phases (≈10 min) ==="
echo "  tail logs in another terminal: tail -f $LOGDIR/verifier-*.log"
for elapsed in 10 20 30 40 50 60 70; do
  sleep 10
  printf -- "--- t+%ds ---\n" "$elapsed"
  show_progress
done


# ---------- 6. settle ----------
echo
echo "=== calling settle ==="
SETTLE_OUT=$(cast send "$AEGIS_ADDR" "settle(uint256)" "$JOB_ID" \
  --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy 2>&1)
SETTLE_TX=$(echo "$SETTLE_OUT" | grep -oE '0x[a-f0-9]{64}' | head -1)
echo "$SETTLE_OUT" | tail -3
sleep 8

# ---------- 7. event extraction (commits, reveals) ----------
EXPLORER="${EXPLORER:-https://chainscan-galileo.0g.ai}"

# event signatures
COMMIT_SIG=$(cast keccak "VoteCommitted(uint256,address,bytes32)")
REVEAL_SIG=$(cast keccak "VoteRevealed(uint256,address,bool)")
JOB_TOPIC=$(printf '0x%064x' "$JOB_ID")

# Scan from the postJob block (captured at step 2) to head. Anchoring on the
# job's actual block guarantees we cover its full lifecycle regardless of
# wall-clock demo length, vs. a fixed `head - N` window that races the demo.
HEAD_HEX=$(cast block-number --rpc-url "$RPC_URL" 2>/dev/null)
HEAD_DEC=$((HEAD_HEX))
FROM_BLOCK=${POST_BLOCK_DEC:-$((HEAD_DEC - 600))}
[[ $FROM_BLOCK -lt 0 ]] && FROM_BLOCK=0

echo
echo "scanning blocks $FROM_BLOCK..$HEAD_DEC for VoteCommitted/VoteRevealed (jobId=$JOB_ID)..."

# 0G's eth_getLogs is finicky with the (--address + topics) combo: filtering
# by address sometimes returns empty even when the events exist. Topic-only
# filtering (sig + jobId) is reliable; we re-check the address client-side.
logs_for() {
  local sig="$1"
  cast logs --rpc-url "$RPC_URL" \
    --from-block "$FROM_BLOCK" --to-block "$HEAD_DEC" \
    "$sig" "$JOB_TOPIC" --json 2>/dev/null \
    | jq --arg a "$AEGIS_ADDR" '[.[] | select((.address // "") | ascii_downcase == ($a | ascii_downcase))]'
}

COMMIT_LOGS=$(logs_for "$COMMIT_SIG")
REVEAL_LOGS=$(logs_for "$REVEAL_SIG")

# pull out (verifier, txHash) pairs
parse_logs() {
  echo "$1" | jq -r '.[]? | "\(.topics[2] | sub("^0x000000000000000000000000"; "0x"))  \(.transactionHash)"'
}

COMMITS=$(parse_logs "$COMMIT_LOGS")
REVEALS=$(parse_logs "$REVEAL_LOGS")

# Proof-bundle roots scraped from verifier logs. The verifier logs each
# upload as `proof bundle <jobID> -> 0g://0x<64-hex>` (see
# internal/verifier/loop.go and internal/ogstorage/og_service.go).
proofbundle_roots() {
  for i in 1 2 3; do
    local addr root
    addr=$(grep -oE 'addr=0x[a-fA-F0-9]{40}' "$LOGDIR/verifier-$i.log" | head -1 | cut -d= -f2)
    root=$(grep -oE '0g://0x[a-fA-F0-9]{64}' "$LOGDIR/verifier-$i.log" | head -1 | sed 's|^0g://||')
    echo "  v$i ${addr:-?}  proofbundle_root=${root:-(not found in log; check 0G storage uploads)}"
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
SWAP_TX=$(grep -oE 'tx=0x[a-f0-9]{64} amount_out' "$LOGDIR/executor.log" | head -1 | grep -oE '0x[a-f0-9]{64}')
[[ -n "$SWAP_TX" ]]   && echo "  swap (synth.): $EXPLORER/tx/$SWAP_TX"
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
proofbundle_roots

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
