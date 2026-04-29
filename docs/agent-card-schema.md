# Agent Card schema

The Agent Card is the public JSON document hosted on 0G Storage that
describes a verifier's capabilities and endpoints. The verifier's
ERC-7857 iNFT carries a pointer (`agent_card_uri`) to this card.

External observers (other agents, AIverse marketplace, etc.) read the
card to decide whether to interact with the verifier.

## Schema

```json
{
  "version": "1.0",
  "name": "Aegis Verifier #42",
  "description": "Verifies Uniswap V3 swap executions for slippage compliance",
  "type": "execution_verifier",
  "supported_actions": ["uniswap_v3_swap"],
  "endpoints": {
    "axl_peer_id": "0xabc... (64 hex chars)",
    "mcp_url": "https://verifier-42.example/mcp"
  },
  "wallet_address": "0x... (20 bytes, EIP-55)",
  "registered_at_block": 123456,
  "stake_token": "USDC",
  "min_stake": "100",
  "verifier_protocol_version": 1
}
```

## Field semantics

| Field                       | Type        | Notes                                                                                       |
| --------------------------- | ----------- | ------------------------------------------------------------------------------------------- |
| `version`                   | string      | Schema version. Bump on breaking change.                                                    |
| `name`                      | string      | Human-readable, free-form.                                                                  |
| `description`               | string      | One-line summary.                                                                           |
| `type`                      | string enum | Currently only `execution_verifier`. Future: `claim_oracle`, `data_signer`, etc.            |
| `supported_actions`         | string[]    | Spec actions this verifier accepts. Step 4–6 only support `uniswap_v3_swap`.                |
| `endpoints.axl_peer_id`     | string      | Verifier's AXL public key, hex-encoded. Used by other agents to send signed envelopes.      |
| `endpoints.mcp_url`         | string      | KeeperHub-compatible MCP URL (step 7). Optional.                                            |
| `wallet_address`            | string      | The EOA registered in `VerifierRegistry`. Should equal `controllerOf(tokenId)` on the iNFT. |
| `registered_at_block`       | int         | Block number on the AegisContract chain where the verifier first registered.                |
| `stake_token`               | string      | Token symbol. Currently always `USDC`.                                                      |
| `min_stake`                 | string      | Minimum stake (decimal, in token units).                                                    |
| `verifier_protocol_version` | int         | Bumps when the verification logic changes (e.g. step 4 = 1, future ZK-attested = 2).        |

## Where it lives

- **On 0G Storage** at `agent_card_uri`. Storage is append-only / log-style;
  to update the card, upload a new version and call
  `VerifierINFT.setAgentCardURI(tokenId, newURI)`.
- **Linked from** `VerifierINFT.agentCardURI(tokenId)`.

## Vote-history Log

A separate 0G Storage append-only Log per verifier holds their vote
history. The contract emits enough state on each settle for an external
observer to reconstruct the verifier's accuracy directly from chain
events — the 0G Log is a convenience that lets new auditors fetch the
full history with one read.

Each entry shape:

```json
{
  "job_id": 42,
  "verdict_voted": "PASS",
  "verdict_final": "PASS",
  "was_correct": true,
  "timestamp": 1714152600000,
  "settle_tx_hash": "0x..."
}
```
