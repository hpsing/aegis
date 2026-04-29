.PHONY: build vet test cover staticcheck fmt-check fresh \
        contracts-build contracts-test contracts-clean contracts-fresh abigen \
        deploy-local deploy-og run-publisher run-executor local-demo \
        setup-axl run-axl stop-axl

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

fresh:
	go clean -cache
	go build ./...

cover:
	go test -cover ./internal/...

# staticcheck must be installed: go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck:
	staticcheck ./...

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "$$out"; exit 1; fi

setup-axl:
	sh scripts/setup-axl.sh

run-axl:
	sh scripts/run-axl.sh

stop-axl:
	sh scripts/stop-axl.sh


contracts-build:
	cd contracts && forge build

contracts-test:
	cd contracts && forge test

contracts-clean:
	cd contracts && forge clean

contracts-fresh:
	cd contracts && forge clean && forge build

# Deploy to a local Anvil node (run `anvil` in another shell first).
# Uses Anvil's first default account.
ANVIL_RPC ?= http://127.0.0.1:8545
ANVIL_PK  ?= 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
deploy-local:
	cd contracts && forge script script/DeployLocal.s.sol:DeployLocal \
		--broadcast --rpc-url $(ANVIL_RPC) --private-key $(ANVIL_PK)

# Deploy to 0G Galileo testnet. Requires:
#   OG_GALILEO_RPC, DEPLOYER_PRIVATE_KEY, TREASURY_ADDRESS
#   USDC_ADDRESS (optional — if unset, deploys MockUSDC)
# 0G's reported base fee is ~0 but the chain enforces a min priority fee of
# 2 gwei, which trips up forge's EIP-1559 math. Use legacy txs with a flat
# gas price to sidestep. Override via OG_GAS_GWEI.
OG_GAS_GWEI ?= 5
deploy-og:
	cd contracts && forge script script/DeployToOG.s.sol:DeployToOG \
		--broadcast --legacy --rpc-url $$OG_GALILEO_RPC \
		--private-key $$DEPLOYER_PRIVATE_KEY \
		--gas-price $(shell echo $$(( $(OG_GAS_GWEI) * 1000000000 )))

# Regenerate abigen bindings from contracts/out/. Run after changing
# contract ABIs.
abigen:
	cd contracts && forge build
	mkdir -p internal/aegis
	jq '.abi' contracts/out/AegisContract.sol/AegisContract.json > /tmp/aegis.abi.json
	jq '.abi' contracts/out/VerifierRegistry.sol/VerifierRegistry.json > /tmp/registry.abi.json
	jq '.abi' contracts/out/MockUSDC.sol/MockUSDC.json > /tmp/mockusdc.abi.json
	abigen --abi /tmp/aegis.abi.json --pkg aegis --type AegisContract --out internal/aegis/aegis_contract.go
	abigen --abi /tmp/registry.abi.json --pkg aegis --type VerifierRegistry --out internal/aegis/verifier_registry.go
	abigen --abi /tmp/mockusdc.abi.json --pkg aegis --type MockUSDC --out internal/aegis/mock_usdc.go
	go build ./...

# Run the publisher (client) CLI. Requires env:
#   AEGIS_RPC, AEGIS_CONTRACT, USDC_CONTRACT, CLIENT_PRIVATE_KEY, EXECUTOR_ADDRESS
# Optional: override amounts with REIMBURSEMENT/FEE/BOUNTY (USDC, 6 decimals).
REIMBURSEMENT ?= 100000000
FEE           ?= 10000000
BOUNTY        ?= 30000000
run-publisher:
	go run ./cmd/publisher \
		--reimbursement=$(REIMBURSEMENT) --fee=$(FEE) --bounty=$(BOUNTY)

# Run the executor CLI. Requires env:
#   AEGIS_RPC, AEGIS_CONTRACT, EXECUTOR_PRIVATE_KEY, JOB_ID, CLIENT_ADDRESS
# Override mode with: make run-executor MODE=high-slippage
MODE ?= honest
run-executor:
	go run ./cmd/executor --mode=$(MODE)

# Full e2e: register 3 verifiers, post a job, submitClaim, commit/reveal, settle.
# Anvil must be running and `make deploy-local` must have run.
local-demo:
	bash scripts/local-demo.sh