#!/usr/bin/env bash
# aegis-discover-peers.sh — installed at /opt/aegis/bin/.
#
# Runs once at boot (oneshot systemd unit) after the AXL daemons start.
# Polls each daemon's /topology API to read the peer id (deterministic
# from the ed25519 key on disk, so safe to write once and forget) and
# emits per-verifier env files so the verifier units can launch with
# the right AXL_NODE_URL + AXL_REPLICATE_PEERS.
#
# Sourced env (from /etc/aegis/aegis.env via systemd EnvironmentFile=):
#   RPC_URL SWAP_RPC AEGIS_ADDR REGISTRY_ADDR
#   V1_PK V2_PK V3_PK
#   ORG1_API_KEY ORG2_API_KEY ORG3_API_KEY

set -euo pipefail

ROLES=(pub v1 v2 v3)
PORTS=(9002 9012 9022 9032)
RUN_DIR=/run/aegis
mkdir -p "$RUN_DIR"

read_peer() {
  local role="$1" port="$2"
  local out="$RUN_DIR/peer-$role"
  for _ in $(seq 1 120); do
    if curl -sf "http://127.0.0.1:$port/topology" -o "$RUN_DIR/topo-$role.json" 2>/dev/null; then
      python3 -c "import json,sys; print(json.load(open('$RUN_DIR/topo-$role.json'))['our_public_key'])" \
        > "$out" 2>/dev/null && return 0
    fi
    sleep 0.5
  done
  echo "ERROR: AXL daemon $role (:$port) never responded with topology" >&2
  return 1
}

for i in "${!ROLES[@]}"; do
  read_peer "${ROLES[i]}" "${PORTS[i]}"
done

PEERS="$(cat "$RUN_DIR/peer-v1"),$(cat "$RUN_DIR/peer-v2"),$(cat "$RUN_DIR/peer-v3")"
echo "discovered AXL verifier peers: $PEERS"

write_verifier_env() {
  local i="$1" port="$2"
  local vpk_var="V${i}_PK"; local vpk="${!vpk_var}"
  local vkey_var="ORG${i}_API_KEY"; local vkey="${!vkey_var}"
  local out="/etc/aegis/verifier-$i.env"
  cat > "$out" <<EOF
KEEPERHUB_API_KEY=$vkey
KEEPERHUB_VERIFIER_INDEX=$i
AEGIS_RPC=$RPC_URL
AEGIS_CONTRACT=$AEGIS_ADDR
REGISTRY_CONTRACT=$REGISTRY_ADDR
VERIFIER_PRIVATE_KEY=$vpk
SWAP_RPC=$SWAP_RPC
OG_PRIVATE_KEY=$vpk
AXL_NODE_URL=http://127.0.0.1:$port
AXL_REPLICATE_PEERS=$PEERS
EOF
  chmod 640 "$out"
  chown root:aegis "$out"
  echo "wrote $out (verifier $i, axl :$port)"
}

write_verifier_env 1 9012
write_verifier_env 2 9022
write_verifier_env 3 9032
