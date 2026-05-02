// Friendly labels for well-known addresses we encounter in the demo.
// The map covers (a) Base mainnet tokens used in the synthetic swap
// spec, (b) 0G Galileo deployment addresses (USDC, contracts), and
// (c) per-org KeeperHub wallets we ourselves operate. Keys MUST be
// lowercased — caller normalizes before lookup.

export interface KnownAddr {
  symbol: string
  name: string
  chain?: string // e.g. 'Base', '0G Galileo'
}

const REGISTRY: Record<string, KnownAddr> = {
  // Base — what the synthetic swap claims to have done
  '0x036cbd53842c5426634e7929541ec2318f3dcf7e': { symbol: 'USDC', name: 'USD Coin', chain: 'Base' },
  '0x4200000000000000000000000000000000000006': { symbol: 'WETH', name: 'Wrapped Ether', chain: 'Base' },

  // 0G Galileo deployment — see contracts/deployments/og-galileo.json
  '0xe9da98eb0af68cc48be7f71c29a7bc5ba7fb45eb': { symbol: 'mUSDC', name: 'Mock USDC', chain: '0G Galileo' },
  '0xa1a6327a64502a66565a6a11f6754cd9b867d018': { symbol: 'VerifierINFT', name: 'ERC-7857 Verifier iNFT', chain: '0G Galileo' },
  '0x50ce23ae35bbe43ffad0b36fd3f567560b8efb18': { symbol: 'VerifierRegistry', name: 'Verifier Registry', chain: '0G Galileo' },
  '0xa89833fbd1844763cc77c0a3afae32697a2f990f': { symbol: 'AegisContract', name: 'Aegis (escrow + commit-reveal)', chain: '0G Galileo' },
  '0x7c9dca2fb05cfc732794cee846888f3b1d7fe06a': { symbol: 'Treasury', name: 'Treasury / client wallet', chain: '0G Galileo' },

  // Per-org KH wallets — same as VerifierRegistry's registered EOAs
  '0x1bd14313fe46bc9ff2a2042898009e0c42abaa4a': { symbol: 'verifier-1', name: 'Verifier 1 wallet (KH org-1)', chain: '0G Galileo' },
  '0x6125c17e4893d7d8a04a1f962794f2da80803e31': { symbol: 'verifier-2', name: 'Verifier 2 wallet (KH org-2)', chain: '0G Galileo' },
  '0x8f3e641e80954c2a7dda9ca13d1197d049eee37f': { symbol: 'verifier-3', name: 'Verifier 3 wallet (KH org-3)', chain: '0G Galileo' },

  // 0G Storage flow contract (where ProofBundle uploads land on chain)
  '0x22e03a6a89b950f1c82ec5e74f8eca321a105296': { symbol: 'OGStorageFlow', name: '0G Storage flow contract', chain: '0G Galileo' },
}

export function knownAddr(addr?: string): KnownAddr | undefined {
  if (!addr) return undefined
  return REGISTRY[addr.toLowerCase()]
}

export function addrLabel(addr?: string): string {
  const k = knownAddr(addr)
  if (!k) return ''
  return k.symbol
}
