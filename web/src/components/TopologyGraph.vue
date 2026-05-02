<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useSwarmStore } from '@/stores/swarm'
import type { ServerEvent, TopologyNode } from '@/api/types'

const store = useSwarmStore()

interface Coord { x: number; y: number }

const SIZE = 360
const CENTER: Coord = { x: SIZE / 2, y: SIZE / 2 }
const ORBIT = 130
const NODE_R = 28

// Lay nodes out radially with pub at center, verifiers around it.
const layout = computed<Record<string, Coord>>(() => {
  const out: Record<string, Coord> = {}
  const nodes = store.topology.nodes
  const verifiers = nodes.filter((n) => n.role !== 'pub')
  const pub = nodes.find((n) => n.role === 'pub')
  if (pub) out[pub.role] = CENTER
  verifiers.forEach((v, i) => {
    const angle = (-Math.PI / 2) + (i / Math.max(verifiers.length, 1)) * 2 * Math.PI
    out[v.role] = {
      x: CENTER.x + ORBIT * Math.cos(angle),
      y: CENTER.y + ORBIT * Math.sin(angle),
    }
  })
  return out
})

type EnvKind = 'spec_publish' | 'vote_commit' | 'vote_reveal'

interface Flow {
  id: number
  from: Coord
  to: Coord
  envelope: EnvKind
  // edgeKey is "from→to" for the static edge label location
  edgeLabelX: number
  edgeLabelY: number
}

let flowSeq = 0
const flows = ref<Flow[]>([])

function pulseColor(env: EnvKind): string {
  switch (env) {
    case 'spec_publish': return '#c084fc'
    case 'vote_commit': return '#fbbf24'
    case 'vote_reveal': return '#3ddc97'
  }
}

function spawnFlow(fromRole: string | undefined, toRole: string, env: EnvKind) {
  const from = fromRole ? layout.value[fromRole] : CENTER
  const to = layout.value[toRole]
  if (!from || !to) return
  const id = ++flowSeq
  flows.value.push({
    id,
    from,
    to,
    envelope: env,
    edgeLabelX: (from.x + to.x) / 2,
    edgeLabelY: (from.y + to.y) / 2,
  })
  // Keep the flow visible long enough for the eye to catch it.
  window.setTimeout(() => {
    flows.value = flows.value.filter((p) => p.id !== id)
  }, 1400)
}

function roleForAddr(addr: string): string {
  const a = (addr || '').toLowerCase()
  for (const v of store.verifiers) {
    if ((v.address || '').toLowerCase() === a && v.role && v.role !== 'verifier') {
      return v.role
    }
  }
  return ''
}

watch(
  () => store.events[0],
  (ev?: ServerEvent) => {
    if (!ev) return
    if (ev.kind === 'axl.send') {
      const send = ev as { kind: 'axl.send'; from?: string; to: string; envelope: EnvKind }
      spawnFlow(send.from || 'pub', send.to, send.envelope)
      return
    }
    if (ev.kind === 'chain.VoteCommitted') {
      const e = ev as { verifier: string }
      const role = roleForAddr(e.verifier)
      if (role) spawnFlow(role, 'pub', 'vote_commit')
      return
    }
    if (ev.kind === 'chain.VoteRevealed') {
      const e = ev as { verifier: string }
      const role = roleForAddr(e.verifier)
      if (role) spawnFlow(role, 'pub', 'vote_reveal')
      return
    }
  },
)

function tooltipFor(n: TopologyNode): string {
  const parts: string[] = [n.role.toUpperCase()]
  if (n.peerId) parts.push('peer ' + n.peerId)
  if (n.ipv6) parts.push('ipv6 ' + n.ipv6)
  parts.push(n.online ? 'online' : 'offline')
  return parts.join('\n')
}

// edgePath returns a path string for the line from (a) toward (b),
// stopping NODE_R-2 px short of (b) so the arrowhead doesn't overlap
// the destination circle.
function edgeEndpoint(a: Coord, b: Coord): Coord {
  const dx = b.x - a.x
  const dy = b.y - a.y
  const len = Math.hypot(dx, dy) || 1
  const back = NODE_R - 2
  return { x: b.x - (dx / len) * back, y: b.y - (dy / len) * back }
}
</script>

<template>
  <div class="relative w-full h-full flex items-center justify-center">
    <svg :viewBox="`0 0 ${SIZE} ${SIZE}`" class="w-full max-w-[420px] h-auto">
      <defs>
        <!-- Subtle arrowhead for static edges -->
        <marker id="arrow-static" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M0,0 L10,5 L0,10 z" fill="rgba(122,130,148,0.6)" />
        </marker>
        <!-- Bright arrowheads for active flows, one per envelope -->
        <marker id="arrow-spec" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M0,0 L10,5 L0,10 z" fill="#c084fc" />
        </marker>
        <marker id="arrow-commit" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M0,0 L10,5 L0,10 z" fill="#fbbf24" />
        </marker>
        <marker id="arrow-reveal" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M0,0 L10,5 L0,10 z" fill="#3ddc97" />
        </marker>
      </defs>

      <!-- Static edges (pub ↔ each verifier) — show the mesh shape -->
      <g class="text-bg-border" stroke="currentColor" stroke-width="1" stroke-dasharray="3 4">
        <line
          v-for="n in store.topology.nodes.filter((x) => x.role !== 'pub')"
          :key="`edge-${n.role}`"
          :x1="layout.pub?.x ?? CENTER.x"
          :y1="layout.pub?.y ?? CENTER.y"
          :x2="layout[n.role]?.x ?? CENTER.x"
          :y2="layout[n.role]?.y ?? CENTER.y"
        />
      </g>

      <!-- Active flows: animated drawing line + arrowhead + envelope label -->
      <g>
        <g v-for="f in flows" :key="f.id" :class="`flow flow-${f.envelope}`">
          <!-- Glow underlay for emphasis -->
          <line
            :x1="f.from.x" :y1="f.from.y"
            :x2="edgeEndpoint(f.from, f.to).x" :y2="edgeEndpoint(f.from, f.to).y"
            :stroke="pulseColor(f.envelope)"
            stroke-width="6"
            stroke-linecap="round"
            opacity="0.18"
          />
          <!-- The actual drawing line; stroke-dashoffset animates 0→length -->
          <line
            class="flow-line"
            :x1="f.from.x" :y1="f.from.y"
            :x2="edgeEndpoint(f.from, f.to).x" :y2="edgeEndpoint(f.from, f.to).y"
            :stroke="pulseColor(f.envelope)"
            stroke-width="2.5"
            stroke-linecap="round"
            :marker-end="`url(#arrow-${f.envelope === 'spec_publish' ? 'spec' : f.envelope === 'vote_commit' ? 'commit' : 'reveal'})`"
          />
          <!-- Envelope-type label near the midpoint -->
          <text
            :x="f.edgeLabelX"
            :y="f.edgeLabelY - 6"
            text-anchor="middle"
            class="flow-label pointer-events-none"
            :fill="pulseColor(f.envelope)"
          >{{ f.envelope }}</text>
        </g>
      </g>

      <!-- Nodes -->
      <g v-for="n in store.topology.nodes" :key="n.role" :transform="`translate(${layout[n.role]?.x ?? CENTER.x}, ${layout[n.role]?.y ?? CENTER.y})`">
        <title>{{ tooltipFor(n) }}</title>
        <circle
          :r="NODE_R"
          :class="[
            'fill-bg-subtle stroke-2 transition-colors',
            n.online ? 'stroke-accent-pass' : 'stroke-accent-fail',
            n.role === 'pub' ? 'stroke-accent-axl' : ''
          ]"
        />
        <text text-anchor="middle" dominant-baseline="central" class="fill-ink-0 text-[12px] font-bold pointer-events-none">
          {{ n.role }}
        </text>
        <text text-anchor="middle" :y="NODE_R + 14" class="fill-ink-3 text-[10px] pointer-events-none">
          {{ (n.peerId || '').slice(0, 8) }}{{ n.peerId ? '…' : '—' }}
        </text>
      </g>
    </svg>

    <!-- Legend -->
    <div class="absolute bottom-2 right-2 flex flex-col gap-1 text-[10px] text-ink-3 panel px-2 py-1.5 bg-bg-panel/90">
      <div class="flex items-center gap-1.5"><span class="inline-block w-3 h-0.5" style="background:#c084fc"></span> spec_publish (pub→V)</div>
      <div class="flex items-center gap-1.5"><span class="inline-block w-3 h-0.5" style="background:#fbbf24"></span> vote_commit (V→pub)</div>
      <div class="flex items-center gap-1.5"><span class="inline-block w-3 h-0.5" style="background:#3ddc97"></span> vote_reveal (V→pub)</div>
    </div>
  </div>
</template>

<style scoped>
/*
 * Each .flow's drawing line uses stroke-dasharray to "fill in" from
 * source to destination over ~1s, then fade. Length=300 is a safe
 * upper bound for our ORBIT=130 radius geometry; the dash pattern
 * exceeds the actual segment so dashoffset goes from 300→0 visibly.
 */
.flow-line {
  stroke-dasharray: 300;
  stroke-dashoffset: 300;
  animation: draw 0.9s cubic-bezier(0.2, 0.8, 0.4, 1) forwards,
             fadeout 0.4s 1.0s forwards;
  filter: drop-shadow(0 0 4px currentColor);
}
.flow-label {
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.5px;
  opacity: 0;
  animation: labelPop 1.4s ease-out forwards;
}

@keyframes draw {
  to   { stroke-dashoffset: 0; }
}
@keyframes fadeout {
  from { opacity: 1; }
  to   { opacity: 0; }
}
@keyframes labelPop {
  0%   { opacity: 0; }
  20%  { opacity: 1; }
  80%  { opacity: 1; }
  100% { opacity: 0; }
}
</style>
