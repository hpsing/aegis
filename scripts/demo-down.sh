#!/usr/bin/env bash
# demo-down.sh — stop everything started by demo-up.sh: verifiers, UI,
# AXL swarm.

set -uo pipefail
cd "$(dirname "$0")/.."

echo "=== stopping verifiers ==="
for i in 1 2 3; do
  pidfile=".demo/verifier-$i.pid"
  if [[ -f "$pidfile" ]]; then
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      echo "  v$i pid=$pid"
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
done

echo "=== stopping UI ==="
if [[ -f .demo/ui.pid ]]; then
  pid=$(cat .demo/ui.pid)
  if kill -0 "$pid" 2>/dev/null; then
    echo "  ui pid=$pid"
    kill "$pid" 2>/dev/null || true
  fi
  rm -f .demo/ui.pid
fi

# Catch-alls. The go-run wrapper PID we tracked in pidfile is the
# parent; the actual verifier binary lives at
# $GOCACHE/.../<hash>/verifier and survives the wrapper's death
# (orphaned with PPID=1) when go run doesn't propagate SIGTERM. We
# match both /verifier --label (compiled binary) and the wrapper
# explicitly to leave nothing behind.
for pat in "/verifier --label" "go run ./cmd/verifier" "go run ./cmd/ui" "exe/ui" "bin/ui"; do
  pgrep -f "$pat" 2>/dev/null | xargs -r kill -9 2>/dev/null || true
done
sleep 1
remaining=$(pgrep -f '/verifier --label' 2>/dev/null || true)
if [[ -n "$remaining" ]]; then
  echo "  WARN: verifier still running after kill: $remaining"
fi

echo "=== stopping AXL swarm ==="
bash scripts/stop-axl-swarm.sh 2>/dev/null || true

rm -f .demo/logdir
echo "done."
