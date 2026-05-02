#!/usr/bin/env bash
# install-remote.sh — run ONCE on the EC2 box (as a sudo-capable user).
# Lays down everything systemd needs to manage the aegis stack.
#
# Idempotent — safe to re-run after edits to systemd units / Caddyfile;
# it will overwrite them and `daemon-reload`.
#
# Inputs (env, optional):
#   PUBLIC_HOST   — sslip.io hostname for Caddy. Defaults to the public
#                   IPv4 with dots replaced by hyphens, e.g.
#                   "13-232-33-65.sslip.io".
#
# Run after `scripts/deploy-ec2.sh` has rsync'd the payload to /opt/aegis.

set -euo pipefail

if [[ "$EUID" -ne 0 ]]; then
  exec sudo -E bash "$0" "$@"
fi

ROOT=/opt/aegis
PAYLOAD="$ROOT"     # rsync target

# ---------- 1. user + dirs ----------
if ! id aegis >/dev/null 2>&1; then
  useradd --system --home /opt/aegis --shell /usr/sbin/nologin aegis
fi
install -d -o aegis -g aegis /opt/aegis /opt/aegis/bin /opt/aegis/axl /opt/aegis/axl/cfgs /opt/aegis/axl/keys /var/log/aegis /run/aegis
install -d -m 750 -o root -g aegis /etc/aegis

# ---------- 2. binaries ----------
install -m 755 -o aegis -g aegis "$PAYLOAD/dist/linux-amd64/aegis-ui"        /opt/aegis/bin/aegis-ui
install -m 755 -o aegis -g aegis "$PAYLOAD/dist/linux-amd64/aegis-verifier"  /opt/aegis/bin/aegis-verifier
install -m 755 -o aegis -g aegis "$PAYLOAD/dist/linux-amd64/aegis-executor"  /opt/aegis/bin/aegis-executor
install -m 755 -o aegis -g aegis "$PAYLOAD/dist/linux-amd64/aegis-publisher" /opt/aegis/bin/aegis-publisher
install -m 755 -o aegis -g aegis "$PAYLOAD/dist/linux-amd64/axl-node"        /opt/aegis/bin/axl-node
install -m 755 -o root  -g root  "$PAYLOAD/deploy/aegis-discover-peers.sh"   /opt/aegis/bin/aegis-discover-peers.sh

# Verifier reads configs/keeperhub.toml relative to its working dir.
install -d -o aegis -g aegis /opt/aegis/configs
install -m 644 -o aegis -g aegis "$PAYLOAD/configs-src/keeperhub.toml"       /opt/aegis/configs/keeperhub.toml

# ---------- 3. AXL keys + configs ----------
# ed25519 keys are persistent identity; only generate if missing.
ensure_key() {
  local pem="$1"
  if [[ ! -f "$pem" ]]; then
    openssl genpkey -algorithm Ed25519 -out "$pem" >/dev/null
    chown aegis:aegis "$pem"
    chmod 600 "$pem"
  fi
}
ensure_key /opt/aegis/axl/keys/private-pub.pem
ensure_key /opt/aegis/axl/keys/private-v1.pem
ensure_key /opt/aegis/axl/keys/private-v2.pem
ensure_key /opt/aegis/axl/keys/private-v3.pem

# AXL configs — pub listens, v1/v2/v3 dial it. All four share tcp_port
# 7001 (the AXL dialer routes via the SENDER's tcp_port; per-node ports
# break cross-peer /send — see scripts/run-axl-swarm.sh comment).
write_axl_cfg() {
  local role="$1" api_port="$2" listener="$3"
  local key="/opt/aegis/axl/keys/private-${role}.pem"
  local cfg="/opt/aegis/axl/cfgs/node-${role}.json"
  if [[ "$listener" == "yes" ]]; then
    cat > "$cfg" <<EOF
{
  "PrivateKeyPath": "${key}",
  "Listen": ["tls://127.0.0.1:9001"],
  "Peers": [],
  "api_port": ${api_port},
  "tcp_port": 7001
}
EOF
  else
    cat > "$cfg" <<EOF
{
  "PrivateKeyPath": "${key}",
  "Listen": [],
  "Peers": ["tls://127.0.0.1:9001"],
  "api_port": ${api_port},
  "tcp_port": 7001
}
EOF
  fi
  chown aegis:aegis "$cfg"
}
write_axl_cfg pub 9002 yes
write_axl_cfg v1  9012 no
write_axl_cfg v2  9022 no
write_axl_cfg v3  9032 no

# ---------- 4. env file ----------
if [[ ! -f /etc/aegis/aegis.env ]]; then
  install -m 640 -o root -g aegis "$PAYLOAD/deploy/aegis.env.example" /etc/aegis/aegis.env
  echo
  echo "*** /etc/aegis/aegis.env created from template — fill in the secret values: ***"
  echo "    sudo nano /etc/aegis/aegis.env"
  echo "    sudo systemctl restart aegis.target"
  echo
fi

# ---------- 5. systemd units ----------
for unit in aegis.target aegis-axl@.service aegis-axl-peers.service \
            aegis-verifier@.service aegis-ui.service; do
  install -m 644 -o root -g root "$PAYLOAD/deploy/systemd/$unit" "/etc/systemd/system/$unit"
done
systemctl daemon-reload

# ---------- 6. Caddy ----------
if ! command -v caddy >/dev/null 2>&1; then
  echo "Caddy not installed — installing from cloudsmith…"
  apt-get install -y curl ca-certificates debian-keyring debian-archive-keyring apt-transport-https
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' > /etc/apt/sources.list.d/caddy-stable.list
  apt-get update
  apt-get install -y caddy
fi

# Substitute the public hostname into the Caddyfile if PUBLIC_HOST is set.
if [[ -n "${PUBLIC_HOST:-}" ]]; then
  sed "s/13-232-33-65\.sslip\.io/${PUBLIC_HOST}/g" "$PAYLOAD/deploy/Caddyfile" > /etc/caddy/Caddyfile
else
  install -m 644 "$PAYLOAD/deploy/Caddyfile" /etc/caddy/Caddyfile
fi
systemctl reload caddy || systemctl restart caddy

# ---------- 7. enable + start ----------
systemctl enable aegis.target \
  aegis-axl@pub.service aegis-axl@v1.service aegis-axl@v2.service aegis-axl@v3.service \
  aegis-axl-peers.service \
  aegis-verifier@1.service aegis-verifier@2.service aegis-verifier@3.service \
  aegis-ui.service

# Only start if the env file has real values (TREASURY_PK still placeholder = block).
if grep -q '^TREASURY_PK=0\.\.\.\|^TREASURY_PK=$\|^TREASURY_PK=0x\.\.\.' /etc/aegis/aegis.env; then
  echo
  echo "WARNING: /etc/aegis/aegis.env still has placeholder values — NOT starting aegis.target."
  echo "         Edit the file, then run: sudo systemctl start aegis.target"
else
  systemctl restart aegis.target
  echo
  echo "aegis.target restarted. Tail logs with:"
  echo "  sudo journalctl -u 'aegis-*' -f"
fi

echo
echo "install-remote.sh done."
