// Minimal wallet-connect using the injected EIP-1193 provider
// (window.ethereum). We don't pull in ethers/viem because for our needs
// — eth_requestAccounts + personal_sign — the raw RPC is two lines.
//
// The connected address is persisted to localStorage so reloads don't
// force a reconnect prompt. We do NOT persist any signed material —
// every POST to a gated endpoint signs a fresh auth message.

import { ref } from 'vue'

const STORAGE_KEY = 'aegis.wallet'

declare global {
  interface Window {
    ethereum?: {
      request: (args: { method: string; params?: unknown[] }) => Promise<unknown>
      on?: (event: string, handler: (...args: unknown[]) => void) => void
      removeListener?: (event: string, handler: (...args: unknown[]) => void) => void
    }
  }
}

function loadInitial(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY)
  } catch {
    return null
  }
}

export const connectedWallet = ref<string | null>(loadInitial())

export function hasInjectedWallet(): boolean {
  return typeof window !== 'undefined' && !!window.ethereum
}

export async function connectWallet(): Promise<string> {
  if (!window.ethereum) {
    throw new Error('No injected wallet found — install MetaMask (or another EIP-1193 wallet)')
  }
  const accounts = (await window.ethereum.request({ method: 'eth_requestAccounts' })) as string[]
  if (!accounts || accounts.length === 0) {
    throw new Error('No accounts returned by wallet')
  }
  const addr = accounts[0].toLowerCase()
  connectedWallet.value = addr
  try {
    localStorage.setItem(STORAGE_KEY, addr)
  } catch {
    // ignore — non-fatal
  }
  return addr
}

export function disconnectWallet() {
  connectedWallet.value = null
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // ignore
  }
}

// signAuthMessage produces the headers the server's walletAuthMW
// expects for a gated POST. Message format MUST match auth.go:
//   "Aegis API access\nwallet: <wallet>\nts: <ts>"
export async function signAuthMessage(): Promise<{
  wallet: string
  sig: string
  ts: string
}> {
  if (!window.ethereum) {
    throw new Error('No wallet available to sign')
  }
  const wallet = connectedWallet.value
  if (!wallet) {
    throw new Error('Connect a wallet first')
  }
  const ts = Math.floor(Date.now() / 1000)
  const msg = `Aegis API access\nwallet: ${wallet}\nts: ${ts}`
  // personal_sign signs UTF-8 text, with the EIP-191 prefix added
  // automatically by the wallet. Server-side recovery uses the same
  // prefix (see auth.go::verifyWalletAuth).
  const sig = (await window.ethereum.request({
    method: 'personal_sign',
    params: [msg, wallet],
  })) as string
  return { wallet, sig, ts: String(ts) }
}
