#!/usr/bin/env bash
# Boot two AXL daemons (node A on 9002, node B on 9012) in background.
# Polls /topology on each until ready, prints both peer ids.

set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
RUN_DIR="$ROOT/.axl/run"
mkdir -p "$RUN_DIR"

BIN="$ROOT/.axl/bin/node"
if [[ ! -x "$BIN" ]]; then
  echo "[run-axl] $BIN missing — run scripts/setup-axl.sh first" >&2
  exit 1
fi

start_node() {
  local who=$1 cfg=$2 api_port=$3
  local pidfile="$RUN_DIR/pid-$who"
  local logfile="$RUN_DIR/log-$who"

  if [[ -f "$pidfile" ]] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    echo "[run-axl] node $who already running pid=$(cat "$pidfile")"
  else
    echo "[run-axl] starting node $who -> $logfile"
    ( "$BIN" -config "$cfg" >"$logfile" 2>&1 & echo $! >"$pidfile" )
  fi

  echo "[run-axl] waiting for node $who topology on :$api_port"
  for _ in $(seq 1 50); do
    if curl -sf "http://127.0.0.1:$api_port/topology" -o "$RUN_DIR/topology-$who.json"; then
      python3 -c "import json,sys; d=json.load(open('$RUN_DIR/topology-$who.json')); print(d['our_public_key'])" >"$RUN_DIR/peer-$who"
      echo "[run-axl] node $who peer=$(cat "$RUN_DIR/peer-$who")"
      return 0
    fi
    sleep 0.2
  done
  echo "[run-axl] node $who never came up — last 20 log lines:" >&2
  tail -n 20 "$logfile" >&2 || true
  return 1
}

start_node a configs/node-a.json 9002
start_node b configs/node-b.json 9012

cat <<EOF
[run-axl] both nodes ready.
[run-axl]   node A: 127.0.0.1:9002  peer=$(cat "$RUN_DIR/peer-a")
[run-axl]   node B: 127.0.0.1:9012  peer=$(cat "$RUN_DIR/peer-b")
[run-axl] stop with: ./scripts/stop-axl.sh
EOF
