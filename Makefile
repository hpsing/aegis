.PHONY: build vet test cover staticcheck fmt-check fresh \
        contracts-build contracts-test contracts-clean contracts-fresh \
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

contracts-build:
	cd contracts && forge build

contracts-test:
	cd contracts && forge test

contracts-clean:
	cd contracts && forge clean

contracts-fresh:
	cd contracts && forge clean && forge build

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
