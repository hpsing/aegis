// API base URL — empty string means "same origin" (which is what the
// embedded build serves: SPA + API on one Go binary).
//
// On GitHub Pages the SPA is hosted on a different origin (github.io)
// while the Go server runs locally behind a cloudflared tunnel. The
// operator pastes the tunnel URL into the SPA via the header pill;
// we persist it to localStorage so reloads remember.
//
// Build-time `VITE_API_BASE` (set by the GH Pages CI) provides the
// default if the user hasn't pasted one yet.

import { ref } from 'vue'

const STORAGE_KEY = 'aegis.apiBase'

function trimTrailingSlash(s: string): string {
  return s.replace(/\/+$/, '')
}

function loadInitial(): string {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored !== null) return trimTrailingSlash(stored)
  } catch {
    // localStorage unavailable (e.g. private mode); fall through
  }
  const fromBuild = import.meta.env.VITE_API_BASE as string | undefined
  return trimTrailingSlash(fromBuild ?? '')
}

// Reactive ref so components rebind when the user changes the base.
export const apiBase = ref<string>(loadInitial())

export function setApiBase(value: string) {
  const v = trimTrailingSlash(value.trim())
  apiBase.value = v
  try {
    localStorage.setItem(STORAGE_KEY, v)
  } catch {
    // ignore — non-fatal
  }
}

// Build a full URL for a given API path (already starts with /).
export function apiUrl(path: string): string {
  return apiBase.value + path
}

// Health probe: GET /api/state with a short timeout. Returns true if
// the server responds with 200 + valid JSON containing `totals`.
export async function probeApi(base: string): Promise<{ ok: true } | { ok: false; error: string }> {
  const url = trimTrailingSlash(base) + '/api/state'
  const ctrl = new AbortController()
  const timer = setTimeout(() => ctrl.abort(), 5000)
  try {
    const res = await fetch(url, { signal: ctrl.signal, headers: { Accept: 'application/json' } })
    clearTimeout(timer)
    if (!res.ok) return { ok: false, error: `HTTP ${res.status}` }
    const body = await res.json().catch(() => null)
    if (!body || typeof body !== 'object' || !('totals' in body)) {
      return { ok: false, error: 'response missing `totals` — wrong server?' }
    }
    return { ok: true }
  } catch (e) {
    clearTimeout(timer)
    const msg = (e as Error).message
    if (msg.includes('Failed to fetch') || msg.includes('NetworkError')) {
      return { ok: false, error: 'network error — server unreachable, CORS blocked, or wrong URL' }
    }
    return { ok: false, error: msg }
  }
}
