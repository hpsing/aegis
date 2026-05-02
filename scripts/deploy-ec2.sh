#!/usr/bin/env bash
# deploy-ec2.sh — laptop-side: cross-compile, rsync to EC2, restart units.
#
# Required env:
#   EC2_HOST       — e.g. 13.232.33.65 or ec2-13-232-33-65.ap-south-1.compute.amazonaws.com
#   EC2_SSH_KEY    — path to .pem (e.g. ~/Downloads/hackathon_machine.pem)
#
# Optional env:
#   EC2_USER       — defaults to "ubuntu" (Ubuntu AMI). Use "ec2-user" for AL2023.
#   PUBLIC_HOST    — sslip.io hostname for Caddy. Defaults to <ip>.sslip.io with
#                    dots replaced by hyphens.
#   SKIP_BUILD=1   — reuse existing dist/linux-amd64/ (faster iteration).
#
# Run:
#   EC2_HOST=13.232.33.65 EC2_SSH_KEY=~/Downloads/hackathon_machine.pem \
#     bash scripts/deploy-ec2.sh

set -euo pipefail
cd "$(dirname "$0")/.."

: "${EC2_HOST:?set EC2_HOST}"
: "${EC2_SSH_KEY:?set EC2_SSH_KEY (path to .pem)}"
EC2_USER="${EC2_USER:-ubuntu}"
SKIP_BUILD="${SKIP_BUILD:-0}"

# Default sslip.io host derived from the EC2 IP. If EC2_HOST is a DNS
# name we fall back to resolving it.
default_public_host() {
  local h="$EC2_HOST"
  if [[ "$h" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "${h//./-}.sslip.io"
    return
  fi
  local ip
  ip="$(getent ahostsv4 "$h" 2>/dev/null | awk 'NR==1{print $1}')" || true
  if [[ -z "$ip" ]]; then
    ip="$(dig +short "$h" 2>/dev/null | head -1)" || true
  fi
  if [[ -z "$ip" ]]; then
    echo "ERROR: could not resolve $EC2_HOST to an IPv4 address; set PUBLIC_HOST explicitly" >&2
    exit 1
  fi
  echo "${ip//./-}.sslip.io"
}
PUBLIC_HOST="${PUBLIC_HOST:-$(default_public_host)}"

SSH=(ssh -i "$EC2_SSH_KEY" -o StrictHostKeyChecking=accept-new "$EC2_USER@$EC2_HOST")
RSYNC_RSH="ssh -i $EC2_SSH_KEY -o StrictHostKeyChecking=accept-new"

echo "EC2:   $EC2_USER@$EC2_HOST"
echo "Host:  $PUBLIC_HOST  (Caddy will issue a Let's Encrypt cert for this)"

# ---------- 1. cross-compile ----------
if [[ "$SKIP_BUILD" != "1" ]]; then
  bash scripts/build-linux.sh
fi
[[ -f dist/linux-amd64/aegis-ui ]] || { echo "missing dist/linux-amd64/aegis-ui — run scripts/build-linux.sh"; exit 1; }

# ---------- 2. rsync payload ----------
# Bootstrap a writable staging dir on the box first; aegis user owns
# /opt/aegis but the SSH user (ubuntu) doesn't, so we land in /tmp and
# install-remote.sh moves things into place with the right ownership.
echo
echo "=== rsync ==="
"${SSH[@]}" 'rm -rf /tmp/aegis-payload && mkdir -p /tmp/aegis-payload'
rsync -az --delete \
  --rsh="$RSYNC_RSH" \
  --include='dist/' --include='dist/linux-amd64/' --include='dist/linux-amd64/**' \
  --include='deploy/' --include='deploy/**' \
  --include='configs/' --include='configs/**' \
  --exclude='*' \
  ./ "$EC2_USER@$EC2_HOST:/tmp/aegis-payload/"

# ---------- 3. install + restart ----------
echo
echo "=== install + restart on remote ==="
"${SSH[@]}" "sudo install -d /opt/aegis && sudo rsync -a --delete /tmp/aegis-payload/dist/ /opt/aegis/dist/ && sudo rsync -a /tmp/aegis-payload/deploy/ /opt/aegis/deploy/ && sudo rsync -a /tmp/aegis-payload/configs/ /opt/aegis/configs-src/ && sudo PUBLIC_HOST=${PUBLIC_HOST} bash /opt/aegis/deploy/install-remote.sh"

echo
echo "Deployed. Public URL once Caddy gets its cert (~30s on first run):"
echo "  https://$PUBLIC_HOST/api/state"
echo
echo "Paste that base URL into the SPA's API pill on https://hpsing.github.io/aegis/"
