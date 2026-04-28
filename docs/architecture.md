# Architecture

## What is Quorum?

**Quorum** is a peer-to-peer swarm of independent verifier agents that checks whether on-chain executions actually met their specification. When one agent claims to have completed a paid job for another, the swarm re-checks the on-chain outcome, votes via Byzantine-fault-tolerant consensus, and settles escrow accordingly. Dishonest verifiers are slashed; accurate ones earn fees and accumulate reputation as ERC-7857 iNFTs on 0G.

## Why Quorum?

The agent-economy stack as of now:

| Layer                            | Status      | Standard                                                    |
| -------------------------------- | ----------- | ----------------------------------------------------------- |
| Agent identity                   | **Solved**  | ERC-8004 (live on Ethereum mainnet, Jan 2026)               |
| Agent payments                   | **Solved**  | x402 (Coinbase, with Stripe/Cloudflare/AWS support)         |
| Agent reputation                 | **Partial** | ERC-8004 reputation registry, but accuracy is self-attested |
| **Agent execution verification** | **Empty**   | **Specifically left to "the app layer" by ERC-8004 spec**   |

x402 settles on HTTP 200, not on verified outcome. ERC-8004's validation registry is a slot the standard reserves but does not fill. Today, "Agent B says it executed your trade correctly" and "Agent B actually executed your trade correctly" are unverifiable claims unless you trust a single oracle. Which defeats the point of a decentralized agent economy.

Quorum fills the empty validation slot with a swarm of staked, independent verifiers running on separate machines. Trust is replaced by economically
costly dishonesty.

## The Actors

| Actor                                           | Role                                                                                                                                                                                                                                                                   | What they put in                                                                                                                                                                            | What they get back                                                                                                                                                                                                                                                                                                                           |
| ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Client agent**                                | Wants a job done. Posts the job spec, locks executor reimbursement + fee + bounty in escrow. **Does not fund the swap directly** — the executor brings working capital.                                                                                                | `executor_reimbursement` (e.g. 10,000 USDC, refundable on FAIL) + `executor_fee` (e.g. 30 USDC, refundable on FAIL) + `verifier_bounty` (e.g. 20 USDC, consumed regardless) + gas           | If PASS: gets the swap output (ETH directly from Uniswap, recipient=client). Net cost: ~50 USDC overhead on a 10,000 USDC job. If FAIL: gets the (bad) swap output + 10,030 USDC refund + half of executor's slashed stake (~62.5 USDC compensation). **Principal is never at risk; only the small fee + bounty.**                           |
| **Executor agent**                              | **Bonded liquidity provider.** Reads job specs over AXL, brings their own working capital to perform the on-chain work (e.g. floats the 10,000 USDC for a Uniswap swap), earns a service fee on successful execution. Same economic role as a UniswapX/CowSwap solver. | Their own working capital (e.g. 10,000 USDC, temporarily — locked in the swap, ETH goes straight to client) + 500 USDC long-term stake/bond + gas for the swap tx + gas for `submitClaim()` | If PASS: working capital fully reimbursed from escrow + executor fee (e.g. 30 USDC). If FAIL: working capital is gone (the ETH from the swap went to client per the `recipient_mismatch` invariant; client keeps it) + 25% of stake slashed. **Asymmetric: small fee on success, large capital loss on failure → strong honesty incentive.** |
| **Verifier agent** (×3 minimum, more is better) | Independently checks the executor's claim. Runs on its own AXL node.                                                                                                                                                                                                   | 100 USDC stake (locked in `VerifierRegistry`) + gas for commit + reveal txs                                                                                                                 | Equal share of `verifier_bounty` if in majority (e.g. 6.67 USDC each on a 20-USDC bounty split 3 ways; more if some verifiers forfeited their share). 10% stake slash if in minority.                                                                                                                                                        |
| **The Quorum contract**                         | Holds escrow. Tallies votes. Settles.                                                                                                                                                                                                                                  | (no balance of its own; just routes funds)                                                                                                                                                  | (no profit; gas-neutral)                                                                                                                                                                                                                                                                                                                     |
| **The protocol treasury**                       | Holds slashed funds from minority/dishonest verifiers.                                                                                                                                                                                                                 | Receives slashed amounts                                                                                                                                                                    | (in a future step: distributed to honest long-term verifiers as a bonus, or burned.)                                                                                                                                                                                                                                 |
| **KeeperHub**                                   | Reliable execution rail for settlement txs (commit, reveal, settle, slash).                                                                                                                                                                                            | (its own service infra)                                                                                                                                                                     | Service fee on each tx (paid by the verifier or settler in gas-equivalent)                                                                                                                                                                                                                                                                   |
| **Gensyn AXL**                                  | Peer-to-peer transport layer between agents.                                                                                                                                                                                                                           | (its own infra; verifiers run AXL nodes)                                                                                                                                                    | (free for the hackathon; long-term: $AI token incentives for relays)                                                                                                                                                                                                                                                                         |
| **0G Storage**                                  | Append-only log of vote history per verifier; ProofBundle storage.                                                                                                                                                                                                     | Storage fees in $OG (small, per-write)                                                                                                                                                      | (the data — public, queryable)                                                                                                                                                                                                                                                                                                               |
| **0G Chain**                                    | Hosts ERC-7857 verifier iNFTs.                                                                                                                                                                                                                                         | Gas in $OG for mint, transfer, controller updates                                                                                                                                           | (identity persistence)                                                                                                                                                                                                                                                                                                                       |

## System Architecture

```text
                          ┌───────────────────────────────┐
                          │       CLIENT AGENT            │
                          │  (job poster, holds USDC)     │
                          └──────────────┬────────────────┘
                                         │ 1. postJob()
                                         │    locks 10,050 USDC
                                         ▼
   ┌─────────────────────────────────────────────────────────────────┐
   │                  QUORUM CONTRACT (on chain)                     │
   │   ┌─────────┐  ┌──────────────┐  ┌─────────┐  ┌─────────────┐   │
   │   │ escrow  │  │ commit-reveal│  │settle() │  │  events     │   │
   │   │  vault  │  │   ledger     │  │  logic  │  │  for agents │   │
   │   └─────────┘  └──────────────┘  └─────────┘  └─────────────┘   │
   └────────────┬────────────────┬───────────────────┬───────────────┘
                │                │                   │
                │ JobPosted      │ ClaimSubmitted    │ JobSettled
                ▼                ▼                   ▼
         ┌────────────┐    ┌────────────┐    ┌──────────────┐
         │  EXECUTOR  │    │ VERIFIERS  │    │   CLIENT     │
         │   AGENT    │    │ × 3 over   │    │  + 0G        │
         │            │    │   AXL      │    │   Storage    │
         └─────┬──────┘    └─────┬──────┘    └──────────────┘
               │                 │
               │ 2. perform      │ 3. fetch tx receipt,
               │    Uniswap swap │    decode swap event,
               │ 3. submitClaim()│    compute slippage,
               ▼                 │    commit + reveal
       ┌─────────────────┐       │
       │ Uniswap V3 Pool │ ◄─────┘  via go-ethereum/RPC (read-only)
       │   on Base       │
       └─────────────────┘

   ┌─────────────────────────────────────────────────────────────────┐
   │            GENSYN AXL MESH (transport for ALL agents)           │
   │     • verifier ↔ verifier gossip                                │
   │     • client ↔ executor handshake                               │
   │     • encrypted, peer-to-peer, no broker                        │
   └─────────────────────────────────────────────────────────────────┘

   ┌─────────────────────────────────────────────────────────────────┐
   │            KEEPERHUB MCP (settlement-tx rail)                   │
   │     Every commit/reveal/settle/slash tx routes through here     │
   │     for retry, gas optimization, MEV protection                 │
   └─────────────────────────────────────────────────────────────────┘

   ┌─────────────────────────────────────────────────────────────────┐
   │            0G STORAGE & CHAIN                                   │
   │     • Verifier iNFTs (ERC-7857) on 0G Chain                     │
   │     • Vote-history Logs on 0G Storage (append-only)             │
   │     • ProofBundles per settled job on 0G Storage                │
   └─────────────────────────────────────────────────────────────────┘
```
