#!/usr/bin/env bash
# Boot 4 AXL daemons — one per Quorum agent — so each has a distinct
# peer id addressable by /send. Mesh topology:
#
#   pub  (api 9002, listen 9001) ←─┐
#   v1   (api 9012)               │
#   v2   (api 9022)  ─── peers ───┤  all dial tls://127.0.0.1:9001
#   v3   (api 9032)               │
#                                 │
#   (publisher is the only listener; everyone else dials it)
#
# All four daemons share the same tcp_port (7001). The AXL dialer uses
# the SENDER's tcp_port to reach the destination's listener (see
# .axl/axl/internal/tcp/dial/dial.go) so per-node tcp_ports break
# cross-peer /send. Yggdrasil IPv6 is derived from each node's key, so
# sharing 7001 is collision-free.
#
# Per-role peer ids land in .axl/run/peer-{pub,v1,v2,v3} for the e2e
# demo orchestrator to read.
#
# Run:
#   bash scripts/run-axl-swarm.sh
#   ./scripts/stop-axl-swarm.sh   # to tear down

set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
RUN_DIR="$ROOT/.axl/run"
KEYS_DIR="$ROOT/.axl/keys"
CFGS_DIR="$ROOT/.axl/cfgs"
mkdir -p "$RUN_DIR" "$CFGS_DIR"

BIN="$ROOT/.axl/bin/node"
if [[ ! -x "$BIN" ]]; then
  echo "[axl-swarm] $BIN missing — run scripts/setup-axl.sh first" >&2
  exit 1
fi

OPENSSL_BIN="${OPENSSL_BIN:-/opt/homebrew/opt/openssl@3/bin/openssl}"
if [[ ! -x "$OPENSSL_BIN" ]]; then
  OPENSSL_BIN="$(command -v openssl || true)"
fi

# Roles + their api ports. Order matters for the listener.
# tcp_port is shared (see header) — the AXL dialer uses the sender's port
# to reach the destination's listener.
ROLES=(pub v1 v2 v3)
API_PORTS=(9002 9012 9022 9032)
SHARED_TCP_PORT=7001

# Generate any missing keys (pub uses existing private-a.pem, v1 uses
# private-b.pem from setup-axl.sh; v2 + v3 are net-new).
key_for() {
  local role="$1"
  case "$role" in
    pub) echo "$KEYS_DIR/private-a.pem" ;;
    v1)  echo "$KEYS_DIR/private-b.pem" ;;
    *)   echo "$KEYS_DIR/private-${role}.pem" ;;
  esac
}

ensure_key() {
  local role="$1"
  local key
  key=$(key_for "$role")
  if [[ -f "$key" ]]; then return 0; fi
  if [[ -z "$OPENSSL_BIN" ]]; then
    echo "[axl-swarm] need openssl to generate $key" >&2; exit 1
  fi
  echo "[axl-swarm] generating $key"
  "$OPENSSL_BIN" genpkey -algorithm Ed25519 -out "$key" >/dev/null
}

write_cfg() {
  local role="$1" listen_port="$2" api_port="$3" tcp_port="$4"
  local cfg="$CFGS_DIR/node-${role}.json"
  local key
  key=$(key_for "$role")
  local key_rel="${key#$ROOT/}"

  if [[ "$role" == "pub" ]]; then
    cat >"$cfg" <<EOF
{
  "PrivateKeyPath": "${key_rel}",
  "Listen": ["tls://127.0.0.1:${listen_port}"],
  "Peers": [],
  "api_port": ${api_port},
  "tcp_port": ${tcp_port}
}
EOF
  else
    cat >"$cfg" <<EOF
{
  "PrivateKeyPath": "${key_rel}",
  "Listen": [],
  "Peers": ["tls://127.0.0.1:9001"],
  "api_port": ${api_port},
  "tcp_port": ${tcp_port}
}
EOF
  fi
  echo "$cfg"
}

start_node() {
  local role="$1" cfg="$2" api_port="$3"
  local pidfile="$RUN_DIR/pid-$role"
  local logfile="$RUN_DIR/log-$role"

  # Always (re)start. Reusing a running daemon is unsafe because a
  # config edit (e.g. tcp_port) won't take effect — the daemon would
  # keep listening on the OLD port, mismatching what cross-peer dials
  # use, and /send fails with 502 connection refused.
  if [[ -f "$pidfile" ]] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    local oldpid
    oldpid=$(cat "$pidfile")
    echo "[axl-swarm] $role: stopping stale pid=$oldpid"
    kill "$oldpid" 2>/dev/null || true
    # Wait briefly for the process to release its API port.
    for _ in $(seq 1 20); do
      kill -0 "$oldpid" 2>/dev/null || break
      sleep 0.1
    done
    rm -f "$pidfile"
  fi
  ( "$BIN" -config "$cfg" >"$logfile" 2>&1 & echo $! >"$pidfile" )
  echo "[axl-swarm] started $role pid=$(cat "$pidfile")"

  for _ in $(seq 1 60); do
    if curl -sf "http://127.0.0.1:$api_port/topology" -o "$RUN_DIR/topology-$role.json" 2>/dev/null; then
      python3 -c "import json,sys; d=json.load(open('$RUN_DIR/topology-$role.json')); print(d['our_public_key'])" >"$RUN_DIR/peer-$role"
      echo "[axl-swarm] $role api=:$api_port peer=$(cat "$RUN_DIR/peer-$role")"
      return 0
    fi
    sleep 0.25
  done
  echo "[axl-swarm] $role never came up — last 20 log lines:" >&2
  tail -n 20 "$logfile" >&2 || true
  return 1
}

# Generate keys + write configs.
for role in "${ROLES[@]}"; do
  ensure_key "$role"
done

for i in "${!ROLES[@]}"; do
  role="${ROLES[i]}"
  api_port="${API_PORTS[i]}"
  # listen port is only used by pub (the listener); pass 9001.
  write_cfg "$role" 9001 "$api_port" "$SHARED_TCP_PORT" >/dev/null
done

# Start pub first so others can dial it.
start_node pub "$CFGS_DIR/node-pub.json" 9002
sleep 1
for i in "${!ROLES[@]}"; do
  role="${ROLES[i]}"
  [[ "$role" == "pub" ]] && continue
  start_node "$role" "$CFGS_DIR/node-${role}.json" "${API_PORTS[i]}"
done

echo
echo "[axl-swarm] mesh ready:"
for role in "${ROLES[@]}"; do
  echo "  $role: api=http://127.0.0.1:$(grep -oE '"api_port":[[:space:]]*[0-9]+' "$CFGS_DIR/node-${role}.json" | grep -oE '[0-9]+')  peer=$(cat "$RUN_DIR/peer-$role")"
done
echo
echo "[axl-swarm] stop with: ./scripts/stop-axl-swarm.sh"
