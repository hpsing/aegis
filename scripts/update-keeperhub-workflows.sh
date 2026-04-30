#!/usr/bin/env bash
# Patches all 9 KeeperHub workflows so the trigger declares inputFields
# and the action uses real {{name}} template substitution. Replaces the
# initial AI-suggested-but-broken `[$jobId$, $commitHash$]` string with
# the correct ["{{jobId}}", "{{commitHash}}"] array form.
#
# Required env (same as setup-keeperhub-workflows.sh):
#   ORG1_API_KEY, ORG2_API_KEY, ORG3_API_KEY
#
# Optional env:
#   KH_URL          default https://app.keeperhub.com/mcp
#   AEGIS_ADDR      default 0xeF60Ad6aB86101F3f787C33957649b5c7a9Ca858
#   CHAIN_ID        default 16602
#   CONFIG_PATH     default configs/keeperhub.toml (where workflow ids live)

set -uo pipefail

: "${ORG1_API_KEY:?set ORG1_API_KEY}"
: "${ORG2_API_KEY:?set ORG2_API_KEY}"
: "${ORG3_API_KEY:?set ORG3_API_KEY}"

KH_URL="${KH_URL:-https://app.keeperhub.com/mcp}"
AEGIS_ADDR="${AEGIS_ADDR:-0xeF60Ad6aB86101F3f787C33957649b5c7a9Ca858}"
CHAIN_ID="${CHAIN_ID:-16602}"
CONFIG_PATH="${CONFIG_PATH:-configs/keeperhub.toml}"

ABI_COMMIT='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"bytes32","name":"commitHash","type":"bytes32"}],"name":"commitVote","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_REVEAL='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"bool","name":"verdict","type":"bool"},{"internalType":"bytes32","name":"nonce","type":"bytes32"}],"name":"revealVote","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_SETTLE='[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"}],"name":"settle","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

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

# Extract a quoted string value from configs/keeperhub.toml under [section].
# sed-based: more reliable than awk regex escaping in shell.
toml_field() {
  local section="$1" field="$2"
  sed -n "/^\[$section\]/,/^\[/{/^$field[[:space:]]*=/p;}" "$CONFIG_PATH" \
    | head -1 \
    | sed -E 's/^[^=]*=[[:space:]]*"([^"]+)".*/\1/'
}

# update_workflow API_KEY SID WORKFLOW_ID NAME DESC INPUT_FIELDS_JSON ACT_LABEL ABI FUNC FUNC_ARGS_JSON
update_workflow() {
  local key="$1" sid="$2" wfid="$3" name="$4" desc="$5" inputFields="$6" actLabel="$7" abi="$8" fn="$9" funcArgs="${10}"

  local body
  body=$(jq -nc \
    --arg id "$wfid" \
    --arg name "$name" \
    --arg desc "$desc" \
    --argjson inputFields "$inputFields" \
    --arg actLabel "$actLabel" \
    --arg net "$CHAIN_ID" \
    --arg addr "$AEGIS_ADDR" \
    --arg abi "$abi" \
    --arg fn "$fn" \
    --argjson funcArgs "$funcArgs" \
    '{
      jsonrpc:"2.0", id:20, method:"tools/call",
      params:{
        name:"update_workflow",
        arguments:{
          workflowId:$id,
          name:$name,
          description:$desc,
          nodes:[
            {id:"manual-trigger",type:"trigger",position:{x:100,y:200},
             data:{label:"Manual Trigger",type:"trigger",
                   config:{triggerType:"Manual", inputFields:$inputFields},
                   status:"idle"}},
            {id:"action",type:"action",position:{x:350,y:200},
             data:{label:$actLabel,type:"action",
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
    | tee /tmp/kh-update-raw.json \
    | jq -r '.result.content[0].text // .error.message // .'
}

update_org_workflows() {
  local idx="$1" key="$2"
  echo "=== org-$idx ==="
  local sid; sid=$(new_session "$key")
  echo "session=$sid"

  local commit_id reveal_id settle_id
  commit_id=$(toml_field "verifier_$idx" "commit_workflow_id")
  reveal_id=$(toml_field "verifier_$idx" "reveal_workflow_id")
  settle_id=$(toml_field "verifier_$idx" "settle_workflow_id")
  echo "  commit_id=$commit_id reveal_id=$reveal_id settle_id=$settle_id"

  # All inputFields use type="string" so KH's substitution keeps them as
  # strings end-to-end. web3/write-contract has the ABI in its config and
  # coerces each arg to its real Solidity type before calling the contract.
  # If inputFields claim uint256/bytes32/bool, KH casts to BigInt/Buffer/bool
  # at substitution time and writeContractStep blows up calling .trim() on
  # the typed value. Strings are safe.
  local IF_COMMIT='[{"name":"jobId","type":"string"},{"name":"commitHash","type":"string"}]'
  local IF_REVEAL='[{"name":"jobId","type":"string"},{"name":"verdict","type":"string"},{"name":"nonce","type":"string"}]'
  local IF_SETTLE='[{"name":"jobId","type":"string"}]'

  # functionArgs is the literal string KH expects (see action schema:
  # "string (JSON array of function arguments)"). KH calls .trim() on it,
  # substitutes {{@nodeId:Label.field}} references, then JSON.parses.
  # The values inside the resulting JSON are strings, which the ABI encoder
  # then coerces to uint256/bytes32/bool per the function signature.
  #
  # Encoded as JSON-string-literals so jq's --argjson reads them as strings.
  local FA_COMMIT='"[\"{{@manual-trigger:Manual Trigger.jobId}}\",\"{{@manual-trigger:Manual Trigger.commitHash}}\"]"'
  local FA_REVEAL='"[\"{{@manual-trigger:Manual Trigger.jobId}}\",\"{{@manual-trigger:Manual Trigger.verdict}}\",\"{{@manual-trigger:Manual Trigger.nonce}}\"]"'
  local FA_SETTLE='"[\"{{@manual-trigger:Manual Trigger.jobId}}\"]"'

  echo "  patching v${idx}/commit..."
  update_workflow "$key" "$sid" "$commit_id" \
    "aegis-commit-vote" \
    "Quorum: verifier commits a vote hash to AegisContract on 0G Galileo." \
    "$IF_COMMIT" "Commit Vote" "$ABI_COMMIT" "commitVote" "$FA_COMMIT" \
    | head -3

  echo "  patching v${idx}/reveal..."
  update_workflow "$key" "$sid" "$reveal_id" \
    "aegis-reveal-vote" \
    "Quorum: verifier reveals their vote on AegisContract." \
    "$IF_REVEAL" "Reveal Vote" "$ABI_REVEAL" "revealVote" "$FA_REVEAL" \
    | head -3

  echo "  patching v${idx}/settle..."
  update_workflow "$key" "$sid" "$settle_id" \
    "aegis-settle" \
    "Quorum: anyone can call settle on AegisContract once reveal deadline passes." \
    "$IF_SETTLE" "Settle" "$ABI_SETTLE" "settle" "$FA_SETTLE" \
    | head -3
  echo
}

update_org_workflows 1 "$ORG1_API_KEY"
update_org_workflows 2 "$ORG2_API_KEY"
update_org_workflows 3 "$ORG3_API_KEY"

echo "all 9 workflows updated. Re-run keeperhub-smoke to validate."
