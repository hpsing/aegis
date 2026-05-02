// Types mirror the Go server's response shapes. Kept stringly-typed
// where the wire is decimal/hex strings (chain ids, big ints) so the
// UI never coerces precision.

export type JobStatus = 'None' | 'Posted' | 'ClaimSubmitted' | 'Settled' | 'Cancelled'

export interface JobSummary {
  id: string
  status: JobStatus
  client: string
  executor: string
  postedBlock: number
  postedAt: number
  settledAt?: number
  finalVerdict?: 'PASS' | 'FAIL'
  forVotes?: number
  totalReveals?: number
}

export interface JobDetail extends JobSummary {
  specHash: string
  txHash: string
  reportedOutcomeHash: string
  claimDeadline: number
  commitDeadline: number
  revealDeadline: number
  postJobTx?: string
  submitClaimTx?: string
  settleTx?: string
  commits: VerifierVote[]
  reveals: VerifierVote[]
  proofBundles: ProofBundleRef[]
}

export interface VerifierVote {
  verifier: string
  txHash: string
  blockNumber?: number
  verdict?: 'PASS' | 'FAIL'
  commitHash?: string
}

export interface ProofBundleRef {
  verifier: string
  root: string
  voteRecordRoot?: string
}

export interface VerifierInfo {
  address: string
  iNFTId?: string
  agentCardURI?: string
  stake: string
  votesTotal: number
  votesCorrect: number
  accuracyBps: number
  active: boolean
  axlPeerId?: string
  role: string
}

export interface PostJobResult {
  jobId?: string
  txHash?: string
}

export interface PreviewSpec {
  action: string
  chainId: number
  intent: string
  tokenIn: string
  tokenOut: string
  amountIn: string
  maxSlippageBps: number
}

export interface PreviewPeer {
  role: string
  peerId: string
}

export interface PostJobPreview {
  spec: PreviewSpec
  specHash: string
  client: string
  executor: string
  reimbursement: string
  fee: string
  bounty: string
  totalEscrow: string
  verifierPeers: PreviewPeer[]
  ready: boolean
  blockers: string[]
}

export interface SystemTotals {
  jobsVerified: number
  passRateBps: number
  totalKHInvocations: number
  meanCommitToSettleSec: number
  axlMessageRatePerMin: number
  ogUploadsTotal: number
  swarmOnline: number
  swarmTotal: number
}

export interface TopologyState {
  freshnessMs: number
  nodes: TopologyNode[]
  edges: TopologyEdge[]
}

export interface TopologyNode {
  role: 'pub' | 'v1' | 'v2' | 'v3'
  peerId: string
  apiUrl: string
  ipv6: string
  online: boolean
}

export interface TopologyEdge {
  from: string
  to: string
  treeParent: boolean
}

export interface StateSnapshot {
  totals: SystemTotals
  recentJobs: JobSummary[]
  verifiers: VerifierInfo[]
  topology: TopologyState
  serverTime: number
}

// SSE events. Discriminated by `kind`. The server may add new kinds; the
// UI must ignore unknown kinds rather than crash.
export type ServerEvent =
  | { kind: 'chain.JobPosted'; jobId: string; client: string; executor: string; tx: string; ts: number }
  | { kind: 'chain.ClaimSubmitted'; jobId: string; tx: string; reportedTxHash: string; ts: number }
  | { kind: 'chain.VoteCommitted'; jobId: string; verifier: string; tx: string; ts: number }
  | { kind: 'chain.VoteRevealed'; jobId: string; verifier: string; verdict: 'PASS' | 'FAIL'; tx: string; ts: number }
  | { kind: 'chain.JobSettled'; jobId: string; finalVerdict: 'PASS' | 'FAIL'; forVotes: number; totalReveals: number; tx: string; ts: number }
  | { kind: 'axl.send'; from: string; to: string; envelope: 'spec_publish' | 'vote_commit' | 'vote_reveal'; ts: number }
  | { kind: 'kh.workflow'; verifier: string; purpose: 'commit' | 'reveal' | 'settle'; jobId: string; ts: number }
  | { kind: 'og.upload'; verifier: string; jobId: string; root: string; objectType: 'ProofBundle' | 'VoteRecord'; ts: number }
  | { kind: 'topology.update'; topology: TopologyState; ts: number }
  | { kind: 'totals.update'; totals: SystemTotals; ts: number }
  | { kind: 'postjob.step'; stage: string; msg: string; jobId?: string; tx?: string; ts: number; [k: string]: unknown }
  | { kind: 'axl.spec_received'; role: string; specHash: string; ts: number }
  | { kind: 'hello'; ts: number }
