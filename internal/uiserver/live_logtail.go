package uiserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// startLogTail discovers the most recent logs/e2e-* subdir and tails
// each verifier-N.log forever. Lines are parsed against a small set of
// regexes to derive KH workflow events, 0G upload events, and the
// (role → wallet, role → AXL peerID) bindings used to populate
// VerifierInfo.AxlPeerId.
func (s *LiveSource) startLogTail(ctx context.Context) {
	if s.cfg.LogsRoot == "" {
		return
	}
	t := s.cfg.LogPollInterval
	if t == 0 {
		t = 1 * time.Second
	}
	tail := newLogTailer(s, s.cfg.LogsRoot)
	go func() {
		// Initial scan — pick up any in-flight or finished run.
		tail.refreshTargets()
		ticker := time.NewTicker(t)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				tail.closeAll()
				return
			case <-ticker.C:
				tail.refreshTargets()
				tail.pollAll()
			}
		}
	}()
}

// roleBinding captures what the log tailer learns about a verifier:
// its 0x EOA + its AXL peer ID, derived from the startup log lines.
type roleBinding struct {
	addr   string
	peerID string
}

type logTailer struct {
	src        *LiveSource
	root       string
	mu         sync.Mutex
	currentDir string                 // most-recent logs/e2e-* dir
	files      map[string]*tailedFile // role → open file handle + offset
	bindings   map[string]roleBinding // role → (addr, peerID)
}

type tailedFile struct {
	path   string
	f      *os.File
	offset int64
	role   string
}

func newLogTailer(src *LiveSource, root string) *logTailer {
	return &logTailer{
		src:      src,
		root:     root,
		files:    map[string]*tailedFile{},
		bindings: map[string]roleBinding{},
	}
}

// refreshTargets finds the most-recent logs/e2e-* dir under root. If
// it differs from the current one, close existing handles and reopen
// against the new dir from offset 0 (so we replay the new run cleanly).
func (l *logTailer) refreshTargets() {
	dir, err := latestRunDir(l.root)
	if err != nil || dir == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if dir != l.currentDir {
		// New run: close old handles, drop bindings (they're stale).
		for _, t := range l.files {
			if t.f != nil {
				_ = t.f.Close()
			}
		}
		l.files = map[string]*tailedFile{}
		l.bindings = map[string]roleBinding{}
		l.currentDir = dir
	}

	// Open whichever verifier-N.log files exist that we haven't yet.
	for _, role := range []string{"v1", "v2", "v3"} {
		idx := role[1:] // "1"|"2"|"3"
		path := filepath.Join(dir, "verifier-"+idx+".log")
		if _, exists := l.files[role]; exists {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue // not yet created
		}
		l.files[role] = &tailedFile{path: path, f: f, offset: 0, role: role}
	}
}

func (l *logTailer) closeAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, t := range l.files {
		if t.f != nil {
			_ = t.f.Close()
		}
	}
}

// pollAll reads new content from each open file and parses each new
// line. Stops at EOF; resumes from saved offset next tick.
func (l *logTailer) pollAll() {
	l.mu.Lock()
	files := make([]*tailedFile, 0, len(l.files))
	for _, t := range l.files {
		files = append(files, t)
	}
	l.mu.Unlock()

	for _, t := range files {
		if _, err := t.f.Seek(t.offset, io.SeekStart); err != nil {
			continue
		}
		r := bufio.NewReader(t.f)
		for {
			line, err := r.ReadString('\n')
			if line != "" {
				l.handleLine(t.role, strings.TrimRight(line, "\n"))
				t.offset += int64(len(line))
			}
			if err != nil {
				break // io.EOF on partial lines we'll pick up next tick
			}
		}
	}
}

// regexes for the line patterns we extract. See the verifier source
// (internal/verifier/loop.go, axl_bridge.go) for where they're emitted.
var (
	reLoopRunning = regexp.MustCompile(`\[(v\d)\] loop running, addr=(0x[a-fA-F0-9]{40})`)
	reAXLBoot     = regexp.MustCompile(`\[verifier\] AXL: (\S+) peer=([a-f0-9]+)…`)
	reCommit      = regexp.MustCompile(`\[(v\d)\] commit (\d+) verdict=(\w+) tx=(0x[a-fA-F0-9]{64})`)
	reReveal      = regexp.MustCompile(`\[(v\d)\] reveal (\d+) verdict=(\w+) tx=(0x[a-fA-F0-9]{64})`)
	reStorageApp  = regexp.MustCompile(`\[(v\d)\] storage append job=(\d+) entry=0g://(0x[a-fA-F0-9]{64})`)
	reProofBundle = regexp.MustCompile(`\[(v\d)\] proof bundle (\d+) -> 0g://(0x[a-fA-F0-9]{64})`)
	reAXLSpec     = regexp.MustCompile(`\[(v\d)\] axl spec: cached ([a-f0-9]+)…`)
	reAXLVote     = regexp.MustCompile(`\[(v\d)\] axl: (vote_commit|vote_reveal) envelope from`)
)

func (l *logTailer) handleLine(role, line string) {
	// Order matters: addr binding lines first (they set up the role→addr
	// map used by event emitters); the AXL boot line may precede the
	// loop-running line, so handle both before the events.
	if m := reLoopRunning.FindStringSubmatch(line); m != nil {
		l.bind(m[1], "addr", strings.ToLower(m[2]))
		return
	}
	// The AXL boot line uses the [verifier] prefix (not [vN]); it
	// appears in the same file as loop-running so role is the file's
	// role, not what's in the line.
	if m := reAXLBoot.FindStringSubmatch(line); m != nil {
		// peer is logged truncated (10 chars + …); we still record it.
		l.bind(role, "peer", m[2])
		return
	}
	if m := reCommit.FindStringSubmatch(line); m != nil {
		l.src.recordKHInvocation()
		addr := l.addrOf(role)
		l.src.hub.Broadcast(EventEnvelope{
			"kind":     "kh.workflow",
			"verifier": addr,
			"role":     role,
			"purpose":  "commit",
			"jobId":    m[2],
			"verdict":  verdictFromBool(m[3]),
			"tx":       m[4],
		})
		return
	}
	if m := reReveal.FindStringSubmatch(line); m != nil {
		l.src.recordKHInvocation()
		addr := l.addrOf(role)
		l.src.hub.Broadcast(EventEnvelope{
			"kind":     "kh.workflow",
			"verifier": addr,
			"role":     role,
			"purpose":  "reveal",
			"jobId":    m[2],
			"verdict":  verdictFromBool(m[3]),
			"tx":       m[4],
		})
		return
	}
	if m := reStorageApp.FindStringSubmatch(line); m != nil {
		addr := l.addrOf(role)
		l.src.recordOGUpload(role, addr, m[2], m[3], "VoteRecord")
		l.src.hub.Broadcast(EventEnvelope{
			"kind":       "og.upload",
			"verifier":   addr,
			"role":       role,
			"jobId":      m[2],
			"root":       m[3],
			"objectType": "VoteRecord",
		})
		return
	}
	if m := reProofBundle.FindStringSubmatch(line); m != nil {
		addr := l.addrOf(role)
		l.src.recordOGUpload(role, addr, m[2], m[3], "ProofBundle")
		l.src.hub.Broadcast(EventEnvelope{
			"kind":       "og.upload",
			"verifier":   addr,
			"role":       role,
			"jobId":      m[2],
			"root":       m[3],
			"objectType": "ProofBundle",
		})
		return
	}
	if m := reAXLSpec.FindStringSubmatch(line); m != nil {
		l.src.hub.Broadcast(EventEnvelope{
			"kind":     "axl.spec_received",
			"role":     m[1],
			"specHash": m[2],
		})
		return
	}
	if m := reAXLVote.FindStringSubmatch(line); m != nil {
		l.src.recordAXLSend(0)
		l.src.hub.Broadcast(EventEnvelope{
			"kind":     "axl.send",
			"to":       m[1],
			"envelope": m[2],
		})
		return
	}
}

func (l *logTailer) bind(role, field, value string) {
	l.mu.Lock()
	b := l.bindings[role]
	switch field {
	case "addr":
		b.addr = value
	case "peer":
		b.peerID = value
	}
	l.bindings[role] = b
	l.mu.Unlock()
	if b.addr != "" || b.peerID != "" {
		l.src.applyBinding(role, b.addr, b.peerID)
	}
}

func (l *logTailer) addrOf(role string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.bindings[role].addr
}

// latestRunDir picks the most-recent run subdir of root. Two layouts
// are tolerated:
//
//  1. root is the PARENT of run dirs (e.g. "logs/"). We look for
//     `e2e-*` / `demo-*` subdirs by lexical-newest, requiring at
//     least one verifier-N.log inside.
//  2. root IS a run dir already (e.g. "logs/demo-20260501-…"). If
//     root itself contains verifier-N.log files, we return it
//     verbatim. This makes UI_LOGS_DIR forgiving — operators can
//     point at the parent or the specific run; both work.
func latestRunDir(root string) (string, error) {
	if hasVerifierLog(root) {
		return root, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	candidates := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "e2e-") || strings.HasPrefix(n, "demo-") {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return "", nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(candidates)))
	for _, name := range candidates {
		dir := filepath.Join(root, name)
		if hasVerifierLog(dir) {
			return dir, nil
		}
	}
	return filepath.Join(root, candidates[0]), nil
}

func hasVerifierLog(dir string) bool {
	for _, idx := range []string{"1", "2", "3"} {
		if _, err := os.Stat(filepath.Join(dir, "verifier-"+idx+".log")); err == nil {
			return true
		}
	}
	return false
}

func verdictFromBool(s string) string {
	if s == "true" {
		return "PASS"
	}
	return "FAIL"
}

var _ = fmt.Sprintf // keep fmt import if unused later
