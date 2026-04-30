#!/usr/bin/env bash
# Tear down the 4-node AXL swarm started by run-axl-swarm.sh.
set -euo pipefail
cd "$(dirname "$0")/.."
RUN_DIR=".axl/run"
for role in pub v1 v2 v3; do
  pidfile="$RUN_DIR/pid-$role"
  if [[ -f "$pidfile" ]]; then
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" && echo "[stop] $role pid=$pid stopped"
    fi
    rm -f "$pidfile"
  fi
done
