package verifier

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hpsing/aegis/internal/axl"
)

// axlRecvLoop polls the local AXL daemon for inbound messages, dispatches
// them by envelope type, and updates the verifier's local caches.
//
// The loop runs as a side-channel to the chain event loop — chain
// remains the source of truth for consensus; AXL is the off-chain
// transport for unstructured data (specs, vote replicas).
func (l *Loop) axlRecvLoop(ctx context.Context) {
	if l.cfg.AXL == nil {
		return
	}
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		// Drain anything buffered.
		for {
			msg, err := l.cfg.AXL.Recv(ctx)
			if err != nil {
				l.logf("axl recv: %v", err)
				break
			}
			if msg == nil {
				break
			}
			l.handleAXL(msg)
		}
	}
}

// handleAXL dispatches an inbound AXL message by envelope type.
func (l *Loop) handleAXL(msg *axl.Message) {
	env, err := axl.UnmarshalEnvelope(msg.Payload)
	if err != nil {
		l.logf("axl: bad envelope from %s…: %v", short(msg.FromPeerID), err)
		return
	}
	switch env.Type {
	case axl.TypeSpecPublish:
		l.onSpecPublish(env, msg.FromPeerID)
	case axl.TypeVoteCommit, axl.TypeVoteReveal:
		// Replicated vote from another verifier. Useful for the demo UI's
		// live event stream + auditor reconstruction. We don't act on it
		// for consensus — chain commit-reveal is authoritative.
		l.logf("axl: %s envelope from %s…", env.Type, short(msg.FromPeerID))
	case axl.TypeAgentCard:
		l.logf("axl: agent_card from %s…", short(msg.FromPeerID))
	default:
		l.logf("axl: unknown envelope type %q from %s…", env.Type, short(msg.FromPeerID))
	}
}

// onSpecPublish validates a received spec against its claimed hash and
// caches it. The chain's specHash is the authoritative bind; we verify
// keccak256(spec) == specHash so a malicious peer can't poison.
func (l *Loop) onSpecPublish(env axl.Envelope, fromPeer string) {
	var p axl.SpecPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		l.logf("axl spec: payload decode: %v", err)
		return
	}
	want := strings.ToLower(strings.TrimPrefix(p.SpecHash, "0x"))
	got := hex.EncodeToString(crypto.Keccak256(p.Spec))
	if want != got {
		l.logf("axl spec: hash mismatch (claimed=0x%s computed=0x%s)", want, got)
		return
	}
	l.mu.Lock()
	if l.specCache == nil {
		l.specCache = map[string][]byte{}
	}
	l.specCache[want] = p.Spec
	l.mu.Unlock()
	l.logf("axl spec: cached %s… (%d bytes) from %s…", want[:10], len(p.Spec), short(fromPeer))
}

// publishVoteEnvelope replicates a vote (commit or reveal) to AXL so
// auditors / the demo UI / other verifiers can index it. Best-effort:
// chain side has already landed by the time this runs.
func (l *Loop) publishVoteEnvelope(ctx context.Context, p axl.VotePayload, peerIDs []string) {
	if l.cfg.AXL == nil || len(peerIDs) == 0 {
		return
	}
	envType := axl.TypeVoteCommit
	if p.Phase == "reveal" {
		envType = axl.TypeVoteReveal
	}
	body, err := axl.MarshalEnvelope(envType, p)
	if err != nil {
		l.logf("axl publish vote: marshal: %v", err)
		return
	}
	for _, peer := range peerIDs {
		if err := l.cfg.AXL.Send(ctx, peer, body); err != nil {
			l.logf("axl publish vote → %s…: %v", short(peer), err)
		}
	}
}

// axlPeerList returns the verifier's configured peer list for vote
// envelope replication. We don't have explicit peers wired today —
// publish to whatever has been seen via spec_publish or agent_card.
// For the demo this returns a static env-derived list.
func (l *Loop) axlPeerList() []string {
	if v := os.Getenv("AXL_REPLICATE_PEERS"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return nil
}

func short(s string) string {
	if len(s) <= 10 {
		return s
	}
	return s[:10]
}
