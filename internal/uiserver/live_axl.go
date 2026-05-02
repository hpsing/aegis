package uiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

// pollTopology hits each AXL daemon's /topology and aggregates into
// the mirror. Pushes a topology.update event when the snapshot changes.
func (s *LiveSource) pollTopology(ctx context.Context) {
	t := time.NewTicker(s.cfg.TopologyInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		s.refreshTopology(ctx)
	}
}

type axlTopologyDoc struct {
	OurIPv6      string          `json:"our_ipv6"`
	OurPublicKey string          `json:"our_public_key"`
	Peers        json.RawMessage `json:"peers"`
}

type axlPeer struct {
	URI       string `json:"uri"`
	Up        bool   `json:"up"`
	Inbound   bool   `json:"inbound"`
	PublicKey string `json:"public_key"`
}

func (s *LiveSource) refreshTopology(ctx context.Context) {
	if len(s.cfg.AXLDaemons) == 0 {
		return
	}
	httpc := &http.Client{Timeout: 1500 * time.Millisecond}

	roles := make([]string, 0, len(s.cfg.AXLDaemons))
	for r := range s.cfg.AXLDaemons {
		roles = append(roles, r)
	}
	sort.Strings(roles) // deterministic order

	nodes := make([]TopologyNode, 0, len(roles))
	peerToRole := map[string]string{}
	edges := []TopologyEdge{}
	allPeers := map[string][]axlPeer{} // role → peers list

	for _, role := range roles {
		url := s.cfg.AXLDaemons[role]
		req, _ := http.NewRequestWithContext(ctx, "GET", url+"/topology", nil)
		res, err := httpc.Do(req)
		online := err == nil && res != nil && res.StatusCode == 200
		var doc axlTopologyDoc
		var peers []axlPeer
		if online {
			defer res.Body.Close()
			if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
				online = false
			} else {
				_ = json.Unmarshal(doc.Peers, &peers)
				allPeers[role] = peers
			}
		}
		nodes = append(nodes, TopologyNode{
			Role:   role,
			PeerID: doc.OurPublicKey,
			APIURL: url,
			IPv6:   doc.OurIPv6,
			Online: online,
		})
		if doc.OurPublicKey != "" {
			peerToRole[doc.OurPublicKey] = role
		}
	}

	// Build edges from each node's peers list.
	for _, role := range roles {
		for _, p := range allPeers[role] {
			if !p.Up || p.PublicKey == "" {
				continue
			}
			toRole, ok := peerToRole[p.PublicKey]
			if !ok {
				continue
			}
			// Deduplicate: only emit edge once per pair.
			if role >= toRole {
				continue
			}
			edges = append(edges, TopologyEdge{
				From:       role,
				To:         toRole,
				TreeParent: !p.Inbound,
			})
		}
	}

	freshness := time.Now().UnixMilli()
	s.mu.Lock()
	prev := s.topology
	s.topology = TopologyState{FreshnessMs: freshness, Nodes: nodes, Edges: edges}
	online := 0
	for _, n := range nodes {
		if n.Online && n.Role != "pub" {
			online++
		}
	}
	s.totals.SwarmOnline = online
	s.mu.Unlock()

	// Only broadcast on meaningful change to avoid spamming the SSE channel.
	if topologyDiffers(prev, nodes, edges) {
		s.hub.Broadcast(EventEnvelope{
			"kind": "topology.update",
			"topology": map[string]any{
				"freshnessMs": freshness,
				"nodes":       nodes,
				"edges":       edges,
			},
		})
	}
}

func topologyDiffers(prev TopologyState, nodes []TopologyNode, edges []TopologyEdge) bool {
	if len(prev.Nodes) != len(nodes) || len(prev.Edges) != len(edges) {
		return true
	}
	for i := range nodes {
		if prev.Nodes[i].Online != nodes[i].Online || prev.Nodes[i].PeerID != nodes[i].PeerID {
			return true
		}
	}
	return false
}
