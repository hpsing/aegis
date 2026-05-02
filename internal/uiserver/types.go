package uiserver

import "encoding/json"

// JSON shapes mirror web/src/api/types.ts. Keep field names in sync —
// the front end is stringly-typed for big-ints and addresses.

type JobSummary struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Client        string `json:"client"`
	Executor      string `json:"executor"`
	PostedBlock   uint64 `json:"postedBlock"`
	PostedAt      int64  `json:"postedAt"` // unix ms
	SettledAt     int64  `json:"settledAt,omitempty"`
	FinalVerdict  string `json:"finalVerdict,omitempty"` // PASS | FAIL
	ForVotes      int    `json:"forVotes,omitempty"`
	TotalReveals  int    `json:"totalReveals,omitempty"`
}

type JobDetail struct {
	JobSummary
	SpecHash             string             `json:"specHash"`
	TxHash               string             `json:"txHash"`
	ReportedOutcomeHash  string             `json:"reportedOutcomeHash"`
	ClaimDeadline        int64              `json:"claimDeadline"`
	CommitDeadline       int64              `json:"commitDeadline"`
	RevealDeadline       int64              `json:"revealDeadline"`
	PostJobTx            string             `json:"postJobTx,omitempty"`
	SubmitClaimTx        string             `json:"submitClaimTx,omitempty"`
	SettleTx             string             `json:"settleTx,omitempty"`
	Commits              []VerifierVote     `json:"commits"`
	Reveals              []VerifierVote     `json:"reveals"`
	ProofBundles         []ProofBundleRef   `json:"proofBundles"`
}

type VerifierVote struct {
	Verifier    string `json:"verifier"`
	TxHash      string `json:"txHash"`
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	Verdict     string `json:"verdict,omitempty"` // PASS|FAIL on reveals
	CommitHash  string `json:"commitHash,omitempty"`
}

type ProofBundleRef struct {
	Verifier        string `json:"verifier"`
	Root            string `json:"root"`
	VoteRecordRoot  string `json:"voteRecordRoot,omitempty"`
}

type VerifierInfo struct {
	Address       string `json:"address"`
	INFTID        string `json:"iNFTId,omitempty"`
	AgentCardURI  string `json:"agentCardURI,omitempty"`
	Stake         string `json:"stake"`
	VotesTotal    int    `json:"votesTotal"`
	VotesCorrect  int    `json:"votesCorrect"`
	AccuracyBps   int    `json:"accuracyBps"`
	Active        bool   `json:"active"`
	AxlPeerID     string `json:"axlPeerId,omitempty"`
	Role          string `json:"role"`
}

type SystemTotals struct {
	JobsVerified           int `json:"jobsVerified"`
	PassRateBps            int `json:"passRateBps"`
	TotalKHInvocations     int `json:"totalKHInvocations"`
	MeanCommitToSettleSec  int `json:"meanCommitToSettleSec"`
	AxlMessageRatePerMin   int `json:"axlMessageRatePerMin"`
	OGUploadsTotal         int `json:"ogUploadsTotal"`
	SwarmOnline            int `json:"swarmOnline"`
	SwarmTotal             int `json:"swarmTotal"`
}

type TopologyState struct {
	FreshnessMs int64           `json:"freshnessMs"`
	Nodes       []TopologyNode  `json:"nodes"`
	Edges       []TopologyEdge  `json:"edges"`
}

type TopologyNode struct {
	Role   string `json:"role"`
	PeerID string `json:"peerId"`
	APIURL string `json:"apiUrl"`
	IPv6   string `json:"ipv6"`
	Online bool   `json:"online"`
}

type TopologyEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	TreeParent bool   `json:"treeParent"`
}

type StateSnapshot struct {
	Totals     SystemTotals    `json:"totals"`
	RecentJobs []JobSummary    `json:"recentJobs"`
	Verifiers  []VerifierInfo  `json:"verifiers"`
	Topology   TopologyState   `json:"topology"`
	ServerTime int64           `json:"serverTime"`
}

type PostJobResult struct {
	JobID string `json:"jobId"`
	TxHash string `json:"txHash"`
}

// PostJobPreview is what the server would post if /api/post-job were
// called RIGHT NOW. The UI shows this in a confirmation modal so the
// operator can see what the job is about. Currently read-only — the
// server enforces hardcoded amounts; future versions can take an
// edited spec back via POST body.
type PostJobPreview struct {
	Spec            PreviewSpec   `json:"spec"`
	SpecHash        string        `json:"specHash"`
	Client          string        `json:"client"`
	Executor        string        `json:"executor"`
	Reimbursement   string        `json:"reimbursement"`     // mUSDC, 6 decimals
	Fee             string        `json:"fee"`
	Bounty          string        `json:"bounty"`
	TotalEscrow     string        `json:"totalEscrow"`
	VerifierPeers   []PreviewPeer `json:"verifierPeers"`
	Ready           bool          `json:"ready"`
	Blockers        []string      `json:"blockers"`
}

type PreviewSpec struct {
	Action          string `json:"action"`
	ChainID         int64  `json:"chainId"`
	Intent          string `json:"intent"`
	TokenIn         string `json:"tokenIn"`
	TokenOut        string `json:"tokenOut"`
	AmountIn        string `json:"amountIn"`
	MaxSlippageBps  int    `json:"maxSlippageBps"`
}

type PreviewPeer struct {
	Role   string `json:"role"`
	PeerID string `json:"peerId"`
}

// ServerEvent is the SSE payload. The front end parses by `kind`.
// The server's job is to serialize to JSON and push via Hub.Broadcast.
type ServerEvent struct {
	Kind string          `json:"kind"`
	Ts   int64           `json:"ts"`
	Data json.RawMessage `json:"-"` // set by helper; flattened on marshal
}

// EventEnvelope is a flat shape suitable for JSON serialization. The
// server constructs this directly when emitting events.
type EventEnvelope = map[string]any
