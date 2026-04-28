.PHONY: build vet test cover staticcheck fmt-check \
        setup-axl run-axl stop-axl

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

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

