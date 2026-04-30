#!/usr/bin/env bash
# Creates the 9 KeeperHub workflows (3 orgs x commit/reveal/settle) that route
# Quorum's on-chain writes through KeeperHub MCP.
#
# Each org has exactly one wallet integration (the verifier's wallet). KH uses
# that wallet implicitly when the workflow's web3/write-contract action runs.
#
# Required env:
#   ORG1_API_KEY  KeeperHub API key for verifier-1's org
#   ORG2_API_KEY  KeeperHub API key for verifier-2's org
#   ORG3_API_KEY  KeeperHub API key for verifier-3's org
#
# Optional env:
#   KH_URL        defaults to https://app.keeperhub.com/mcp
#   AEGIS_ADDR    defaults to 0xeF60Ad6aB86101F3f787C33957649b5c7a9Ca858
#   CHAIN_ID      defaults to 16602 (0G Galileo testnet)
#   SKIP_V1_COMMIT  set to 1 to skip org-1 commit (already created earlier)
#
# Run:
#   export ORG1_API_KEY=... ORG2_API_KEY=... ORG3_API_KEY=...
#   export SKIP_V1_COMMIT=1   # we already created v1/commit (id=7c611oldnlfch263ug9he)
#   bash scripts/setup-keeperhub-workflows.sh
#
# Output: prints a TOML block at the end mapping verifier+purpose to workflow
# id, ready to drop into configs/keeperhub.toml.

set -euo pipefail

: "${ORG1_API_KEY:?set ORG1_API_KEY}"
: "${ORG2_API_KEY:?set ORG2_API_KEY}"
: "${ORG3_API_KEY:?set ORG3_API_KEY}"

KH_URL="${KH_URL:-https://app.keeperhub.com/mcp}"
AEGIS_ADDR="${AEGIS_ADDR:-0xeF60Ad6aB86101F3f787C33957649b5c7a9Ca858}"
CHAIN_ID="${CHAIN_ID:-16602}"
SKIP_V1_COMMIT="${SKIP_V1_COMMIT:-0}"

# ABI fragments per method, embedded in the workflow's action node.
ABI_COMMIT='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"bytes32","name":"commitHash","type":"bytes32"}],"name":"commitVote","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_REVEAL='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"bool","name":"verdict","type":"bool"},{"internalType":"bytes32","name":"nonce","type":"bytes32"}],"name":"revealVote","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_SETTLE='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"}],"name":"settle","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

# new_session API_KEY -> Mcp-Session-Id
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

# create_workflow API_KEY SID NAME DESC TRIGGER_DESC ACTION_LABEL ACTION_DESC ABI_JSON FUNC_NAME INPUT_FIELDS_JSON FUNC_ARGS_JSON
create_workflow() {
  local key="$1" sid="$2" name="$3" desc="$4" trigDesc="$5" actLabel="$6" actDesc="$7" abi="$8" fn="$9" inputFields="${10}" funcArgs="${11}"
  # Build the JSON-RPC body via jq to avoid manual escape headaches.
  # inputFields and funcArgs come in as JSON strings (e.g. '[{"name":"jobId",...}]')
  # and are spliced via --argjson so they remain typed (NOT stringified).
  local body
  body=$(jq -nc \
    --arg name "$name" \
    --arg desc "$desc" \
    --arg trigDesc "$trigDesc" \
    --arg actLabel "$actLabel" \
    --arg actDesc "$actDesc" \
    --arg net "$CHAIN_ID" \
    --arg addr "$AEGIS_ADDR" \
    --arg abi "$abi" \
    --arg fn "$fn" \
    --argjson inputFields "$inputFields" \
    --argjson funcArgs "$funcArgs" \
    '{
      jsonrpc:"2.0", id:10, method:"tools/call",
      params:{
        name:"create_workflow",
        arguments:{
          name:$name,
          description:$desc,
          nodes:[
            {id:"manual-trigger",type:"trigger",position:{x:100,y:200},
             data:{label:"Manual Trigger",description:$trigDesc,type:"trigger",
                   config:{triggerType:"Manual", inputFields:$inputFields},
                   status:"idle"}},
            {id:"action",type:"action",position:{x:350,y:200},
             data:{label:$actLabel,description:$actDesc,type:"action",
                   config:{actionType:"web3/write-contract",network:$net,contractAddress:$addr,abi:$abi,abiFunction:$fn,functionArgs:$funcArgs},
                   status:"idle"}}
          ],
          edges:[{id:"e1",type:"default",source:"manual-trigger",target:"action"}]
        }
      }
    }')

  curl -sS \
    -H "Authorization: Bearer $key" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json, text/event-stream" \
    -H "Mcp-Session-Id: $sid" \
    -X POST "$KH_URL" \
    -d "$body" \
    | jq -r '.result.content[0].text' \
    | jq -r '.id'
}

# macOS ships bash 3.2 (no associative arrays), so we use plain globals.
V1_COMMIT="" V1_REVEAL="" V1_SETTLE=""
V2_COMMIT="" V2_REVEAL="" V2_SETTLE=""
V3_COMMIT="" V3_REVEAL="" V3_SETTLE=""

fire_org_workflows() {
  local idx="$1" key="$2"
  echo "=== org-$idx ==="
  local sid
  sid=$(new_session "$key")
  echo "session=$sid"

  local commit_id reveal_id settle_id

  if [[ "$idx" == "1" && "$SKIP_V1_COMMIT" == "1" ]]; then
    echo "  skipping v1/commit (already created earlier)"
    commit_id="${V1_COMMIT_ID:-7c611oldnlfch263ug9he}"
  else
    echo -n "  v${idx}/commit  ... "
    commit_id=$(create_workflow "$key" "$sid" \
      "aegis-commit-vote" \
      "Quorum: verifier commits a vote hash to AegisContract on 0G Galileo. Inputs: jobId (uint256), commitHash (bytes32)." \
      "Inputs: jobId (uint256), commitHash (bytes32)" \
      "Commit Vote" "Calls commitVote on AegisContract" \
      "$ABI_COMMIT" "commitVote" \
      '[{"name":"jobId","type":"string"},{"name":"commitHash","type":"string"}]' \
      '"[\"{{@manual-trigger:Manual Trigger.jobId}}\",\"{{@manual-trigger:Manual Trigger.commitHash}}\"]"')
    echo "$commit_id"
  fi

  echo -n "  v${idx}/reveal  ... "
  reveal_id=$(create_workflow "$key" "$sid" \
    "aegis-reveal-vote" \
    "Quorum: verifier reveals their vote on AegisContract. Inputs: jobId (uint256), verdict (bool), nonce (bytes32)." \
    "Inputs: jobId (uint256), verdict (bool), nonce (bytes32)" \
    "Reveal Vote" "Calls revealVote on AegisContract" \
    "$ABI_REVEAL" "revealVote" \
    '[{"name":"jobId","type":"string"},{"name":"verdict","type":"string"},{"name":"nonce","type":"string"}]' \
    '"[\"{{@manual-trigger:Manual Trigger.jobId}}\",\"{{@manual-trigger:Manual Trigger.verdict}}\",\"{{@manual-trigger:Manual Trigger.nonce}}\"]"')
  echo "$reveal_id"

  echo -n "  v${idx}/settle  ... "
  settle_id=$(create_workflow "$key" "$sid" \
    "aegis-settle" \
    "Quorum: anyone can call settle on AegisContract once reveal deadline passes. Input: jobId (uint256)." \
    "Input: jobId (uint256)" \
    "Settle" "Calls settle on AegisContract" \
    "$ABI_SETTLE" "settle" \
    '[{"name":"jobId","type":"string"}]' \
    '"[\"{{@manual-trigger:Manual Trigger.jobId}}\"]"')
  echo "$settle_id"

  case "$idx" in
    1) V1_COMMIT="$commit_id"; V1_REVEAL="$reveal_id"; V1_SETTLE="$settle_id" ;;
    2) V2_COMMIT="$commit_id"; V2_REVEAL="$reveal_id"; V2_SETTLE="$settle_id" ;;
    3) V3_COMMIT="$commit_id"; V3_REVEAL="$reveal_id"; V3_SETTLE="$settle_id" ;;
  esac
}

fire_org_workflows 1 "$ORG1_API_KEY"
fire_org_workflows 2 "$ORG2_API_KEY"
fire_org_workflows 3 "$ORG3_API_KEY"

echo
echo "=================================================================="
echo "All 9 workflows created. Drop this into configs/keeperhub.toml:"
echo "=================================================================="
cat <<TOML

[verifier_1]
commit_workflow_id = "${V1_COMMIT}"
reveal_workflow_id = "${V1_REVEAL}"
settle_workflow_id = "${V1_SETTLE}"

[verifier_2]
commit_workflow_id = "${V2_COMMIT}"
reveal_workflow_id = "${V2_REVEAL}"
settle_workflow_id = "${V2_SETTLE}"

[verifier_3]
commit_workflow_id = "${V3_COMMIT}"
reveal_workflow_id = "${V3_REVEAL}"
settle_workflow_id = "${V3_SETTLE}"
TOML
