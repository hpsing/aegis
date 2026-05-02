<script setup lang="ts">
import { useSwarmStore } from '@/stores/swarm'

const store = useSwarmStore()
</script>

<template>
  <div class="max-w-5xl mx-auto">
    <h1 class="panel-title px-0 border-0">System health</h1>
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mt-3">
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">Jobs verified</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.jobsVerified }}</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">PASS rate</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.passRatePct }}%</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">KH calls</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.totalKHInvocations }}</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">0G uploads</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.ogUploadsTotal }}</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">Mean commit→settle</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.meanCommitToSettleSec }}s</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">AXL msg/min</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.axlMessageRatePerMin }}</div>
      </div>
      <div class="panel p-3">
        <div class="text-ink-3 text-xs uppercase tracking-wider">Swarm online</div>
        <div class="text-2xl text-ink-0 mt-1">{{ store.totals.swarmOnline }}/{{ store.totals.swarmTotal }}</div>
      </div>
    </div>
  </div>
</template>
