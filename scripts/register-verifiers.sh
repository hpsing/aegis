#!/usr/bin/env bash
# Registers the 3 verifier wallets in VerifierRegistry on 0G Galileo with
# 100 mUSDC stake each. Each verifier signs its own register() so
# msg.sender is the verifier address.
#
# 0G Galileo's RPC sometimes returns "server returned a null response"
# even though the tx actually landed. This script treats every cast send
# as best-effort and verifies registration via cast call at the end.
# Re-run is safe: an already-registered verifier is a no-op.
#
# Required env (3 verifier private keys):
#   V1_PK   private key of 0x1Bd14313fe46Bc9ff2a2042898009E0C42aBaA4A
#   V2_PK   private key of 0x6125c17E4893D7d8a04a1f962794f2da80803E31
#   V3_PK   private key of 0x8F3e641E80954C2a7DdA9CA13d1197D049EeE37f
#
# Optional env:
#   RPC_URL          0G Galileo RPC (default https://evmrpc-testnet.0g.ai)
#   USDC_ADDR        MockUSDC (default from configs/keeperhub.toml deployments)
#   REGISTRY_ADDR    VerifierRegistry (default ditto)
#   STAKE            stake in mUSDC base units (default 100e6 = 100 mUSDC)

set -uo pipefail   # NB: not -e — we tolerate cast's null-response error

: "${V1_PK:?set V1_PK to verifier-1 private key}"
: "${V2_PK:?set V2_PK to verifier-2 private key}"
: "${V3_PK:?set V3_PK to verifier-3 private key}"

RPC_URL="${RPC_URL:-https://evmrpc-testnet.0g.ai}"
USDC_ADDR="${USDC_ADDR:-0xC74c0D2e91B715C89474c8480C812170054ef422}"
REGISTRY_ADDR="${REGISTRY_ADDR:-0x5ec805A1991ECa4Fb876E86866bc32D9095F1270}"
STAKE="${STAKE:-100000000}"   # 100 mUSDC (6 decimals)

# best_effort_send runs `cast send` and swallows the null-response error
# 0G's RPC sometimes throws even on successful txs. We verify state
# afterward via cast call, so a swallowed error is fine.
best_effort_send() {
  local label="$1"; shift
  echo -n "  $label ... "
  if cast send "$@" >/dev/null 2>&1; then
    echo "ok"
  else
    echo "(rpc returned null — will verify on-chain state)"
  fi
}

# read_stake returns the verifier's bonded stake (first field of the
# verifiers(address) tuple). 0 means not registered.
read_stake() {
  local addr="$1"
  cast call "$REGISTRY_ADDR" \
    "verifiers(address)(uint256,uint256,uint256,uint256,uint256,bool)" \
    "$addr" --rpc-url "$RPC_URL" 2>/dev/null \
    | head -1 | awk '{print $1}'
}

register_one() {
  local pk="$1" idx="$2"
  local addr; addr=$(cast wallet address --private-key "$pk")
  echo "=== verifier-$idx ($addr) ==="

  local stake_before; stake_before=$(read_stake "$addr")
  if [[ -n "$stake_before" && "$stake_before" != "0" ]]; then
    echo "  already registered (stake=$stake_before), skipping"
    return 0
  fi

  best_effort_send "mint    " \
    "$USDC_ADDR" "mint(address,uint256)" "$addr" "$STAKE" \
    --rpc-url "$RPC_URL" --private-key "$pk" --legacy

  best_effort_send "approve " \
    "$USDC_ADDR" "approve(address,uint256)" "$REGISTRY_ADDR" "$STAKE" \
    --rpc-url "$RPC_URL" --private-key "$pk" --legacy

  best_effort_send "register" \
    "$REGISTRY_ADDR" "register(uint256,uint256)" "$STAKE" "0" \
    --rpc-url "$RPC_URL" --private-key "$pk" --legacy

  local stake_after; stake_after=$(read_stake "$addr")
  if [[ -n "$stake_after" && "$stake_after" != "0" ]]; then
    echo "  ✓ registered (stake=$stake_after)"
  else
    echo "  ✗ register did NOT land (stake still 0). Inspect tx hashes manually."
    return 1
  fi
  echo
}

register_one "$V1_PK" 1 || true
register_one "$V2_PK" 2 || true
register_one "$V3_PK" 3 || true

echo "=================================================================="
echo "Final state:"
echo "=================================================================="
for tup in "1:$V1_PK" "2:$V2_PK" "3:$V3_PK"; do
  idx="${tup%%:*}"; pk="${tup#*:}"
  addr=$(cast wallet address --private-key "$pk")
  stake=$(read_stake "$addr")
  echo "verifier-$idx  $addr  stake=${stake:-?}"
done
