package uiserver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hpsing/aegis/internal/aegis"
)

// LiveConfig is what cmd/ui passes in. RPC + contract addresses are
// required; AXL daemon URLs are optional (topology shows offline if
// missing). Verifier wallet addresses come from on-chain registry events.
type LiveConfig struct {
	RPC              string
	AegisAddr        common.Address
	RegistryAddr     common.Address
	USDCAddr         common.Address    // required for in-process post-job
	INFTAddr         common.Address    // optional, used for agentCardURI lookups
	AXLDaemons       map[string]string // role → http url, e.g. {"pub": "http://127.0.0.1:9002"}
	LogsRoot         string            // dir holding e2e-* run logs; tailer auto-discovers latest subdir
	BackfillBlocks   uint64            // how far back to scan on startup; 0 = head only
	PollInterval     time.Duration     // chain event poll cadence; 0 = 5s
	TopologyInterval time.Duration     // axl topology poll; 0 = 2s
	LogPollInterval  time.Duration     // verifier log tail poll; 0 = 1s

	// Optional: if all three are set, POST /api/post-job runs publisher
	// + executor in-process. If unset, the endpoint returns 503.
	TreasuryPKHex   string
	ExecutorPKHex   string
	ExecutorAddress common.Address
}

// LiveSource is the production DataSource. Holds an in-memory mirror
// that's updated by background goroutines (chain watcher, AXL topology
// poller). DataSource methods read from the mirror.
type LiveSource struct {
	cfg    LiveConfig
	hub    *Hub
	client *ethclient.Client
	aegis  *aegis.AegisContract
	reg    *aegis.VerifierRegistry

	mu          sync.RWMutex
	jobs        map[uint64]*jobState   // keyed by jobID
	verifiers   []VerifierInfo         // sorted by address
	topology    TopologyState
	totals      SystemTotals
	lastBlock   uint64

	tsMu     sync.Mutex
	blockTs  map[uint64]int64 // block number → unix ms; bounded by tsCache
	tsOrder  []uint64         // FIFO eviction

	// Windowed metrics. Updated by event handlers; read by recomputeTotals.
	commitTimes  map[string]int64 // "<jobID>/<verifierAddr>" → first commit ts (ms)
	settleLatNs  []int64          // commit→settle deltas per (job,verifier) in ns; capped
	axlSendsTs   []int64          // axl.send timestamps in ms; capped to last hour
}

const blockTsCacheMax = 5000

type jobState struct {
	summary  JobSummary
	detail   JobDetail
}

func NewLiveSource(ctx context.Context, hub *Hub, cfg LiveConfig) (*LiveSource, error) {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 5 * time.Second
	}
	if cfg.TopologyInterval == 0 {
		cfg.TopologyInterval = 2 * time.Second
	}
	client, err := ethclient.DialContext(ctx, cfg.RPC)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}
	a, err := aegis.NewAegisContract(cfg.AegisAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind aegis: %w", err)
	}
	r, err := aegis.NewVerifierRegistry(cfg.RegistryAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind registry: %w", err)
	}
	s := &LiveSource{
		cfg:       cfg,
		hub:       hub,
		client:    client,
		aegis:     a,
		reg:       r,
		jobs:      map[uint64]*jobState{},
		verifiers: []VerifierInfo{},
		topology: TopologyState{
			Nodes: []TopologyNode{},
			Edges: []TopologyEdge{},
		},
		totals:      SystemTotals{SwarmTotal: 3},
		blockTs:     map[uint64]int64{},
		commitTimes: map[string]int64{},
	}
	return s, nil
}

// blockTimestamp returns the unix-ms timestamp of `block`, fetching
// the header on cache miss. The chain RPC may be slow (HTTP only on
// 0G Galileo), so callers tolerate a fallback to wall-clock when the
// header fetch fails.
func (s *LiveSource) blockTimestamp(ctx context.Context, block uint64) int64 {
	s.tsMu.Lock()
	if v, ok := s.blockTs[block]; ok {
		s.tsMu.Unlock()
		return v
	}
	s.tsMu.Unlock()

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	h, err := s.client.HeaderByNumber(cctx, new(big.Int).SetUint64(block))
	if err != nil || h == nil {
		return time.Now().UnixMilli() // last-resort fallback
	}
	ms := int64(h.Time) * 1000

	s.tsMu.Lock()
	if _, exists := s.blockTs[block]; !exists {
		s.blockTs[block] = ms
		s.tsOrder = append(s.tsOrder, block)
		if len(s.tsOrder) > blockTsCacheMax {
			drop := s.tsOrder[0]
			s.tsOrder = s.tsOrder[1:]
			delete(s.blockTs, drop)
		}
	}
	s.tsMu.Unlock()
	return ms
}

// Start spawns the background pollers. Stops when ctx is cancelled.
func (s *LiveSource) Start(ctx context.Context) error {
	if err := s.refreshVerifiers(ctx); err != nil {
		return fmt.Errorf("initial verifiers: %w", err)
	}
	if err := s.backfillEvents(ctx); err != nil {
		return fmt.Errorf("backfill: %w", err)
	}
	s.recomputeTotals()
	// First topology refresh synchronously so /api/state has it before
	// the SPA mounts.
	s.refreshTopology(ctx)
	go s.watchChain(ctx)
	go s.pollTopology(ctx)
	go s.refreshLoop(ctx)
	s.startLogTail(ctx)
	return nil
}

// applyBinding is called by the log tailer when it learns a verifier's
// EOA or AXL peer ID from startup logs. We patch the matching
// VerifierInfo so the roster page surfaces axlPeerId without needing a
// separate join config.
func (s *LiveSource) applyBinding(role, addr, peerID string) {
	if addr == "" && peerID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.verifiers {
		if addr != "" && strings.EqualFold(s.verifiers[i].Address, addr) {
			s.verifiers[i].Role = role
			if peerID != "" {
				s.verifiers[i].AxlPeerID = peerID
			}
			return
		}
	}
	// We may have a peer ID for a role whose addr we already know but
	// haven't matched yet (race: addr line came in after peer line).
	// Re-walk on next bind once both fields populated.
}

// recordKHInvocation bumps the KH workflow counter. Called once per
// observed commit/reveal/settle in the verifier log tailer.
func (s *LiveSource) recordKHInvocation() {
	s.mu.Lock()
	s.totals.TotalKHInvocations++
	s.mu.Unlock()
}

// recordCommitTime stores the first observed commit timestamp for a
// (job, verifier) pair. Latency measurements anchor on this.
func (s *LiveSource) recordCommitTime(jobID, verifier string, ts int64) {
	if ts <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := jobID + "/" + strings.ToLower(verifier)
	if _, exists := s.commitTimes[key]; !exists {
		s.commitTimes[key] = ts
	}
}

// recordSettleLatency walks the job's commits and records (settleTs -
// commitTs) per verifier, in nanoseconds. recomputeTotals averages
// these into meanCommitToSettleSec.
func (s *LiveSource) recordSettleLatency(jobID string, settleTs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := jobID + "/"
	for k, ct := range s.commitTimes {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if settleTs <= ct {
			continue
		}
		s.settleLatNs = append(s.settleLatNs, (settleTs-ct)*1_000_000) // ms→ns
	}
	// Bound to the most recent 200 measurements (≈ 60+ jobs worth).
	if len(s.settleLatNs) > 200 {
		s.settleLatNs = s.settleLatNs[len(s.settleLatNs)-200:]
	}
}

// recordAXLSend timestamps an observed axl.send. recomputeTotals
// derives a per-minute rate over the last 60s.
func (s *LiveSource) recordAXLSend(ts int64) {
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.axlSendsTs = append(s.axlSendsTs, ts)
	// Drop entries older than 5 minutes.
	cutoff := time.Now().Add(-5 * time.Minute).UnixMilli()
	for len(s.axlSendsTs) > 0 && s.axlSendsTs[0] < cutoff {
		s.axlSendsTs = s.axlSendsTs[1:]
	}
}

// recordOGUpload bumps the OG upload counter and stashes the latest
// root per (role, jobId) into the job's ProofBundles list. This lets
// the UI's job inspector show all 3 bundle roots even when the chain
// itself doesn't carry them.
func (s *LiveSource) recordOGUpload(role, addr, jobIDStr, root, objectType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totals.OGUploadsTotal++
	// Only ProofBundle entries appear in the job detail; VoteRecord is
	// a per-iNFT log, attached to the verifier roster instead.
	if objectType != "ProofBundle" {
		return
	}
	var jobID uint64
	fmt.Sscanf(jobIDStr, "%d", &jobID)
	j, ok := s.jobs[jobID]
	if !ok {
		return
	}
	verifierAddr := addr
	if verifierAddr == "" {
		verifierAddr = role
	}
	for i := range j.detail.ProofBundles {
		if strings.EqualFold(j.detail.ProofBundles[i].Verifier, verifierAddr) {
			j.detail.ProofBundles[i].Root = root
			return
		}
	}
	j.detail.ProofBundles = append(j.detail.ProofBundles, ProofBundleRef{
		Verifier: verifierAddr,
		Root:     root,
	})
}

// refreshLoop refreshes verifier registry state and totals every 30s.
// Cheaper than re-running on every event and good enough for a roster
// page that updates after each settle.
func (s *LiveSource) refreshLoop(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = s.refreshVerifiers(ctx)
			s.recomputeTotals()
			s.hub.Broadcast(EventEnvelope{
				"kind":   "totals.update",
				"totals": s.snapshotTotals(),
			})
		}
	}
}

func (s *LiveSource) recomputeTotals() {
	s.mu.Lock()
	defer s.mu.Unlock()
	settled := 0
	pass := 0
	for _, j := range s.jobs {
		if j.summary.Status == "Settled" {
			settled++
			if j.summary.FinalVerdict == "PASS" {
				pass++
			}
		}
	}
	s.totals.JobsVerified = settled
	if settled > 0 {
		s.totals.PassRateBps = (pass * 10000) / settled
	}
	online := 0
	for _, n := range s.topology.Nodes {
		if n.Online && n.Role != "pub" {
			online++
		}
	}
	s.totals.SwarmOnline = online

	// Mean commit→settle latency, in seconds.
	if len(s.settleLatNs) > 0 {
		var sum int64
		for _, v := range s.settleLatNs {
			sum += v
		}
		s.totals.MeanCommitToSettleSec = int(sum / int64(len(s.settleLatNs)) / 1_000_000_000)
	}

	// AXL message rate over the trailing minute.
	cutoff := time.Now().Add(-1 * time.Minute).UnixMilli()
	cnt := 0
	for _, t := range s.axlSendsTs {
		if t >= cutoff {
			cnt++
		}
	}
	s.totals.AxlMessageRatePerMin = cnt
}

func (s *LiveSource) snapshotTotals() SystemTotals {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totals
}

// State, Jobs, Job, Verifiers, Topology, ProofBundle, PostJob.

func (s *LiveSource) State() (StateSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return StateSnapshot{
		Totals:     s.totals,
		RecentJobs: s.recentJobsLocked(20),
		Verifiers:  append([]VerifierInfo{}, s.verifiers...),
		Topology:   s.topology,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

func (s *LiveSource) Jobs(limit, offset int) ([]JobSummary, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.recentJobsLocked(0)
	total := len(all)
	if offset > total {
		return []JobSummary{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (s *LiveSource) Job(id string) (JobDetail, error) {
	jid, ok := new(big.Int).SetString(id, 10)
	if !ok {
		return JobDetail{}, fmt.Errorf("bad job id: %q", id)
	}
	s.mu.RLock()
	j, found := s.jobs[jid.Uint64()]
	s.mu.RUnlock()
	if found {
		// Defensive: ensure slices are non-nil so JSON serializes as
		// `[]` (the SPA's array operations choke on `null`). The mirror
		// only allocates these on first append.
		d := j.detail
		if d.Commits == nil {
			d.Commits = []VerifierVote{}
		}
		if d.Reveals == nil {
			d.Reveals = []VerifierVote{}
		}
		if d.ProofBundles == nil {
			d.ProofBundles = []ProofBundleRef{}
		}
		return d, nil
	}
	// Not in mirror — fetch directly from chain.
	return s.fetchJob(context.Background(), jid)
}

func (s *LiveSource) Verifiers() ([]VerifierInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]VerifierInfo{}, s.verifiers...), nil
}

func (s *LiveSource) Topology() (TopologyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.topology, nil
}

func (s *LiveSource) ProofBundle(root string) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	return downloadProofBundle(ctx, root)
}

func (s *LiveSource) PostJob() (PostJobResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	return s.runPostJob(ctx)
}

func (s *LiveSource) PostJobPreview() (PostJobPreview, error) {
	return s.buildPostJobPreview()
}

// recentJobsLocked returns jobs sorted by postedAt desc. Caller holds mu.
func (s *LiveSource) recentJobsLocked(limit int) []JobSummary {
	out := make([]JobSummary, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j.summary)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PostedAt > out[j].PostedAt })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// helpers shared by chain watchers

func hexHash(b [32]byte) string {
	return "0x" + hex.EncodeToString(b[:])
}

func verdictStr(v bool) string {
	if v {
		return "PASS"
	}
	return "FAIL"
}

func statusStr(s uint8) string {
	switch s {
	case 0:
		return "None"
	case 1:
		return "Posted"
	case 2:
		return "ClaimSubmitted"
	case 3:
		return "Settled"
	case 4:
		return "Cancelled"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}
