// API base URL — empty string means "same origin" (which is what the
// embedded build serves: SPA + API on one Go binary, and what `vite
// dev` does via its /api proxy to localhost:3000).
//
// On GitHub Pages the SPA is hosted on a different origin (github.io)
// while the Go server runs on EC2. The deploy-pages workflow bakes the
// EC2 URL into the build via VITE_API_BASE — see .github/workflows/
// deploy-pages.yml.

import { ref } from 'vue'

const fromBuild = (import.meta.env.VITE_API_BASE as string | undefined) ?? ''
const trimmed = fromBuild.replace(/\/+$/, '')

// Kept as a ref so existing call sites (apiBase.value) compile, even
// though the value is now build-time and never mutates at runtime.
export const apiBase = ref<string>(trimmed)

export function apiUrl(path: string): string {
  return apiBase.value + path
}
