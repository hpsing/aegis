<script setup lang="ts">
import { useSwarmStore } from '@/stores/swarm'
import { addrLink } from '@/api/links'
import Copyable from '@/components/Copyable.vue'

const store = useSwarmStore()

function pct(bps: number): string {
  return (bps / 100).toFixed(1) + '%'
}
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="panel-title px-0 border-0 mb-3">Verifier swarm registry</h1>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
      <div v-for="v in store.verifiers" :key="v.address" class="panel p-4 space-y-3 text-xs">
        <!-- Header: role + iNFT chip + active dot -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="inline-block w-2 h-2 rounded-full" :class="v.active ? 'bg-accent-pass' : 'bg-accent-fail'"></span>
            <span class="text-ink-0 font-bold text-sm">{{ v.role || 'verifier' }}</span>
          </div>
          <span v-if="v.iNFTId" class="pill bg-accent-og/15 text-accent-og">iNFT #{{ v.iNFTId }}</span>
          <span v-else class="pill bg-bg-subtle text-ink-3">no iNFT</span>
        </div>

        <!-- Wallet address -->
        <div>
          <div class="text-ink-3 uppercase tracking-wider mb-1">wallet</div>
          <Copyable :value="v.address" :href="addrLink(v.address)" size="sm" />
        </div>

        <!-- AXL peer -->
        <div v-if="v.axlPeerId">
          <div class="text-ink-3 uppercase tracking-wider mb-1">AXL peer id</div>
          <Copyable :value="v.axlPeerId" size="sm" />
        </div>

        <!-- Stats -->
        <dl class="grid grid-cols-2 gap-y-1 pt-2 border-t border-bg-border">
          <dt class="text-ink-3">stake</dt><dd>{{ v.stake }} mUSDC</dd>
          <dt class="text-ink-3">votes</dt><dd>{{ v.votesTotal }}</dd>
          <dt class="text-ink-3">accurate</dt>
          <dd :class="v.accuracyBps === 10000 ? 'text-accent-pass' : v.accuracyBps < 6700 ? 'text-accent-fail' : 'text-ink-1'">
            {{ v.votesCorrect }} ({{ pct(v.accuracyBps) }})
          </dd>
        </dl>

        <a v-if="v.agentCardURI" :href="v.agentCardURI" target="_blank" class="block pt-2 border-t border-bg-border text-accent-og">
          View Agent Card on 0G ↗
        </a>
      </div>
    </div>

    <p v-if="!store.verifiers.length" class="text-ink-3 text-xs mt-3">
      No registered verifiers. Run <code class="text-ink-1">scripts/register-verifiers.sh</code> first.
    </p>
  </div>
</template>
