<script setup lang="ts">
import { ref } from 'vue'
import { connectedWallet, connectWallet, disconnectWallet, hasInjectedWallet } from '@/wallet'

// Truncated display: 0xabcd…1234
function short(addr: string): string {
  return addr.slice(0, 6) + '…' + addr.slice(-4)
}

const error = ref<string | null>(null)

async function onClick() {
  error.value = null
  if (connectedWallet.value) {
    disconnectWallet()
    return
  }
  try {
    await connectWallet()
  } catch (e) {
    error.value = (e as Error).message
    setTimeout(() => { error.value = null }, 6000)
  }
}
</script>

<template>
  <div class="relative">
    <button
      v-if="hasInjectedWallet()"
      @click="onClick"
      class="pill bg-bg-subtle border border-bg-border hover:border-accent-chain text-ink-2"
      :title="connectedWallet ? `Connected: ${connectedWallet} — click to disconnect` : 'Connect a wallet to authorize Post Job'"
    >
      <span v-if="connectedWallet" class="flex items-center gap-1.5">
        <span class="inline-block w-1.5 h-1.5 rounded-full bg-accent-pass"></span>
        {{ short(connectedWallet) }}
      </span>
      <span v-else>Connect wallet</span>
    </button>
    <span
      v-else
      class="pill bg-bg-subtle border border-bg-border text-ink-2 opacity-60"
      title="Install MetaMask (or another EIP-1193 wallet) to authorize Post Job"
    >
      No wallet
    </span>
    <div
      v-if="error"
      class="absolute right-0 mt-1 panel border-accent-fail/40 text-accent-fail px-2 py-1 text-xs whitespace-nowrap"
    >
      {{ error }}
    </div>
  </div>
</template>
