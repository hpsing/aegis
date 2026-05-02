#!/usr/bin/env bash
# logs-ec2.sh — stream Aegis logs from the EC2 box to your terminal.
#
# Required env (same as scripts/deploy-ec2.sh):
#   EC2_HOST       — e.g. 13.232.33.65
#   EC2_SSH_KEY    — path to .pem
#
# Optional env:
#   EC2_USER       — defaults to "ubuntu"
#
# Usage:
#   scripts/logs-ec2.sh                # tail -f all aegis-*.log files (default)
#   scripts/logs-ec2.sh ui             # only ui.log
#   scripts/logs-ec2.sh v1             # only verifier-1.log
#   scripts/logs-ec2.sh v2             # only verifier-2.log
#   scripts/logs-ec2.sh v3             # only verifier-3.log
#   scripts/logs-ec2.sh axl            # all 4 axl-*.log files
#   scripts/logs-ec2.sh systemd        # journalctl -fu 'aegis-*' (systemd's view)
#   scripts/logs-ec2.sh status         # one-shot: systemctl status of every aegis unit
#
# Ctrl-C to stop.

set -uo pipefail

: "${EC2_HOST:?set EC2_HOST}"
: "${EC2_SSH_KEY:?set EC2_SSH_KEY (path to .pem)}"
EC2_USER="${EC2_USER:-ubuntu}"

WHAT="${1:-all}"

case "$WHAT" in
  all)
    # Everything: UI + 3 verifiers + 4 AXL daemons. tail -F prefixes
    # each chunk with "==> /path/to/file <==" so you can tell which
    # process emitted it. --retry handles log files that don't exist
    # yet (e.g. fresh box where a service hasn't run).
    REMOTE_CMD='sudo tail -F --retry \
      /var/log/aegis/ui.log \
      /var/log/aegis/verifier-1.log \
      /var/log/aegis/verifier-2.log \
      /var/log/aegis/verifier-3.log \
      /var/log/aegis/axl-pub.log \
      /var/log/aegis/axl-v1.log \
      /var/log/aegis/axl-v2.log \
      /var/log/aegis/axl-v3.log'
    ;;
  ui)
    REMOTE_CMD='sudo tail -F /var/log/aegis/ui.log'
    ;;
  v1|v2|v3)
    n="${WHAT#v}"
    REMOTE_CMD="sudo tail -F /var/log/aegis/verifier-${n}.log"
    ;;
  axl)
    REMOTE_CMD='sudo tail -F /var/log/aegis/axl-pub.log /var/log/aegis/axl-v1.log /var/log/aegis/axl-v2.log /var/log/aegis/axl-v3.log'
    ;;
  systemd)
    REMOTE_CMD="sudo journalctl -fu 'aegis-*'"
    ;;
  status)
    REMOTE_CMD="sudo systemctl --no-pager status aegis-axl@pub aegis-axl@v{1,2,3} aegis-axl-peers aegis-verifier@{1,2,3} aegis-ui 2>&1 | grep -E '^(●|     Active)' || true"
    ;;
  -h|--help|help)
    sed -n '2,/^set/p' "$0" | head -25
    exit 0
    ;;
  *)
    echo "unknown target: $WHAT" >&2
    echo "valid: all | ui | v1 | v2 | v3 | axl | systemd | status" >&2
    exit 2
    ;;
esac

exec ssh -i "$EC2_SSH_KEY" -o StrictHostKeyChecking=accept-new \
  -t "$EC2_USER@$EC2_HOST" "$REMOTE_CMD"
