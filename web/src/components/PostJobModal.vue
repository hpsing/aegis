<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/client'
import { addrLink, fmtMUSDC } from '@/api/links'
import { knownAddr } from '@/api/tokens'
import type { KnownAddr } from '@/api/tokens'
import type { PostJobPreview } from '@/api/types'
import Copyable from '@/components/Copyable.vue'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()

const preview = ref<PostJobPreview | null>(null)
const loadErr = ref<string | null>(null)

// Pre-resolve known-address lookups so the template stays cast-free
// (Vue's template parser doesn't accept `expr as Type` inline).
const tokenInInfo = computed<KnownAddr | undefined>(() => knownAddr(preview.value?.spec.tokenIn))
const tokenOutInfo = computed<KnownAddr | undefined>(() => knownAddr(preview.value?.spec.tokenOut))
const clientInfo = computed<KnownAddr | undefined>(() => knownAddr(preview.value?.client))
const executorInfo = computed<KnownAddr | undefined>(() => knownAddr(preview.value?.executor))

async function load() {
  try {
    preview.value = await api.postJobPreview()
    loadErr.value = null
  } catch (e) {
    loadErr.value = (e as Error).message
  }
}

function onConfirm() {
  if (!preview.value?.ready) return
  // Modal does not own the API call — emit intent, parent dispatches
  // and runs the toast pipeline. Modal closes immediately so the
  // operator can watch the timeline / topology react in real time.
  emit('confirm')
}

onMounted(load)
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" @click.self="emit('close')">
    <div class="panel max-w-2xl w-full max-h-[90vh] overflow-y-auto bg-bg-panel">
      <div class="panel-title flex items-center justify-between">
        <span>Confirm post job</span>
        <button class="text-ink-3 hover:text-ink-0" @click="emit('close')">✕</button>
      </div>

      <div v-if="loadErr" class="p-4 text-accent-fail text-xs">{{ loadErr }}</div>
      <div v-else-if="!preview" class="p-4 text-ink-3 text-xs">Loading preview…</div>

      <div v-else class="p-4 space-y-4 text-xs">
        <!-- Live-mode notice: explain what's actually going to happen
             on chain so the operator knows what they're triggering. -->
        <aside class="panel border-accent-pass/40 bg-accent-pass/5 p-3 flex gap-3 items-start">
          <span class="text-accent-pass text-base leading-none mt-0.5">⚡</span>
          <div class="space-y-1">
            <div class="text-accent-pass font-bold">Live mUSDC transfer · expect PASS</div>
            <div class="text-ink-2">
              The <span class="text-ink-1">executor is a solver</span>
              (UniswapX / CowSwap-style). When you confirm, the solver will:
              <span class="text-ink-1">(1)</span> mint mUSDC to itself if low,
              <span class="text-ink-1">(2)</span> sign a real
              <code class="text-ink-1">mUSDC.transfer(client, amount)</code> on 0G Galileo,
              <span class="text-ink-1">(3)</span> submit the claim on chain. All 3 verifiers
              independently read the Transfer event, confirm recipient + amount, and vote PASS.
            </div>
          </div>
        </aside>

        <!-- Spec -->
        <section>
          <div class="text-ink-3 uppercase tracking-wider mb-2">What this job verifies</div>
          <div class="panel bg-bg-subtle/50 p-3 space-y-2">
            <div class="text-ink-0 text-sm">
              {{ preview.spec.intent }} · <span class="text-accent-axl">{{ preview.spec.action }}</span>
            </div>
            <dl class="grid grid-cols-[120px_1fr] gap-y-1.5 items-baseline">
              <dt class="text-ink-3">chain</dt>
              <dd>{{ preview.spec.chainId }} · Base mainnet</dd>

              <dt class="text-ink-3">tokenIn</dt>
              <dd>
                <div v-if="tokenInInfo" class="text-ink-1">
                  <span class="text-accent-pass">{{ tokenInInfo.symbol }}</span>
                  <span class="text-ink-3"> · {{ tokenInInfo.name }}</span>
                </div>
                <Copyable :value="preview.spec.tokenIn" :href="addrLink(preview.spec.tokenIn)" size="sm" />
              </dd>

              <dt class="text-ink-3">tokenOut</dt>
              <dd>
                <div v-if="tokenOutInfo" class="text-ink-1">
                  <span class="text-accent-pass">{{ tokenOutInfo.symbol }}</span>
                  <span class="text-ink-3"> · {{ tokenOutInfo.name }}</span>
                </div>
                <Copyable :value="preview.spec.tokenOut" :href="addrLink(preview.spec.tokenOut)" size="sm" />
              </dd>

              <dt class="text-ink-3">amountIn</dt>
              <dd>{{ preview.spec.amountIn }} <span class="text-ink-3">(raw, 6-decimal)</span></dd>

              <dt class="text-ink-3">max slippage</dt>
              <dd>{{ preview.spec.maxSlippageBps }} bps</dd>

              <dt class="text-ink-3">spec hash</dt>
              <dd><Copyable :value="preview.specHash" size="sm" /></dd>
            </dl>
          </div>
        </section>

        <!-- Parties -->
        <section>
          <div class="text-ink-3 uppercase tracking-wider mb-2">Parties</div>
          <div class="panel bg-bg-subtle/50 p-3 space-y-2">
            <div>
              <div class="flex items-center gap-2">
                <span class="text-ink-3">client (signs postJob)</span>
                <span class="pill bg-accent-chain/20 text-accent-chain">poster</span>
              </div>
              <div class="text-ink-1">{{ clientInfo?.name ?? '—' }}</div>
              <Copyable :value="preview.client" :href="addrLink(preview.client)" size="sm" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="text-ink-3">executor (signs submitClaim)</span>
                <span class="pill bg-accent-axl/20 text-accent-axl" title="Bonded liquidity provider — UniswapX / CowSwap-style solver">solver</span>
              </div>
              <div class="text-ink-1">{{ executorInfo?.name ?? 'fresh executor wallet' }}</div>
              <Copyable :value="preview.executor" :href="addrLink(preview.executor)" size="sm" />
            </div>
          </div>
        </section>

        <!-- Escrow -->
        <section>
          <div class="text-ink-3 uppercase tracking-wider mb-2">Escrow (mUSDC)</div>
          <div class="panel bg-bg-subtle/50 p-3 grid grid-cols-2 gap-y-1">
            <span class="text-ink-3">executor reimbursement</span><span class="text-right">{{ fmtMUSDC(preview.reimbursement) }}</span>
            <span class="text-ink-3">executor fee (on PASS)</span><span class="text-right">{{ fmtMUSDC(preview.fee) }}</span>
            <span class="text-ink-3">verifier bounty</span><span class="text-right">{{ fmtMUSDC(preview.bounty) }}</span>
            <span class="text-ink-0 border-t border-bg-border pt-1">total locked</span>
            <span class="text-ink-0 text-right border-t border-bg-border pt-1">{{ fmtMUSDC(preview.totalEscrow) }}</span>
          </div>
        </section>

        <!-- Verifiers -->
        <section>
          <div class="text-ink-3 uppercase tracking-wider mb-2">Spec will be AXL-published to</div>
          <div class="panel bg-bg-subtle/50 p-3 space-y-2">
            <div v-for="p in preview.verifierPeers" :key="p.role">
              <div class="flex items-baseline gap-2">
                <span class="text-ink-0 font-bold">{{ p.role }}</span>
                <span class="text-ink-3">AXL peer</span>
              </div>
              <Copyable :value="p.peerId" size="sm" />
            </div>
            <div v-if="!preview.verifierPeers.length" class="text-accent-fail">no online verifier peers</div>
          </div>
        </section>

        <!-- Blockers -->
        <section v-if="preview.blockers.length">
          <div class="text-accent-fail uppercase tracking-wider mb-2">Blockers</div>
          <ul class="panel bg-accent-fail/10 border-accent-fail/40 p-3 space-y-0.5 text-accent-fail">
            <li v-for="b in preview.blockers" :key="b">• {{ b }}</li>
          </ul>
        </section>

        <!-- Actions -->
        <div class="flex justify-between items-center pt-2 border-t border-bg-border">
          <span class="text-ink-3">
            <span v-if="preview.ready">Editing the spec is not yet wired — server uses the canonical demo spec.</span>
            <span v-else>Resolve the blockers above to enable Confirm.</span>
          </span>
          <div class="flex gap-2">
            <button class="px-3 py-1.5 rounded bg-bg-subtle hover:bg-bg-border" @click="emit('close')">Cancel</button>
            <button
              class="px-3 py-1.5 rounded bg-accent-chain/30 text-accent-chain hover:bg-accent-chain/40 disabled:opacity-40 disabled:cursor-not-allowed"
              :disabled="!preview.ready"
              @click="onConfirm"
            >
              Confirm post job
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
