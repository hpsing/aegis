#!/usr/bin/env bash
# Funds the 3 verifier addresses on 0G Galileo from the treasury wallet.
#
# Required env:
#   TREASURY_PK   private key of 0x7C9DcA2fB05cFc732794CEe846888f3B1D7FE06a
#
# Optional env:
#   AMOUNT_OG     amount per verifier, in 0G tokens (default 0.05)
#   RPC_URL       0G Galileo EVM RPC (default https://evmrpc-testnet.0g.ai)
#
# Run:
#   export TREASURY_PK=0x...
#   bash scripts/fund-verifiers.sh

set -euo pipefail

: "${TREASURY_PK:?set TREASURY_PK to the treasury wallet private key}"
RPC_URL="${RPC_URL:-https://evmrpc-testnet.0g.ai}"
AMOUNT_OG="${AMOUNT_OG:-0.05}"

VERIFIERS=(
  "0x1Bd14313fe46Bc9ff2a2042898009E0C42aBaA4A"
  "0x6125c17E4893D7d8a04a1f962794f2da80803E31"
  "0x8F3e641E80954C2a7DdA9CA13d1197D049EeE37f"
)

echo "treasury: $(cast wallet address --private-key "$TREASURY_PK")"
echo "rpc:      $RPC_URL"
echo "amount:   ${AMOUNT_OG} 0G per verifier"
echo

for addr in "${VERIFIERS[@]}"; do
  before=$(cast balance "$addr" --rpc-url "$RPC_URL" --ether)
  echo "→ $addr  before=${before}"
  cast send "$addr" \
    --value "${AMOUNT_OG}ether" \
    --rpc-url "$RPC_URL" \
    --private-key "$TREASURY_PK" \
    --legacy \
    >/dev/null
  after=$(cast balance "$addr" --rpc-url "$RPC_URL" --ether)
  echo "  after=${after}"
done

echo
echo "treasury remaining: $(cast balance "$(cast wallet address --private-key "$TREASURY_PK")" --rpc-url "$RPC_URL" --ether) 0G"
