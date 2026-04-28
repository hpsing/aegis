#!/usr/bin/env bash
# Clone gensyn-ai/axl, build the node binary, and generate two ed25519 keys
# for the AXL daemons (NOT for our verifier/publisher — those use
# secp256k1 from internal/envelope.GenerateKeypair). Idempotent.
#
# Output paths (relative to project root):
#   .axl/axl/                <- cloned source
#   .axl/bin/node            <- built binary
#   .axl/keys/private-a.pem  <- node A key (AXL daemon identity)
#   .axl/keys/private-b.pem  <- node B key (AXL daemon identity)

set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
AXL_DIR="$ROOT/.axl"
SRC_DIR="$AXL_DIR/axl"
BIN="$AXL_DIR/bin/node"
KEYS_DIR="$AXL_DIR/keys"
REPO="https://github.com/gensyn-ai/axl.git"

mkdir -p "$AXL_DIR/bin" "$KEYS_DIR"

if [[ ! -d "$SRC_DIR/.git" ]]; then
  echo "[setup-axl] cloning $REPO"
  git clone --depth=1 "$REPO" "$SRC_DIR"
else
  echo "[setup-axl] axl source already present at $SRC_DIR"
fi

if [[ ! -x "$BIN" ]]; then
  echo "[setup-axl] building axl node binary"
  pushd "$SRC_DIR" >/dev/null
  GOTOOLCHAIN=go1.25.5 go build -o "$BIN" ./cmd/node/
  popd >/dev/null
  echo "[setup-axl] built $BIN"
else
  echo "[setup-axl] axl binary already at $BIN"
fi

OPENSSL_BIN="${OPENSSL_BIN:-}"
if [[ -z "$OPENSSL_BIN" ]]; then
  if [[ -x /opt/homebrew/opt/openssl/bin/openssl ]]; then
    OPENSSL_BIN=/opt/homebrew/opt/openssl/bin/openssl
  elif [[ -x /usr/local/opt/openssl/bin/openssl ]]; then
    OPENSSL_BIN=/usr/local/opt/openssl/bin/openssl
  else
    OPENSSL_BIN=$(command -v openssl)
  fi
fi
echo "[setup-axl] using openssl at $OPENSSL_BIN"

for who in a b; do
  KEY="$KEYS_DIR/private-$who.pem"
  if [[ ! -f "$KEY" ]]; then
    "$OPENSSL_BIN" genpkey -algorithm ed25519 -out "$KEY"
    chmod 600 "$KEY"
    echo "[setup-axl] generated $KEY"
  else
    echo "[setup-axl] key $KEY already exists"
  fi
done

echo "[setup-axl] done."
