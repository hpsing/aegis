import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import type {
  JobSummary,
  ServerEvent,
  StateSnapshot,
  SystemTotals,
  TopologyState,
  VerifierInfo,
} from '@/api/types'

// Single global store. Holds the latest snapshot from /api/state plus
// a rolling event log fed by SSE. Views read what they need; reducer
// here decides how each event mutates state.
export const useSwarmStore = defineStore('swarm', () => {
  const totals = ref<SystemTotals>({
    jobsVerified: 0,
    passRateBps: 0,
    totalKHInvocations: 0,
    meanCommitToSettleSec: 0,
    axlMessageRatePerMin: 0,
    ogUploadsTotal: 0,
    swarmOnline: 0,
    swarmTotal: 3,
  })
  const recentJobs = ref<JobSummary[]>([])
  const verifiers = ref<VerifierInfo[]>([])
  const topology = ref<TopologyState>({ freshnessMs: 0, nodes: [], edges: [] })
  const events = ref<ServerEvent[]>([])
  const connected = ref(false)
  const lastError = ref<string | null>(null)

  const passRatePct = computed(() => Math.round(totals.value.passRateBps / 100))

  async function loadSnapshot() {
    try {
      const snap: StateSnapshot = await api.state()
      totals.value = snap.totals
      recentJobs.value = snap.recentJobs
      verifiers.value = snap.verifiers
      topology.value = snap.topology
      lastError.value = null
    } catch (e) {
      lastError.value = (e as Error).message
    }
  }

  function applyEvent(ev: ServerEvent) {
    events.value.unshift(ev)
    if (events.value.length > 500) events.value.length = 500

    switch (ev.kind) {
      case 'totals.update':
        totals.value = ev.totals
        break
      case 'topology.update':
        topology.value = ev.topology
        break
      case 'chain.JobPosted': {
        const exists = recentJobs.value.some((j) => j.id === ev.jobId)
        if (!exists) {
          recentJobs.value.unshift({
            id: ev.jobId,
            status: 'Posted',
            client: ev.client,
            executor: ev.executor,
            postedBlock: 0,
            postedAt: ev.ts,
          })
        }
        break
      }
      case 'chain.ClaimSubmitted': {
        const job = recentJobs.value.find((j) => j.id === ev.jobId)
        if (job) job.status = 'ClaimSubmitted'
        break
      }
      case 'chain.JobSettled': {
        const job = recentJobs.value.find((j) => j.id === ev.jobId)
        if (job) {
          job.status = 'Settled'
          job.finalVerdict = ev.finalVerdict
          job.forVotes = ev.forVotes
          job.totalReveals = ev.totalReveals
          job.settledAt = ev.ts
        }
        break
      }
    }
  }

  let sse: EventSource | null = null
  function connectSSE() {
    if (sse) return
    sse = new EventSource('/api/events')
    sse.onopen = () => {
      connected.value = true
      lastError.value = null
    }
    sse.onerror = () => {
      connected.value = false
    }
    sse.onmessage = (msg) => {
      try {
        const data: ServerEvent = JSON.parse(msg.data)
        applyEvent(data)
      } catch {
        // ignore malformed lines
      }
    }
  }
  function disconnectSSE() {
    sse?.close()
    sse = null
    connected.value = false
  }

  return {
    totals,
    recentJobs,
    verifiers,
    topology,
    events,
    connected,
    lastError,
    passRatePct,
    loadSnapshot,
    connectSSE,
    disconnectSSE,
  }
})
