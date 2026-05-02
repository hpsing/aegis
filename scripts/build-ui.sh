#!/usr/bin/env bash
# Build the SPA and copy it where cmd/ui's go:embed expects it.
#
# Layout:
#   web/dist/            <- Vite output
#   cmd/ui/web-dist/     <- copied here before `go build` so embed can see it
#                            (cmd/ui can't reach ../../web/dist with go:embed)
#
# Run from repo root:
#   bash scripts/build-ui.sh           # build SPA + copy + go build
#   bash scripts/build-ui.sh --skip-go # just SPA + copy (used during dev)

set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"

SKIP_GO=0
for arg in "$@"; do
  case "$arg" in
    --skip-go) SKIP_GO=1 ;;
    *) echo "[build-ui] unknown arg: $arg" >&2; exit 2 ;;
  esac
done

echo "[build-ui] building SPA"
(cd "$ROOT/web" && npm run build)

echo "[build-ui] copying web/dist → cmd/ui/web-dist"
rm -rf "$ROOT/cmd/ui/web-dist"
mkdir -p "$ROOT/cmd/ui/web-dist"
cp -R "$ROOT/web/dist/." "$ROOT/cmd/ui/web-dist/"

# Keep a placeholder so the dir survives even if dist/ is empty (rare).
if ! ls "$ROOT/cmd/ui/web-dist/" >/dev/null 2>&1; then
  echo "placeholder" >"$ROOT/cmd/ui/web-dist/.gitkeep"
fi

if [[ "$SKIP_GO" == "0" ]]; then
  echo "[build-ui] building cmd/ui"
  mkdir -p "$ROOT/bin"
  (cd "$ROOT" && go build -o bin/ui ./cmd/ui)
  echo "[build-ui] done. run: ./bin/ui --addr :3000"
else
  echo "[build-ui] skipped go build (use go run ./cmd/ui)"
fi
