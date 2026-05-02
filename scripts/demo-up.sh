#!/usr/bin/env bash
# demo-up.sh — boots everything for the Aegis live UI demo and leaves
# it running in the background. Idempotent + safe to re-run.
#
# Boot order:
#   1. kill any stale aegis processes (zombies steal AXL /recv messages)
#   2. start the AXL swarm (4 daemons; pub + v1/v2/v3 sharing tcp_port 7001)
#   3. preflight: mint USDC to treasury, approve aegis, fund executor
#   4. spawn 3 verifier processes (long-running, one per org)
#   5. spawn cmd/ui (binds :3000 by default)
#   6. wait until the UI's /api/state shows all 3 verifier AXL peers
#
# Stop everything with: bash scripts/demo-down.sh
# Post a fresh job: bash scripts/demo-post-job.sh
#
# Env (same as e2e-demo.sh — single source of truth):
#   TREASURY_PK / V1_PK / V2_PK / V3_PK / EXECUTOR_PK
#   ORG1_API_KEY / ORG2_API_KEY / ORG3_API_KEY
#
# Optional:
#   UI_ADDR          UI listen addr (default :3000)
#   RPC_URL          default https://evmrpc-testnet.0g.ai
#   SKIP_PREFLIGHT   set to 1 to skip USDC + executor funding

set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$(pwd)"

: "${TREASURY_PK:?set TREASURY_PK}"
: "${V1_PK:?set V1_PK}"
: "${V2_PK:?set V2_PK}"
: "${V3_PK:?set V3_PK}"
: "${EXECUTOR_PK:?set EXECUTOR_PK}"
: "${ORG1_API_KEY:?set ORG1_API_KEY}"
: "${ORG2_API_KEY:?set ORG2_API_KEY}"
: "${ORG3_API_KEY:?set ORG3_API_KEY}"

UI_ADDR="${UI_ADDR:-:3000}"
RPC_URL="${RPC_URL:-https://evmrpc-testnet.0g.ai}"
AEGIS_ADDR="${AEGIS_ADDR:-0xa89833fBD1844763cc77C0a3aFaE32697A2F990f}"
USDC_ADDR="${USDC_ADDR:-0xe9dA98EB0AF68cC48be7F71C29A7Bc5bA7fB45Eb}"
REGISTRY_ADDR="${REGISTRY_ADDR:-0x50ce23AE35bbe43fFAd0B36FD3F567560b8EfB18}"
SWAP_RPC="${SWAP_RPC:-$RPC_URL}"
SKIP_PREFLIGHT="${SKIP_PREFLIGHT:-0}"

LOGDIR="logs/demo-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$LOGDIR" .demo

TREASURY_ADDR=$(cast wallet address --private-key "$TREASURY_PK")
EXECUTOR_ADDR=$(cast wallet address --private-key "$EXECUTOR_PK")

echo "============================================================"
echo "Aegis demo — boot"
echo "============================================================"
echo "  treasury:  $TREASURY_ADDR"
echo "  executor:  $EXECUTOR_ADDR"
echo "  ui:        http://localhost${UI_ADDR}"
echo "  logs:      $LOGDIR/"
echo

# ---------- 1. kill stale processes ----------
# Patterns chosen to catch BOTH variants:
#   - `go run ./cmd/verifier --label v1` (the wrapper we spawn)
#   - `/Users/singh/Library/Caches/go-build/<hash>/verifier --label v1` (the
#     compiled binary go-run actually exec's; orphaned when the wrapper dies)
# The second case is the zombie source — go-run's signal handling
# doesn't always propagate to the child, so each demo-up was leaving
# behind compiled-binary verifiers that kept committing with the same
# wallet, causing AlreadyCommitted reverts on the next run.
echo "=== killing stale aegis processes ==="
for pat in "/verifier --label" "go run ./cmd/verifier" "go run ./cmd/ui" "exe/publisher" "exe/executor" "bin/ui"; do
  pids=$(pgrep -f "$pat" 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "  killing: $pat → $pids"
    kill -9 $pids 2>/dev/null || true
  fi
done
sleep 2
# Defensive double-check; if anything still alive matching the broad
# verifier pattern, abort rather than start more.
remaining=$(pgrep -f '/verifier --label' 2>/dev/null || true)
if [[ -n "$remaining" ]]; then
  echo "ERROR: verifier processes survived kill: $remaining" >&2
  echo "       try: pkill -9 -f '/verifier --label'" >&2
  exit 1
fi

# ---------- 2. AXL swarm ----------
echo
echo "=== AXL swarm ==="
bash scripts/run-axl-swarm.sh
AXL_PUB_URL="http://127.0.0.1:9002"
AXL_V1_URL="http://127.0.0.1:9012"
AXL_V2_URL="http://127.0.0.1:9022"
AXL_V3_URL="http://127.0.0.1:9032"
if [[ ! -f .axl/run/peer-v1 || ! -f .axl/run/peer-v2 || ! -f .axl/run/peer-v3 ]]; then
  echo "ERROR: AXL swarm did not produce peer ids" >&2
  exit 1
fi
AXL_VERIFIER_PEERS="$(cat .axl/run/peer-v1),$(cat .axl/run/peer-v2),$(cat .axl/run/peer-v3)"

# ---------- 3. preflight ----------
if [[ "$SKIP_PREFLIGHT" != "1" ]]; then
  echo
  echo "=== preflight ==="
  treasury_usdc=$(cast call "$USDC_ADDR" "balanceOf(address)(uint256)" "$TREASURY_ADDR" --rpc-url "$RPC_URL" 2>/dev/null | head -1 | awk '{print $1}')
  treasury_usdc="${treasury_usdc:-0}"
  echo "  treasury USDC: $treasury_usdc"
  if (( treasury_usdc < 200000000 )); then
    echo "    minting 200 mUSDC..."
    cast send "$USDC_ADDR" "mint(address,uint256)" "$TREASURY_ADDR" 200000000 \
      --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
  fi
  echo "    approving aegis to pull 200 mUSDC..."
  cast send "$USDC_ADDR" "approve(address,uint256)" "$AEGIS_ADDR" 200000000 \
    --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
  exec_bal=$(cast balance "$EXECUTOR_ADDR" --rpc-url "$RPC_URL" 2>/dev/null)
  echo "  executor 0G: $exec_bal"
  if [[ "$exec_bal" == "0" ]]; then
    echo "    funding executor with 0.05 0G..."
    cast send "$EXECUTOR_ADDR" --value 0.05ether \
      --rpc-url "$RPC_URL" --private-key "$TREASURY_PK" --legacy >/dev/null 2>&1 || true
  fi
fi

# ---------- 4. spawn verifiers ----------
echo
echo "=== starting 3 verifiers ==="
declare -a VPIDS=()
for i in 1 2 3; do
  vpk_var="V${i}_PK"; vpk="$(eval echo \$$vpk_var)"
  vkey_var="ORG${i}_API_KEY"; vkey="$(eval echo \$$vkey_var)"
  axl_url=""
  case "$i" in
    1) axl_url="$AXL_V1_URL" ;;
    2) axl_url="$AXL_V2_URL" ;;
    3) axl_url="$AXL_V3_URL" ;;
  esac
  KEEPERHUB_API_KEY="$vkey" \
  KEEPERHUB_VERIFIER_INDEX="$i" \
  AEGIS_RPC="$RPC_URL" \
  AEGIS_CONTRACT="$AEGIS_ADDR" \
  REGISTRY_CONTRACT="$REGISTRY_ADDR" \
  VERIFIER_PRIVATE_KEY="$vpk" \
  SWAP_RPC="$SWAP_RPC" \
  OG_PRIVATE_KEY="$vpk" \
  AXL_NODE_URL="$axl_url" \
  AXL_REPLICATE_PEERS="$AXL_VERIFIER_PEERS" \
  go run ./cmd/verifier --label "v$i" --auto-settle >"$LOGDIR/verifier-$i.log" 2>&1 &
  pid=$!
  VPIDS+=("$pid")
  echo "$pid" > ".demo/verifier-$i.pid"
  echo "  v$i pid=$pid log=$LOGDIR/verifier-$i.log"
done

# ---------- 5. spawn cmd/ui ----------
echo
echo "=== starting UI ==="
# Use the prebuilt binary if available; fall back to `go run` for
# dev iteration (slower startup but no rebuild step).
UI_BIN="$ROOT/bin/ui"
UI_CMD=()
if [[ -x "$UI_BIN" ]]; then
  UI_CMD=("$UI_BIN")
else
  echo "  bin/ui not built; using go run (slower). Run scripts/build-ui.sh for a fast binary."
  UI_CMD=(go run ./cmd/ui)
fi

UI_TREASURY_PK="$TREASURY_PK" \
UI_EXECUTOR_PK="$EXECUTOR_PK" \
USDC_CONTRACT="$USDC_ADDR" \
AEGIS_RPC="$RPC_URL" \
AEGIS_CONTRACT="$AEGIS_ADDR" \
REGISTRY_CONTRACT="$REGISTRY_ADDR" \
UI_LOGS_DIR="$ROOT/logs" \
"${UI_CMD[@]}" --addr "$UI_ADDR" --backfill 200000 >"$LOGDIR/ui.log" 2>&1 &
UI_PID=$!
echo "$UI_PID" > .demo/ui.pid
echo "  ui pid=$UI_PID log=$LOGDIR/ui.log"

echo "$LOGDIR" > .demo/logdir

# ---------- 6. wait for the UI to see all verifier peers ----------
echo
echo "=== waiting for verifier AXL peers to register on the UI ==="
UI_HOST="127.0.0.1"
UI_PORT="${UI_ADDR#:}"
[[ "$UI_PORT" == "$UI_ADDR" ]] && UI_PORT="${UI_ADDR##*:}"
for _ in $(seq 1 60); do
  state=$(curl -sf "http://${UI_HOST}:${UI_PORT}/api/state" 2>/dev/null || echo '{}')
  online=$(echo "$state" | python3 -c "
import json,sys
try:
  d = json.load(sys.stdin)
  print(sum(1 for n in d.get('topology', {}).get('nodes', []) if n.get('online') and n.get('role') != 'pub'))
except Exception:
  print(0)
" 2>/dev/null || echo 0)
  if [[ "$online" == "3" ]]; then
    echo "  all 3 verifier daemons online"
    break
  fi
  sleep 1
done

echo
echo "============================================================"
echo "Demo is up. Open http://localhost${UI_ADDR}"
echo "Post a job:    bash scripts/demo-post-job.sh"
echo "Tear down:     bash scripts/demo-down.sh"
echo "============================================================"
