<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSwarmStore } from '@/stores/swarm'
import { khOrgLink, shortHash, txLink } from '@/api/links'
import type { ServerEvent } from '@/api/types'
import TopologyGraph from '@/components/TopologyGraph.vue'

const store = useSwarmStore()
const router = useRouter()

const activeJob = computed(() => store.recentJobs[0])

function isJobScoped(ev: ServerEvent): ev is ServerEvent & { jobId: string } {
  return 'jobId' in ev
}

const timelineEvents = computed(() => {
  const id = activeJob.value?.id
  return store.events.filter((e) => {
    if (e.kind === 'postjob.step') {
      const ej = (e as { jobId?: string }).jobId
      return !id || ej === id || !ej
    }
    return id ? isJobScoped(e) && e.jobId === id : false
  })
})

interface TimelineRow {
  ts: number
  label: string
  className: string
  tx?: string
}

function rowFor(ev: ServerEvent): TimelineRow {
  if (ev.kind === 'postjob.step') {
    const e = ev as { stage: string; msg: string; tx?: string }
    return {
      ts: ev.ts,
      label: `↗ ${e.stage}: ${e.msg}`,
      className: 'text-accent-axl',
      tx: e.tx,
    }
  }
  if (ev.kind.startsWith('chain.')) {
    const e = ev as { tx?: string; verifier?: string; verdict?: string; finalVerdict?: string }
    const labelMap: Record<string, string> = {
      'chain.JobPosted': '⛓ JobPosted',
      'chain.ClaimSubmitted': '⛓ ClaimSubmitted',
      'chain.VoteCommitted': '⛓ VoteCommitted',
      'chain.VoteRevealed': '⛓ VoteRevealed',
      'chain.JobSettled': '⛓ JobSettled',
    }
    let label = labelMap[ev.kind] ?? ev.kind.replace('chain.', '')
    if (e.verifier) label += ` · ${shortHash(e.verifier)}`
    if (e.verdict) label += ` · ${e.verdict}`
    if (e.finalVerdict) label += ` · ${e.finalVerdict}`
    return { ts: ev.ts, label, className: 'text-ink-1', tx: e.tx }
  }
  return { ts: ev.ts, label: ev.kind, className: 'text-ink-2' }
}

const timelineRows = computed<TimelineRow[]>(() => timelineEvents.value.map(rowFor))

interface KHEvent {
  kind: 'kh.workflow'
  verifier: string
  role?: string
  purpose: 'commit' | 'reveal' | 'settle'
  jobId: string
  tx?: string
  ts: number
}

const khEvents = computed<KHEvent[]>(() =>
  store.events.filter((e): e is KHEvent => e.kind === 'kh.workflow').slice(0, 50),
)

function gotoJob() {
  if (activeJob.value?.id) router.push(`/jobs/${activeJob.value.id}`)
}
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-[minmax(280px,1fr)_minmax(380px,1.4fr)_minmax(280px,1fr)] gap-3 h-[calc(100vh-7rem)]">
    <!-- Left: timeline -->
    <section class="panel flex flex-col">
      <div class="panel-title flex items-center justify-between">
        <span>Job timeline</span>
        <button v-if="activeJob" class="text-accent-chain hover:underline normal-case tracking-normal" @click="gotoJob">
          #{{ activeJob.id }} ↗
        </button>
      </div>
      <div class="flex-1 overflow-y-auto p-3 text-xs space-y-1.5">
        <div v-if="!activeJob && !timelineRows.length" class="text-ink-3">Waiting for first job…</div>
        <div v-for="(row, i) in timelineRows" :key="i" class="flex gap-2 items-baseline">
          <span class="text-ink-3 tabular-nums">{{ new Date(row.ts).toLocaleTimeString() }}</span>
          <span :class="row.className" class="flex-1 truncate">{{ row.label }}</span>
          <a v-if="row.tx" :href="txLink(row.tx)" target="_blank" class="font-mono text-ink-3 hover:text-accent-chain">{{ shortHash(row.tx) }}↗</a>
        </div>
      </div>
    </section>

    <!-- Center: topology graph -->
    <section class="panel flex flex-col">
      <div class="panel-title flex items-center justify-between">
        <span>AXL mesh</span>
        <span class="text-ink-2 normal-case tracking-normal">
          {{ store.topology.nodes.length }} nodes · refresh
          {{ store.topology.freshnessMs ? new Date(store.topology.freshnessMs).toLocaleTimeString() : '—' }}
        </span>
      </div>
      <div class="flex-1 p-2">
        <TopologyGraph />
      </div>
    </section>

    <!-- Right: KH activity -->
    <section class="panel flex flex-col">
      <div class="panel-title flex items-center justify-between">
        <span>KeeperHub activity (audit trail)</span>
        <a :href="khOrgLink()" target="_blank" class="text-ink-3 hover:text-accent-chain normal-case tracking-normal">dashboard ↗</a>
      </div>
      <!-- Persistent context: the KH dashboard shows red workflow
           errors ("Invalid function arguments", "Failed to acquire
           nonce lock") that look broken at first glance. They're
           expected per the upstream bugs we reported — the real tx
           still lands via the local-broadcast leg. -->
      <div class="px-3 py-2 border-b border-bg-border bg-bg-subtle/50 text-[10px] leading-relaxed text-ink-3">
        Each row = one <code class="text-ink-1">execute_workflow</code> fired into KH for audit.
        KH-side errors on the workflow page are <span class="text-ink-1">expected</span>
        (upstream bugs in their <code class="text-ink-1">web3/write-contract</code> action — see
        <code class="text-ink-1">internal/keeperhub/live.go</code>); the on-chain tx
        <strong class="text-ink-1">does land</strong> — the verifier signs + broadcasts locally
        in parallel. The link in each row points to that real tx.
      </div>
      <div class="flex-1 overflow-y-auto p-3 text-xs space-y-1.5">
        <div v-if="!khEvents.length" class="text-ink-3">No invocations yet.</div>
        <div v-for="(ev, i) in khEvents" :key="i" class="flex gap-2 items-baseline">
          <span class="text-ink-3 tabular-nums">{{ new Date(ev.ts).toLocaleTimeString() }}</span>
          <span class="text-ink-1">{{ ev.role || shortHash(ev.verifier) }}</span>
          <span class="text-accent-kh">{{ ev.purpose }}</span>
          <span class="text-ink-2 flex-1 truncate">job {{ ev.jobId }}</span>
          <a v-if="ev.tx" :href="txLink(ev.tx)" target="_blank" class="font-mono text-ink-3 hover:text-accent-chain" title="The on-chain tx (local broadcast leg)">{{ shortHash(ev.tx) }}↗</a>
        </div>
      </div>
    </section>
  </div>
</template>
