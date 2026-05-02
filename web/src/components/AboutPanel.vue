<script setup lang="ts">
defineEmits<{ (e: 'close'): void }>()
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" @click.self="$emit('close')">
    <div class="panel max-w-3xl w-full max-h-[90vh] overflow-y-auto bg-bg-panel">
      <div class="panel-title flex items-center justify-between">
        <span>About this demo</span>
        <button class="text-ink-3 hover:text-ink-0" @click="$emit('close')">✕</button>
      </div>

      <div class="p-4 space-y-4 text-xs">
        <section>
          <h3 class="text-ink-0 text-sm font-bold mb-2">What you're looking at</h3>
          <p class="text-ink-2">
            Aegis is a peer-to-peer swarm of independent verifier agents that checks whether
            on-chain agent executions actually matched their spec. When one agent claims to
            have done a paid job for another, three verifiers re-check the chain, vote
            commit-reveal, and settle escrow.
          </p>
        </section>

        <section>
          <h3 class="text-ink-0 text-sm font-bold mb-2">The actors</h3>
          <dl class="grid grid-cols-[120px_1fr] gap-y-2">
            <dt class="flex items-start gap-2">
              <span class="pill bg-accent-chain/20 text-accent-chain">poster</span>
              <span class="text-ink-3">client</span>
            </dt>
            <dd class="text-ink-2">
              Posts the job. Locks escrow (executor reimbursement + fee + verifier bounty).
              Gets the swap output on PASS, or a refund + half of the executor's slashed stake
              on FAIL. <span class="text-ink-3">Principal never at risk.</span>
            </dd>

            <dt class="flex items-start gap-2">
              <span class="pill bg-accent-axl/20 text-accent-axl">solver</span>
              <span class="text-ink-3">executor</span>
            </dt>
            <dd class="text-ink-2">
              <strong>Same economic role as a UniswapX / CowSwap solver.</strong>
              Bonded liquidity provider — brings their own working capital to perform the
              on-chain work, submits the claim, earns a service fee on PASS. On FAIL: working
              capital is gone (recipient_mismatch invariant) + 25% bond slashed.
              <span class="text-ink-3">Asymmetric: small fee on success, large capital loss on failure → strong honesty incentive.</span>
            </dd>

            <dt class="flex items-start gap-2">
              <span class="pill bg-accent-pass/20 text-accent-pass">verifier</span>
              <span class="text-ink-3">×3 minimum</span>
            </dt>
            <dd class="text-ink-2">
              Independently re-checks the executor's claim against on-chain truth.
              100 mUSDC stake. Equal share of bounty if in majority; 10% stake slashed if in
              minority. <span class="text-ink-3">BFT-style 2-of-3 majority — tolerates one dishonest verifier.</span>
            </dd>
          </dl>
        </section>

        <section>
          <h3 class="text-ink-0 text-sm font-bold mb-2">The four rails</h3>
          <ul class="text-ink-2 space-y-1 list-disc list-inside marker:text-ink-3">
            <li><strong class="text-ink-1">0G Chain</strong> — the contract: holds escrow, records votes, settles.</li>
            <li><strong class="text-ink-1">Gensyn AXL</strong> — peer-to-peer mesh that carries the job spec from publisher to verifiers.</li>
            <li><strong class="text-ink-1">KeeperHub</strong> — third-party audit dashboard for every settlement tx. <span class="text-ink-3">See dual-leg note below.</span></li>
            <li><strong class="text-ink-1">0G Storage</strong> — per-job ProofBundle + per-verifier accuracy log, content-addressed by merkle root.</li>
          </ul>
        </section>

        <section class="panel border-accent-kh/40 bg-accent-kh/5 p-3">
          <h3 class="text-accent-kh text-sm font-bold mb-1">KeeperHub dual-leg pattern</h3>
          <p class="text-ink-2">
            Every commit / reveal / settle goes through KeeperHub via TWO parallel legs:
          </p>
          <ol class="text-ink-2 list-decimal list-inside mt-1 space-y-0.5 marker:text-ink-3">
            <li>
              <span class="text-ink-1">Audit-trail leg</span> — fires
              <code class="text-ink-1">execute_workflow</code> async to KH so the workflow
              invocation shows up on KH's dashboard.
              <span class="text-ink-3">This is what the "KeeperHub activity" panel renders.</span>
            </li>
            <li>
              <span class="text-ink-1">Broadcast leg</span> — verifier signs the tx locally with
              its own key (msg.sender = verifier) and broadcasts at 2 gwei tip cap.
              <span class="text-ink-3">This is what actually lands on chain.</span>
            </li>
          </ol>
          <p class="text-ink-2 mt-1.5">
            <strong>Heads-up:</strong> the KH dashboard often shows the audit-trail leg as
            <span class="text-accent-fail">failed</span>
            (<code class="text-ink-1">jobId: uint256 is missing</code>,
            <code class="text-ink-1">Failed to acquire nonce lock</code>) — these are
            <strong>known upstream KH bugs we reported</strong> (gas tip cap too low for 0G;
            <code class="text-ink-1">call_workflow</code> + nonce-lock issues). The on-chain
            tx still lands via the broadcast leg. We keep both legs running so the audit
            record exists once KH's bugs are fixed.
          </p>
          <p class="text-ink-3 mt-1">
            See <code class="text-ink-1">internal/keeperhub/live.go</code> for the full
            background + the bug reports.
          </p>
        </section>

        <section class="panel border-accent-axl/40 bg-accent-axl/5 p-3">
          <h3 class="text-accent-axl text-sm font-bold mb-1">Why FAIL? (synthetic-claim demo mode)</h3>
          <p class="text-ink-2">
            This build runs in <strong>synthetic-claim mode</strong>. The solver builds the swap
            receipt locally — the tx never broadcasts. Verifiers correctly can't find it on chain
            and vote FAIL. The unanimous-FAIL outcome is the swarm working as designed: a phantom
            claim gets rejected.
          </p>
          <p class="text-ink-3 mt-1">
            Production (architecture step 7) has the solver perform a real Uniswap V3 swap on Base;
            verifiers find the receipt and vote PASS.
          </p>
        </section>

        <section class="text-ink-3">
          Full economic model + threat model: see <code class="text-ink-1">docs/architecture.md</code> in the repo.
        </section>
      </div>
    </div>
  </div>
</template>
