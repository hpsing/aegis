<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api } from '@/api/client'
import { addrLink, storageLink, txLink } from '@/api/links'
import { knownAddr } from '@/api/tokens'
import type { JobDetail, ProofBundleRef, VerifierVote } from '@/api/types'
import { useSwarmStore } from '@/stores/swarm'
import Copyable from '@/components/Copyable.vue'

const props = defineProps<{ id: string }>()
const store = useSwarmStore()

const job = ref<JobDetail | null>(null)
const error = ref<string | null>(null)
const showBundleModal = ref(false)
const bundleBody = ref<unknown>(null)
const bundleErr = ref<string | null>(null)
const bundleLoading = ref(false)
const bundleSource = ref<ProofBundleRef | null>(null)

async function load() {
  try {
    job.value = await api.job(props.id)
    error.value = null
  } catch (e) {
    error.value = (e as Error).message
  }
}

let pollHandle: number | undefined
onMounted(() => {
  load()
  pollHandle = window.setInterval(() => {
    if (job.value?.status !== 'Settled' || (job.value?.proofBundles?.length ?? 0) < 3) {
      load()
    }
  }, 4000)
})
onUnmounted(() => {
  if (pollHandle) window.clearInterval(pollHandle)
})
watch(() => props.id, load)

watch(
  () => store.events[0],
  (ev) => {
    if (!ev) return
    const sameJob = (ev as { jobId?: string }).jobId === props.id
    if (sameJob && (ev.kind.startsWith('chain.') || ev.kind === 'og.upload')) {
      load()
    }
  },
)

interface VerifierColumn {
  address: string
  role: string
  iNFTId: string
  commit?: VerifierVote
  reveal?: VerifierVote
  bundle?: ProofBundleRef
}

const verifierColumns = computed<VerifierColumn[]>(() => {
  if (!job.value) return []
  const j = job.value
  const findVote = (list: VerifierVote[], addr: string) => {
    const a = addr.toLowerCase()
    return list.find((x) => x.verifier.toLowerCase() === a)
  }
  const findBundle = (list: ProofBundleRef[], addr: string) => {
    const a = addr.toLowerCase()
    return list.find((x) => x.verifier.toLowerCase() === a)
  }
  // Prefer registry-derived columns so v1/v2/v3 are stable. Fallback
  // to whatever addresses appear in the job's commits if roster empty.
  const sources = store.verifiers.length
    ? store.verifiers.map((v) => ({ address: v.address, role: v.role || '', iNFTId: v.iNFTId || '' }))
    : Array.from(new Set(j.commits.map((c) => c.verifier.toLowerCase())))
        .map((addr) => ({ address: addr, role: '', iNFTId: '' }))
  return sources.map((s) => ({
    address: s.address,
    role: s.role,
    iNFTId: s.iNFTId,
    commit: findVote(j.commits, s.address),
    reveal: findVote(j.reveals, s.address),
    bundle: findBundle(j.proofBundles, s.address),
  }))
})

async function openBundle(b: ProofBundleRef) {
  bundleSource.value = b
  showBundleModal.value = true
  bundleLoading.value = true
  bundleErr.value = null
  bundleBody.value = null
  try {
    bundleBody.value = await api.proofbundle(b.root)
  } catch (e) {
    bundleErr.value = (e as Error).message
  } finally {
    bundleLoading.value = false
  }
}

function closeBundle() {
  showBundleModal.value = false
  bundleBody.value = null
  bundleErr.value = null
  bundleSource.value = null
}

const status = computed(() => job.value?.status ?? '—')
const verdictClass = computed(() => {
  const v = job.value?.finalVerdict
  if (v === 'PASS') return 'text-accent-pass'
  if (v === 'FAIL') return 'text-accent-fail'
  return 'text-ink-2'
})

// Unanimous FAIL: every verifier voted against the executor's claim.
// Banner only fires for 3-of-3 FAIL so dissenter-FAIL outcomes (1-2
// verifiers voting FAIL while the rest PASS) don't get a "swarm
// rejected" framing — those are minority-quorum cases, not consensus.
const allRevealsAreFail = computed(() => {
  if (!job.value || job.value.reveals.length === 0) return false
  return job.value.reveals.every((r) => r.verdict === 'FAIL')
})

const allRevealsArePass = computed(() => {
  if (!job.value || job.value.reveals.length === 0) return false
  return job.value.reveals.every((r) => r.verdict === 'PASS')
})
</script>

<template>
  <div class="max-w-6xl mx-auto space-y-4 text-xs">
    <div v-if="error" class="text-accent-fail">{{ error }}</div>
    <div v-else-if="!job" class="text-ink-3">loading…</div>

    <template v-else>
      <!-- Header -->
      <header class="panel p-4 flex items-center justify-between">
        <div>
          <div class="text-ink-3 uppercase tracking-wider">Job</div>
          <div class="text-2xl text-ink-0">#{{ job.id }}</div>
        </div>
        <div class="flex gap-6 items-center">
          <div>
            <div class="text-ink-3 uppercase tracking-wider">status</div>
            <div :class="['text-lg', verdictClass]">
              {{ status }}<span v-if="job.finalVerdict"> · {{ job.finalVerdict }}</span>
            </div>
          </div>
          <div v-if="job.totalReveals !== undefined">
            <div class="text-ink-3 uppercase tracking-wider">reveals</div>
            <div class="text-lg text-ink-0">{{ job.forVotes ?? 0 }} / {{ job.totalReveals }}</div>
          </div>
        </div>
      </header>

      <!-- Unanimous PASS: solver did the work, all verifiers agreed -->
      <aside
        v-if="job.finalVerdict === 'PASS' && allRevealsArePass"
        class="panel border-accent-pass/40 bg-accent-pass/5 p-3 flex gap-3 items-start"
      >
        <span class="text-accent-pass text-lg leading-none mt-0.5">✓</span>
        <div class="space-y-1 flex-1">
          <div class="text-ink-0 font-bold">
            Unanimous PASS — solver delivered, swarm verified
          </div>
          <div class="text-ink-2">
            The <span class="text-ink-1">executor (solver)</span> performed the on-chain mUSDC transfer.
            All <strong>3 verifiers independently read the chain receipt</strong>, confirmed the
            Transfer event matched the spec (recipient + amount), and voted PASS.
          </div>
        </div>
      </aside>

      <!-- Unanimous FAIL: swarm rejected the claim -->
      <aside
        v-else-if="job.finalVerdict === 'FAIL' && allRevealsAreFail"
        class="panel border-accent-axl/40 bg-bg-subtle p-3 flex gap-3 items-start"
      >
        <span class="text-accent-axl text-lg leading-none mt-0.5">⚠</span>
        <div class="space-y-1 flex-1">
          <div class="text-ink-0 font-bold">
            Unanimous FAIL — swarm rejected the claim
          </div>
          <div class="text-ink-2">
            All 3 verifiers independently checked the chain and couldn't confirm the
            <code class="text-ink-1">Transfer</code> event matching the spec
            (recipient + amount on 0G Galileo). The swarm working as designed: an
            unverifiable claim gets rejected.
          </div>
          <div class="text-ink-3 pt-1 border-t border-bg-border">
            Older jobs may be from the legacy synthetic-claim path
            (<code class="text-ink-1">uniswap_v3_swap</code> built a receipt locally
            without broadcasting). New jobs from <strong>"Post Job…"</strong> use
            <code class="text-ink-1">mock_usdc_transfer</code> — a real on-chain
            mUSDC transfer — and should PASS.
          </div>
        </div>
      </aside>

      <!-- Parties -->
      <section class="panel">
        <div class="panel-title">Parties</div>
        <div class="p-3 grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div v-if="job.client">
            <div class="flex items-center gap-2 mb-1">
              <span class="text-ink-3 uppercase tracking-wider">client</span>
              <span class="pill bg-accent-chain/20 text-accent-chain" title="Posts the job, locks escrow, gets the swap output on PASS or a refund + slashed-stake compensation on FAIL.">poster</span>
            </div>
            <div class="text-ink-1">{{ knownAddr(job.client)?.name ?? '' }}</div>
            <Copyable :value="job.client" :href="addrLink(job.client)" size="sm" />
          </div>
          <div v-if="job.executor">
            <div class="flex items-center gap-2 mb-1">
              <span class="text-ink-3 uppercase tracking-wider">executor</span>
              <span class="pill bg-accent-axl/20 text-accent-axl" title="Bonded liquidity provider — same economic role as a UniswapX / CowSwap solver. Brings their own working capital to perform the swap, earns a fee on success, slashes on FAIL.">solver</span>
            </div>
            <div class="text-ink-1">{{ knownAddr(job.executor)?.name ?? 'fresh executor wallet' }}</div>
            <Copyable :value="job.executor" :href="addrLink(job.executor)" size="sm" />
          </div>
        </div>
      </section>

      <!-- Lifecycle -->
      <section class="panel">
        <div class="panel-title">Lifecycle</div>
        <div class="p-3 space-y-2.5">
          <div v-if="job.postJobTx" class="flex flex-wrap gap-2 items-baseline">
            <span class="text-ink-3 w-44 shrink-0">postJob (client)</span>
            <Copyable :value="job.postJobTx" :href="txLink(job.postJobTx)" size="sm" />
          </div>
          <div v-if="job.specHash" class="flex flex-wrap gap-2 items-baseline">
            <span class="text-ink-3 w-44 shrink-0">specHash</span>
            <Copyable :value="job.specHash" size="sm" />
          </div>
          <div v-if="job.submitClaimTx" class="flex flex-wrap gap-2 items-baseline">
            <span class="text-ink-3 w-44 shrink-0">submitClaim (executor)</span>
            <Copyable :value="job.submitClaimTx" :href="txLink(job.submitClaimTx)" size="sm" />
          </div>
          <div v-if="job.txHash" class="flex flex-wrap gap-2 items-baseline">
            <span class="text-ink-3 w-44 shrink-0">claimed swap tx</span>
            <Copyable :value="job.txHash" size="sm" />
          </div>
          <div v-if="job.settleTx" class="flex flex-wrap gap-2 items-baseline">
            <span class="text-ink-3 w-44 shrink-0">settle</span>
            <Copyable :value="job.settleTx" :href="txLink(job.settleTx)" size="sm" />
          </div>
        </div>
      </section>

      <!-- Verifier columns -->
      <section class="panel">
        <div class="panel-title">Verifiers</div>
        <div class="grid grid-cols-1 md:grid-cols-3 divide-x divide-bg-border">
          <div v-for="col in verifierColumns" :key="col.address" class="p-3 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-ink-0 font-bold">{{ col.role || 'verifier' }}</span>
              <span v-if="col.iNFTId" class="pill bg-accent-og/15 text-accent-og">iNFT #{{ col.iNFTId }}</span>
            </div>
            <Copyable :value="col.address" :href="addrLink(col.address)" size="sm" />

            <div class="space-y-1 pt-2 border-t border-bg-border">
              <div class="text-ink-3 uppercase tracking-wider">commit</div>
              <div v-if="col.commit" class="space-y-0.5">
                <div class="text-accent-pass">✓ committed</div>
                <Copyable :value="col.commit.txHash" :href="txLink(col.commit.txHash)" size="sm" />
              </div>
              <div v-else class="text-ink-3">—</div>
            </div>

            <div class="space-y-1 pt-2 border-t border-bg-border">
              <div class="text-ink-3 uppercase tracking-wider">reveal</div>
              <div v-if="col.reveal" class="space-y-0.5">
                <div :class="col.reveal.verdict === 'PASS' ? 'text-accent-pass' : 'text-accent-fail'">
                  {{ col.reveal.verdict ?? '?' }}
                </div>
                <Copyable :value="col.reveal.txHash" :href="txLink(col.reveal.txHash)" size="sm" />
              </div>
              <div v-else class="text-ink-3">—</div>
            </div>

            <div class="space-y-1 pt-2 border-t border-bg-border">
              <div class="text-ink-3 uppercase tracking-wider">proofbundle</div>
              <div v-if="col.bundle" class="space-y-1">
                <Copyable :value="col.bundle.root" :href="storageLink(col.bundle.root)" size="sm" />
                <button class="text-accent-axl hover:underline block" @click="openBundle(col.bundle)">view JSON ↓</button>
              </div>
              <div v-else class="text-ink-3">pending settle</div>
            </div>
          </div>
        </div>
      </section>

      <!-- Bundle modal -->
      <div v-if="showBundleModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" @click.self="closeBundle">
        <div class="panel max-w-3xl w-full max-h-[90vh] overflow-y-auto bg-bg-panel">
          <div class="panel-title flex items-center justify-between">
            <span>ProofBundle</span>
            <button class="text-ink-3 hover:text-ink-0" @click="closeBundle">✕</button>
          </div>
          <div v-if="bundleSource?.root" class="px-4 pt-3 text-xs">
            <div class="text-ink-3 uppercase tracking-wider mb-1">0G Storage root</div>
            <Copyable :value="bundleSource.root" :href="storageLink(bundleSource.root)" size="sm" />
          </div>
          <div v-if="bundleLoading" class="p-4 text-ink-3">downloading from 0G Storage…</div>
          <div v-else-if="bundleErr" class="p-4 text-accent-fail">{{ bundleErr }}</div>
          <pre v-else class="p-4 overflow-x-auto text-[11px] leading-tight">{{ JSON.stringify(bundleBody, null, 2) }}</pre>
        </div>
      </div>
    </template>
  </div>
</template>
