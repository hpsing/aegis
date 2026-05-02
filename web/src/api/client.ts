import type {
  JobDetail,
  JobSummary,
  PostJobPreview,
  PostJobResult,
  StateSnapshot,
  TopologyState,
  VerifierInfo,
} from './types'
import { apiUrl } from './base'
import { connectedWallet, signAuthMessage } from '@/wallet'

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(apiUrl(path), { headers: { Accept: 'application/json' } })
  if (!res.ok) {
    throw new Error(`${path}: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

// Build the wallet-auth headers for gated POST endpoints. If no wallet
// is connected, send nothing — the server only enforces when its
// AEGIS_WHITELIST is set, so local dev keeps working.
async function authHeaders(): Promise<Record<string, string>> {
  if (!connectedWallet.value) return {}
  const { wallet, sig, ts } = await signAuthMessage()
  return {
    'X-Aegis-Wallet': wallet,
    'X-Aegis-Sig': sig,
    'X-Aegis-Ts': ts,
  }
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
  postJob: async () => {
    const headers = await authHeaders()
    const r = await fetch(apiUrl('/api/post-job'), { method: 'POST', headers })
    if (!r.ok) {
      const body = await r.text().catch(() => '')
      throw new Error(body || `postJob: ${r.status}`)
    }
    return (await r.json()) as PostJobResult
  },
}
