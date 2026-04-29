#!/usr/bin/env bash
# End-to-end Anvil demo: register 3 verifiers -> postJob -> submitClaim ->
# commit -> reveal -> settle. Asserts final job status is Settled.
#
# Prereqs: anvil running, `make deploy-local` already ran, jq + cast on PATH.

set -euo pipefail

cd "$(dirname "$0")/.."

DEPLOY=contracts/deployments/local.json
[[ -f $DEPLOY ]] || { echo "missing $DEPLOY — run 'make deploy-local' first" >&2; exit 1; }

RPC=${RPC_URL:-http://127.0.0.1:8545}
USDC=$(jq -r .usdc $DEPLOY)
AEGIS=$(jq -r .quorum $DEPLOY)
REGISTRY=$(jq -r .registry $DEPLOY)

# Anvil's deterministic accounts
DEPLOYER_PK=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
CLIENT=0x70997970C51812dc3A010C7d01b50e0d17dc79C8
CLIENT_PK=0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
EXECUTOR=0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
EXECUTOR_PK=0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
V1=0x90F79bf6EB2c4f870365E785982E1f101E93b906
V1_PK=0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6
V2=0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65
V2_PK=0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a
V3=0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc
V3_PK=0x8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba

STAKE=100000000  # 100 USDC (6 decimals)

step()  { printf '\n==> %s\n' "$*"; }
quiet() { "$@" > /dev/null; }

step "fund verifier 2 and 3 with USDC"
quiet cast send $USDC "mint(address,uint256)" $V2 $STAKE --rpc-url $RPC --private-key $DEPLOYER_PK
quiet cast send $USDC "mint(address,uint256)" $V3 $STAKE --rpc-url $RPC --private-key $DEPLOYER_PK

step "register 3 verifiers (each stakes 100 USDC)"
for entry in "$V1 $V1_PK" "$V2 $V2_PK" "$V3 $V3_PK"; do
  set -- $entry; ADDR=$1; PK=$2
  quiet cast send $USDC "approve(address,uint256)" $REGISTRY $STAKE --rpc-url $RPC --private-key $PK
  quiet cast send $REGISTRY "register(uint256,uint256)" $STAKE 0 --rpc-url $RPC --private-key $PK
  echo "    $ADDR registered"
done

step "publisher: postJob"
AEGIS_RPC=$RPC AEGIS_CONTRACT=$AEGIS USDC_CONTRACT=$USDC \
  CLIENT_PRIVATE_KEY=$CLIENT_PK EXECUTOR_ADDRESS=$EXECUTOR \
  go run ./cmd/publisher >/dev/null

JOB_ID=$(( $(cast call $AEGIS "nextJobId()(uint256)" --rpc-url $RPC) - 1 ))
echo "    jobId=$JOB_ID"

step "executor: submitClaim (mode=honest)"
AEGIS_RPC=$RPC AEGIS_CONTRACT=$AEGIS \
  EXECUTOR_PRIVATE_KEY=$EXECUTOR_PK JOB_ID=$JOB_ID \
  CLIENT_ADDRESS=$CLIENT \
  go run ./cmd/executor --mode=honest >/dev/null

step "verifiers commit (all PASS)"
N1=0x0000000000000000000000000000000000000000000000000000000000000001
N2=0x0000000000000000000000000000000000000000000000000000000000000002
N3=0x0000000000000000000000000000000000000000000000000000000000000003
for entry in "$V1 $V1_PK $N1" "$V2 $V2_PK $N2" "$V3 $V3_PK $N3"; do
  set -- $entry; ADDR=$1; PK=$2; NONCE=$3
  HASH=$(cast keccak "$(cast abi-encode 'f(bool,bytes32,address)' true $NONCE $ADDR)")
  quiet cast send $AEGIS "commitVote(uint256,bytes32)" $JOB_ID $HASH --rpc-url $RPC --private-key $PK
done

step "jump past commit deadline (+6 min)"
quiet cast rpc evm_increaseTime 360 --rpc-url $RPC
quiet cast rpc evm_mine --rpc-url $RPC

step "verifiers reveal"
for entry in "$V1_PK $N1" "$V2_PK $N2" "$V3_PK $N3"; do
  set -- $entry; PK=$1; NONCE=$2
  quiet cast send $AEGIS "revealVote(uint256,bool,bytes32)" $JOB_ID true $NONCE --rpc-url $RPC --private-key $PK
done

step "jump past reveal deadline (+6 min)"
quiet cast rpc evm_increaseTime 360 --rpc-url $RPC
quiet cast rpc evm_mine --rpc-url $RPC

step "settle"
quiet cast send $AEGIS "settle(uint256)" $JOB_ID --rpc-url $RPC --private-key $DEPLOYER_PK

step "verify final state"
STATUS=$(cast call $AEGIS \
  "jobs(uint256)(address,address,uint256,uint256,uint256,bytes32,bytes32,bytes32,uint64,uint64,uint64,uint8)" \
  $JOB_ID --rpc-url $RPC | tail -1)
[[ $STATUS == "3" ]] || { echo "FAIL: status=$STATUS expected 3 (Settled)"; exit 1; }
echo "    status=Settled (3)"
echo
echo "==> demo complete"
