<script setup lang="ts">
import { ref } from 'vue'
import { apiBase, setApiBase, probeApi } from '@/api/base'
import { useSwarmStore } from '@/stores/swarm'

const emit = defineEmits<{ (e: 'close'): void }>()
const store = useSwarmStore()

const draft = ref<string>(apiBase.value)
const probing = ref(false)
const probeResult = ref<{ ok: boolean; msg: string } | null>(null)

async function onProbe() {
  if (probing.value) return
  probing.value = true
  probeResult.value = null
  const target = draft.value.trim()
  const res = await probeApi(target)
  probeResult.value = res.ok
    ? { ok: true, msg: 'API reachable; response shape matches.' }
    : { ok: false, msg: res.error }
  probing.value = false
}

async function onSave() {
  setApiBase(draft.value)
  await store.loadSnapshot()
  store.reconnectSSE()
  emit('close')
}

function onClear() {
  draft.value = ''
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" @click.self="emit('close')">
    <div class="panel max-w-xl w-full bg-bg-panel">
      <div class="panel-title flex items-center justify-between">
        <span>API endpoint</span>
        <button class="text-ink-3 hover:text-ink-0" @click="emit('close')">✕</button>
      </div>

      <div class="p-4 space-y-3 text-xs">
        <p class="text-ink-2">
          The SPA fetches data + opens the SSE stream from this base URL.
          Leave empty to use the same origin as the page (works when the SPA
          is served by the embedded Go binary).
        </p>
        <p class="text-ink-3">
          For a public demo: run the Go server locally, expose it with
          <code class="text-ink-1">cloudflared tunnel --url http://localhost:3000</code>,
          paste the resulting <code class="text-ink-1">https://&lt;words&gt;.trycloudflare.com</code>
          URL here.
        </p>

        <div class="space-y-1">
          <label class="text-ink-3 uppercase tracking-wider">base url</label>
          <div class="flex gap-2">
            <input
              v-model="draft"
              type="url"
              placeholder="https://your-tunnel.trycloudflare.com"
              class="flex-1 bg-bg-subtle border border-bg-border rounded px-2 py-1.5 text-ink-0 font-mono text-xs focus:border-accent-chain outline-none"
            />
            <button
              class="px-3 py-1.5 rounded bg-bg-subtle hover:bg-bg-border text-ink-2"
              @click="onClear"
              title="Clear and use same origin"
            >Clear</button>
          </div>
          <div class="text-ink-3">
            Currently active: <code class="text-ink-1">{{ apiBase || '(same origin)' }}</code>
          </div>
        </div>

        <div v-if="probeResult" :class="probeResult.ok ? 'text-accent-pass' : 'text-accent-fail'">
          {{ probeResult.ok ? '✓' : '✗' }} {{ probeResult.msg }}
        </div>

        <div class="flex justify-between items-center pt-2 border-t border-bg-border">
          <button
            class="px-3 py-1.5 rounded bg-bg-subtle hover:bg-bg-border text-ink-2 disabled:opacity-40"
            :disabled="probing"
            @click="onProbe"
          >
            {{ probing ? 'Testing…' : 'Test connection' }}
          </button>
          <div class="flex gap-2">
            <button class="px-3 py-1.5 rounded bg-bg-subtle hover:bg-bg-border" @click="emit('close')">Cancel</button>
            <button
              class="px-3 py-1.5 rounded bg-accent-chain/30 text-accent-chain hover:bg-accent-chain/40"
              @click="onSave"
            >
              Save &amp; reconnect
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
