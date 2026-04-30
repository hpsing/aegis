# AI Tool Usage

## Tools used

- **Claude Code** (Anthropic's CLI coding agent) — primary AI assistant, used interactively from the terminal across multiple sessions.

## Where AI was used

| Area                            | Files / scope                                                                                                                                                                                                                                                                      | Nature of AI involvement                                                                                                                     |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Solidity contracts              | [contracts/src/AegisContract.sol](contracts/src/AegisContract.sol), [contracts/src/VerifierRegistry.sol](contracts/src/VerifierRegistry.sol), [contracts/src/VerifierINFT.sol](contracts/src/VerifierINFT.sol), [contracts/src/lib/AegisMath.sol](contracts/src/lib/AegisMath.sol) | Initial scaffolding and iterative edits assisted by AI; money flow, slashing rules, and commit-reveal protocol were human-specified.         |
| Forge tests                     | [contracts/test/](contracts/test/)                                                                                                                                                                                                                                                 | AI-generated test cases including reentrancy, edge paths, 2-of-3 majority, FAIL slash distribution; humans reviewed and added missing cases. |
| Go agents (CLIs + libs)         | [cmd/publisher](cmd/publisher), [cmd/executor](cmd/executor), [internal/client](internal/client), [internal/executor](internal/executor), [internal/envelope](internal/envelope), [internal/claim](internal/claim)                                                                 | AI-assisted; canonical-JSON spec hash is byte-anchored to Solidity via a known-vector test the human validated.                              |
| Deploy scripts + Makefile       | [contracts/script/](contracts/script/), [Makefile](Makefile)                                                                                                                                                                                                                       | AI-written; gas-price workaround for 0G Galileo (legacy txs, 2 gwei floor) found by human after on-chain failures.                           |
| End-to-end demo                 | [scripts/local-demo.sh](scripts/local-demo.sh)                                                                                                                                                                                                                                     | AI-written based on human-specified flow (register → post → claim → commit → reveal → settle).                                               |
| Architecture docs               | [docs/architecture.md](docs/architecture.md), [docs/agent-card-schema.md](docs/agent-card-schema.md)                                                                                                                                                                               | AI-assisted drafting; design decisions and trust-assumption choices were human-driven.                                                       |

## What was human-driven, not AI

- The **system architecture** itself: decomposition into client / executor / verifier roles, the 3-line-item escrow (reimbursement + fee + bounty), the 50/50 FAIL slash split, the 2-of-3 majority threshold, treasury routing, withdrawal cooldown.
- All decisions about **what to deploy, when, and to which chain**.
- All on-chain testing, gas-price diagnostics, and operational fixes.
- Reviewing every AI-written change before committing; rejecting or rewriting suggestions that didn't fit.
