#!/usr/bin/env bash
# Publishes all 9 KeeperHub workflows to the marketplace catalog so
# `call_workflow(slug, inputs)` can invoke them. Required for the
# call_workflow integration path used by LiveClient — KH only exposes
# call_workflow against listed (slugged) workflows.
#
# After this runs, configs/keeperhub.toml gains 9 *_workflow_slug entries
# alongside the existing *_workflow_id entries.
#
# Required env (same orgs as setup-keeperhub-workflows.sh):
#   ORG1_API_KEY, ORG2_API_KEY, ORG3_API_KEY
#
# Optional env:
#   KH_URL          default https://app.keeperhub.com/mcp
#   CHAIN_ID        default 16602 (0G Galileo testnet)
#   CONFIG_PATH     default configs/keeperhub.toml
#   SLUG_PREFIX     default quorum-aegis (slugs become <prefix>-<purpose>-v<idx>)

set -uo pipefail

: "${ORG1_API_KEY:?set ORG1_API_KEY}"
: "${ORG2_API_KEY:?set ORG2_API_KEY}"
: "${ORG3_API_KEY:?set ORG3_API_KEY}"

KH_URL="${KH_URL:-https://app.keeperhub.com/mcp}"
CHAIN_ID="${CHAIN_ID:-16602}"
CONFIG_PATH="${CONFIG_PATH:-configs/keeperhub.toml}"
SLUG_PREFIX="${SLUG_PREFIX:-quorum-aegis}"

# Input schemas declared at listing time so call_workflow validates inputs.
SCHEMA_COMMIT='{"type":"object","properties":{"jobId":{"type":"string"},"commitHash":{"type":"string"}},"required":["jobId","commitHash"]}'
SCHEMA_REVEAL='{"type":"object","properties":{"jobId":{"type":"string"},"verdict":{"type":"string"},"nonce":{"type":"string"}},"required":["jobId","verdict","nonce"]}'
SCHEMA_SETTLE='{"type":"object","properties":{"jobId":{"type":"string"}},"required":["jobId"]}'

new_session() {
  local key="$1"
  curl -sS -D - -o /dev/null \
    -H "Authorization: Bearer $key" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json, text/event-stream" \
    -X POST "$KH_URL" \
    -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"quorum","version":"0.1"}}}' \
    | grep -i '^mcp-session-id:' | awk '{print $2}' | tr -d '\r\n'
}

toml_field() {
  local section="$1" field="$2"
  sed -n "/^\[$section\]/,/^\[/{/^$field[[:space:]]*=/p;}" "$CONFIG_PATH" \
    | head -1 \
    | sed -E 's/^[^=]*=[[:space:]]*"([^"]+)".*/\1/'
}

# list_one API_KEY SID WORKFLOW_ID SLUG SCHEMA_JSON
# echoes the listedSlug from the response (or the slug we asked for if KH
# accepts the request silently).
list_one() {
  local key="$1" sid="$2" wfid="$3" slug="$4" schema="$5"

  local body
  body=$(jq -nc \
    --arg id "$wfid" \
    --arg slug "$slug" \
    --arg chain "$CHAIN_ID" \
    --argjson schema "$schema" \
    '{
      jsonrpc:"2.0", id:30, method:"tools/call",
      params:{
        name:"list_workflow",
        arguments:{
          workflowId:$id,
          slug:$slug,
          chain:$chain,
          category:"verification",
          workflowType:"write",
          inputSchema:$schema
        }
      }
    }')

  local raw
  raw=$(curl -sS \
    -H "Authorization: Bearer $key" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json, text/event-stream" \
    -H "Mcp-Session-Id: $sid" \
    -X POST "$KH_URL" \
    -d "$body")

  # KH wraps the response as {result:{content:[{text:"<json>"}]}}; the
  # inner json carries listedSlug. We tolerate either shape.
  local out
  out=$(echo "$raw" | jq -r '.result.content[0].text // empty' 2>/dev/null)
  if [[ -n "$out" ]]; then
    local got_slug
    got_slug=$(echo "$out" | jq -r '.listedSlug // .slug // empty' 2>/dev/null)
    if [[ -n "$got_slug" ]]; then
      echo "$got_slug"
      return 0
    fi
  fi
  # Error case — surface the message and return our requested slug as a
  # last resort (KH may have accepted the slug even if shape differs).
  local err
  err=$(echo "$raw" | jq -r '.error.message // .result.content[0].text // empty' 2>/dev/null)
  if [[ -n "$err" ]]; then
    echo "list_workflow error: $err" >&2
  fi
  echo "$slug"
}

V1_COMMIT_SLUG="" V1_REVEAL_SLUG="" V1_SETTLE_SLUG=""
V2_COMMIT_SLUG="" V2_REVEAL_SLUG="" V2_SETTLE_SLUG=""
V3_COMMIT_SLUG="" V3_REVEAL_SLUG="" V3_SETTLE_SLUG=""

list_org() {
  local idx="$1" key="$2"
  echo "=== org-$idx ==="
  local sid; sid=$(new_session "$key")
  echo "session=$sid"

  local commit_id reveal_id settle_id
  commit_id=$(toml_field "verifier_$idx" "commit_workflow_id")
  reveal_id=$(toml_field "verifier_$idx" "reveal_workflow_id")
  settle_id=$(toml_field "verifier_$idx" "settle_workflow_id")
  echo "  ids: commit=$commit_id reveal=$reveal_id settle=$settle_id"

  local commit_slug reveal_slug settle_slug
  echo -n "  list commit ... "
  commit_slug=$(list_one "$key" "$sid" "$commit_id" "${SLUG_PREFIX}-commit-v${idx}" "$SCHEMA_COMMIT")
  echo "$commit_slug"

  echo -n "  list reveal ... "
  reveal_slug=$(list_one "$key" "$sid" "$reveal_id" "${SLUG_PREFIX}-reveal-v${idx}" "$SCHEMA_REVEAL")
  echo "$reveal_slug"

  echo -n "  list settle ... "
  settle_slug=$(list_one "$key" "$sid" "$settle_id" "${SLUG_PREFIX}-settle-v${idx}" "$SCHEMA_SETTLE")
  echo "$settle_slug"

  case "$idx" in
    1) V1_COMMIT_SLUG="$commit_slug"; V1_REVEAL_SLUG="$reveal_slug"; V1_SETTLE_SLUG="$settle_slug" ;;
    2) V2_COMMIT_SLUG="$commit_slug"; V2_REVEAL_SLUG="$reveal_slug"; V2_SETTLE_SLUG="$settle_slug" ;;
    3) V3_COMMIT_SLUG="$commit_slug"; V3_REVEAL_SLUG="$reveal_slug"; V3_SETTLE_SLUG="$settle_slug" ;;
  esac
  echo
}

list_org 1 "$ORG1_API_KEY"
list_org 2 "$ORG2_API_KEY"
list_org 3 "$ORG3_API_KEY"

echo "=================================================================="
echo "All 9 listings done. Add these to configs/keeperhub.toml under each"
echo "[verifier_N] section (alongside existing *_workflow_id entries):"
echo "=================================================================="
cat <<TOML

[verifier_1]
commit_workflow_slug = "${V1_COMMIT_SLUG}"
reveal_workflow_slug = "${V1_REVEAL_SLUG}"
settle_workflow_slug = "${V1_SETTLE_SLUG}"

[verifier_2]
commit_workflow_slug = "${V2_COMMIT_SLUG}"
reveal_workflow_slug = "${V2_REVEAL_SLUG}"
settle_workflow_slug = "${V2_SETTLE_SLUG}"

[verifier_3]
commit_workflow_slug = "${V3_COMMIT_SLUG}"
reveal_workflow_slug = "${V3_REVEAL_SLUG}"
settle_workflow_slug = "${V3_SETTLE_SLUG}"
TOML
