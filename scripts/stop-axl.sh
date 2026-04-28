#!/usr/bin/env bash
# Kill both AXL daemons started by scripts/run-axl.sh.

set -euo pipefail

cd "$(dirname "$0")/.."
RUN_DIR="$(pwd)/.axl/run"

for who in a b; do
  pidfile="$RUN_DIR/pid-$who"
  if [[ -f "$pidfile" ]]; then
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      echo "[stop-axl] killing node $who pid=$pid"
      kill "$pid" || true
      for _ in 1 2 3 4 5; do
        kill -0 "$pid" 2>/dev/null || break
        sleep 0.2
      done
      kill -9 "$pid" 2>/dev/null || true
    else
      echo "[stop-axl] node $who pid=$pid not running"
    fi
    rm -f "$pidfile"
  else
    echo "[stop-axl] no pidfile for $who"
  fi
done
