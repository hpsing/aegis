<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  value: string
  href?: string
  // mono = render the value in monospace + word-break (good for hex)
  mono?: boolean
  // size = visual size; 'sm' tightens to text-[10px]
  size?: 'sm' | 'base'
  // label = optional friendly label rendered before the value (e.g. "USDC")
  label?: string
}>(), {
  mono: true,
  size: 'base',
})

const copied = ref(false)
const tip = ref(false)

async function copy(e: MouseEvent) {
  e.preventDefault()
  e.stopPropagation()
  try {
    await navigator.clipboard.writeText(props.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1200)
  } catch {
    // Fallback: select text in a temporary input.
    const ta = document.createElement('textarea')
    ta.value = props.value
    document.body.appendChild(ta)
    ta.select()
    try { document.execCommand('copy') } catch {}
    document.body.removeChild(ta)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1200)
  }
}
</script>

<template>
  <span
    class="inline-flex items-center gap-1 group"
    @mouseenter="tip = true"
    @mouseleave="tip = false"
  >
    <span v-if="label" class="text-ink-2">{{ label }}</span>
    <a
      v-if="href"
      :href="href"
      target="_blank"
      rel="noopener"
      :class="[
        'hover:text-accent-chain break-all',
        mono ? 'font-mono' : '',
        size === 'sm' ? 'text-[10px]' : '',
      ]"
    >{{ value }}</a>
    <span
      v-else
      :class="[
        'break-all',
        mono ? 'font-mono' : '',
        size === 'sm' ? 'text-[10px]' : '',
      ]"
    >{{ value }}</span>
    <button
      type="button"
      :title="copied ? 'Copied!' : 'Copy'"
      class="opacity-0 group-hover:opacity-100 transition text-ink-3 hover:text-ink-0"
      @click="copy"
    >
      <span v-if="copied" class="text-accent-pass text-[10px]">✓</span>
      <span v-else class="text-[10px]">⧉</span>
    </button>
  </span>
</template>
