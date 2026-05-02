// Single source of truth for explorer / gateway URLs the UI links to.
// All hardcoded base URLs live here so we can override at build time
// without grepping across components.

export const EXPLORER = 'https://chainscan-galileo.0g.ai'
export const STORAGE_GATEWAY = 'https://storagescan-galileo.0g.ai'
export const KEEPERHUB_DASHBOARD = 'https://app.keeperhub.com'

export function txLink(hash?: string): string | undefined {
  if (!hash || !hash.startsWith('0x') || hash.length < 10) return undefined
  return `${EXPLORER}/tx/${hash}`
}

export function addrLink(addr?: string): string | undefined {
  if (!addr || !addr.startsWith('0x') || addr.length < 10) return undefined
  return `${EXPLORER}/address/${addr}`
}

// 0G Storage roots are content-addressed merkle hashes. The storage
// scan exposes them via /tx/<root>; if the path schema changes,
// override here.
export function storageLink(root?: string): string | undefined {
  if (!root) return undefined
  const r = root.startsWith('0x') ? root : '0x' + root
  return `${STORAGE_GATEWAY}/tx/${r}`
}

export function khOrgLink(_orgId?: string | number): string {
  // Per-org dashboard URLs are private; we link to the workspace root
  // and the user's authenticated session decides which org to show.
  return KEEPERHUB_DASHBOARD + '/'
}

export function shortHash(s?: string, head = 6, tail = 4): string {
  if (!s) return '—'
  if (s.length <= head + tail + 2) return s
  return `${s.slice(0, head)}…${s.slice(-tail)}`
}

export function fmtMUSDC(s?: string): string {
  if (!s) return '0'
  // mUSDC has 6 decimals; chop them off for display.
  if (s.length <= 6) return '0.' + s.padStart(6, '0')
  const whole = s.slice(0, s.length - 6)
  return whole
}
