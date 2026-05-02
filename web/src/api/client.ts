import type {
  JobDetail,
  JobSummary,
  PostJobPreview,
  PostJobResult,
  StateSnapshot,
  TopologyState,
  VerifierInfo,
} from './types'

// API base. Vite dev proxies /api → :3000 (Go backend). Prod serves
// from the same origin via the embedded SPA, so a relative base works
// in both modes.
const BASE = ''

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { headers: { Accept: 'application/json' } })
  if (!res.ok) {
    throw new Error(`${path}: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  state: () => getJSON<StateSnapshot>('/api/state'),
  jobs: (limit = 20, offset = 0) =>
    getJSON<{ items: JobSummary[]; total: number }>(`/api/jobs?limit=${limit}&offset=${offset}`),
  job: (id: string) => getJSON<JobDetail>(`/api/jobs/${encodeURIComponent(id)}`),
  verifiers: () => getJSON<VerifierInfo[]>('/api/verifiers'),
  topology: () => getJSON<TopologyState>('/api/topology'),
  proofbundle: (root: string) => getJSON<unknown>(`/api/proofbundle/${encodeURIComponent(root)}`),
  postJobPreview: () => getJSON<PostJobPreview>('/api/post-job/preview'),
  postJob: () =>
    fetch(`${BASE}/api/post-job`, { method: 'POST' }).then(async (r) => {
      if (!r.ok) {
        const body = await r.text().catch(() => '')
        throw new Error(body || `postJob: ${r.status}`)
      }
      return r.json() as Promise<PostJobResult>
    }),
}
