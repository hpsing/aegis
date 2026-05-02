#!/usr/bin/env bash
# build-linux.sh — cross-compile every aegis binary (and the AXL node)
# to dist/linux-amd64/ so we can scp them to the EC2 box without
# building on the box (t2.micro has 1 GB RAM and the Go build OOMs).
#
# Outputs:
#   dist/linux-amd64/aegis-ui
#   dist/linux-amd64/aegis-verifier
#   dist/linux-amd64/aegis-executor
#   dist/linux-amd64/aegis-publisher
#   dist/linux-amd64/axl-node                  (built from .axl/axl)
#   dist/linux-amd64/web-dist/                 (SPA assets bundled with the UI binary)
#
# Run:
#   bash scripts/build-linux.sh

set -euo pipefail
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
OUT="$ROOT/dist/linux-amd64"
mkdir -p "$OUT"

echo "=== building SPA into cmd/ui/web-dist (so go:embed picks it up) ==="
bash scripts/build-ui.sh

# CGO_ENABLED=0 → static binary, no glibc version mismatch on EC2.
# GOTOOLCHAIN=local stops `go build` from auto-downloading a different
# toolchain on the build host (CI runners benefit; local dev no-op).
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GOTOOLCHAIN=local

LDFLAGS="-s -w"

echo "=== aegis-ui ==="
go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/aegis-ui" ./cmd/ui

echo "=== aegis-verifier ==="
go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/aegis-verifier" ./cmd/verifier

echo "=== aegis-executor ==="
go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/aegis-executor" ./cmd/executor

echo "=== aegis-publisher ==="
go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/aegis-publisher" ./cmd/publisher

echo "=== axl-node ==="
AXL_SRC="$ROOT/.axl/axl"
if [[ ! -d "$AXL_SRC/.git" ]]; then
  echo "  cloning axl source first…"
  bash scripts/setup-axl.sh
fi
# AXL pulls in gvisor, which has runtime_constants_go125.go +
# runtime_constants_go126.go that BOTH compile under Go 1.26 → duplicate
# symbol error. Pin to 1.25.5 (matches scripts/setup-axl.sh). Go will
# download the toolchain on demand if not already present.
pushd "$AXL_SRC" >/dev/null
GOTOOLCHAIN=go1.25.5 GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/axl-node" ./cmd/node
popd >/dev/null

ls -lh "$OUT" | awk 'NR>1 {printf "  %-22s %s\n", $9, $5}'
echo
echo "Done. Binaries in: $OUT"
