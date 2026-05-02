#!/usr/bin/env bash
# demo-post-job.sh — kick off a fresh job through the running UI.
# Re-runnable for retakes during the recording.
#
# Pre-condition: scripts/demo-up.sh has been run; UI is on UI_ADDR
# (default :3000) and verifier daemons are online.

set -uo pipefail
cd "$(dirname "$0")/.."

UI_ADDR="${UI_ADDR:-:3000}"
UI_HOST="127.0.0.1"
UI_PORT="${UI_ADDR#:}"
[[ "$UI_PORT" == "$UI_ADDR" ]] && UI_PORT="${UI_ADDR##*:}"

echo "=== posting job via UI ==="
echo "  POST http://${UI_HOST}:${UI_PORT}/api/post-job"
res=$(curl -sf -X POST "http://${UI_HOST}:${UI_PORT}/api/post-job" 2>&1) || {
  echo "  ERROR: $res" >&2
  exit 1
}
echo "$res" | python3 -m json.tool 2>/dev/null || echo "$res"
echo
echo "Watch the lifecycle: http://localhost${UI_ADDR}"
