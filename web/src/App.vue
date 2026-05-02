<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useSwarmStore } from '@/stores/swarm'
import { api } from '@/api/client'
import PostJobModal from '@/components/PostJobModal.vue'
import AboutPanel from '@/components/AboutPanel.vue'
import WalletButton from '@/components/WalletButton.vue'
import { connectedWallet } from '@/wallet'

const store = useSwarmStore()
const router = useRouter()

onMounted(async () => {
  await store.loadSnapshot()
  store.connectSSE()
})
onUnmounted(() => store.disconnectSSE())

// Header context: surface the most recent (or in-flight) job inline so
// the operator doesn't have to scroll to the timeline to know what's
// happening.
const activeJob = computed(() => store.recentJobs[0])
const activeJobLabel = computed(() => {
  const j = activeJob.value
  if (!j) return ''
  if (j.status === 'Settled') {
    return `#${j.id} ${j.status} · ${j.finalVerdict ?? ''}`.trim()
  }
  if (j.status === 'ClaimSubmitted') {
    return `#${j.id} ${j.status} · ${j.totalReveals ?? 0}/3 reveals`
  }
  return `#${j.id} ${j.status}`
})
const activeJobClass = computed(() => {
  const v = activeJob.value?.finalVerdict
  if (v === 'PASS') return 'text-accent-pass'
  if (v === 'FAIL') return 'text-accent-fail'
  return 'text-ink-1'
})

interface Toast {
  kind: 'info' | 'ok' | 'err'
  text: string
}
const toast = ref<Toast | null>(null)
const showModal = ref(false)
const showAbout = ref(false)

function openModal() {
  showModal.value = true
}

// onConfirm runs the post-job API call. Triggered by the modal's
// 'confirm' event; the modal closes itself first so the operator can
// watch the timeline + topology react in real time. The toast banner
// shows in-flight + final status so they don't lose track.
async function onConfirm() {
  showModal.value = false
  toast.value = { kind: 'info', text: 'Posting job — see timeline for progress…' }
  try {
    const res = await api.postJob()
    if (res.jobId) {
      toast.value = { kind: 'ok', text: `Job #${res.jobId} posted` }
      router.push(`/jobs/${res.jobId}`)
    } else {
      toast.value = { kind: 'ok', text: 'postJob landed (jobId pending)' }
    }
  } catch (e) {
    toast.value = { kind: 'err', text: (e as Error).message }
  } finally {
    setTimeout(() => { toast.value = null }, 8000)
  }
}
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="panel border-b border-bg-border flex items-center justify-between px-4 py-2 sticky top-0 z-10 bg-bg-panel">
      <div class="flex items-center gap-6">
        <div class="font-bold text-ink-0">Aegis</div>
        <nav class="flex gap-1 text-sm">
          <RouterLink to="/" class="nav-link" :class="$route.name === 'live' ? 'nav-link-active' : ''">Live</RouterLink>
          <RouterLink to="/verifiers" class="nav-link" :class="$route.name === 'verifiers' ? 'nav-link-active' : ''">Verifiers</RouterLink>
          <RouterLink to="/system" class="nav-link" :class="$route.name === 'system' ? 'nav-link-active' : ''">System</RouterLink>
          <button class="nav-link" title="What is this demo?" @click="showAbout = true">About</button>
        </nav>
      </div>
      <div class="flex items-center gap-4 text-xs">
        <div class="flex items-center gap-1.5">
          <span class="inline-block w-2 h-2 rounded-full" :class="store.connected ? 'bg-accent-pass' : 'bg-accent-fail'"></span>
          <span class="text-ink-2">{{ store.connected ? 'live' : 'offline' }}</span>
        </div>
        <div class="text-ink-2">{{ store.totals.jobsVerified }} jobs</div>
        <div class="text-ink-2">{{ store.passRatePct }}% PASS</div>
        <div class="text-ink-2">{{ store.totals.swarmOnline }}/{{ store.totals.swarmTotal }} online</div>
        <RouterLink
          v-if="activeJob"
          :to="`/jobs/${activeJob.id}`"
          class="pill bg-bg-subtle border border-bg-border hover:border-accent-chain"
          :class="activeJobClass"
          :title="`Active: ${activeJobLabel}`"
        >
          {{ activeJobLabel }}
        </RouterLink>
        <WalletButton />
        <button
          v-if="connectedWallet"
          @click="openModal"
          :disabled="store.totals.swarmOnline < 3"
          class="px-3 py-1 rounded bg-accent-chain/20 text-accent-chain hover:bg-accent-chain/30 disabled:opacity-40 disabled:cursor-not-allowed transition"
          :title="store.totals.swarmOnline < 3 ? 'Waiting for all 3 verifiers online' : 'Review and post a fresh job (will sign with connected wallet)'"
        >
          Post Job…
        </button>
        <span
          v-else
          class="px-3 py-1 rounded text-ink-2 text-xs italic opacity-70"
          title="Connect a wallet from the Connect button — only whitelisted wallets can post jobs"
        >
          Connect wallet to post a job
        </span>
      </div>
    </header>

    <main class="flex-1 p-4">
      <RouterView />
    </main>

    <PostJobModal v-if="showModal" @close="showModal = false" @confirm="onConfirm" />
    <AboutPanel v-if="showAbout" @close="showAbout = false" />

    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 translate-y-1"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="toast"
        :class="[
          'fixed bottom-3 right-3 panel px-3 py-2 text-xs',
          toast.kind === 'ok' ? 'border-accent-pass/40 text-accent-pass' :
          toast.kind === 'err' ? 'border-accent-fail/40 text-accent-fail' :
          'text-ink-1'
        ]"
      >
        {{ toast.text }}
      </div>
    </transition>

    <div v-if="store.lastError && !toast" class="fixed bottom-3 right-3 panel border-accent-fail/40 text-accent-fail px-3 py-1.5 text-xs">
      {{ store.lastError }}
    </div>
  </div>
</template>
