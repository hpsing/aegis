package keeperhub

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config matches configs/keeperhub.toml. Each verifier process loads
// the whole file and selects its own [verifier_N] section by index.
type Config struct {
	AegisContract string `toml:"aegis_contract"`
	ChainID       uint64 `toml:"chain_id"`
	MCPEndpoint   string `toml:"mcp_endpoint"`

	Verifier1 VerifierConfig `toml:"verifier_1"`
	Verifier2 VerifierConfig `toml:"verifier_2"`
	Verifier3 VerifierConfig `toml:"verifier_3"`
}

// VerifierConfig is the per-verifier slice of the KH config.
//
// Each verifier has:
//   - one KH API key (via env, not in this file) authenticating to its
//     own org;
//   - a wallet integration in that org (we don't actually need the
//     integration id since we sign locally — kept for reference);
//   - three published workflow ids (one per Purpose) for execute_workflow;
//   - three listed slugs (one per Purpose) for call_workflow. The
//     LiveClient uses slugs (call_workflow path).
type VerifierConfig struct {
	Address             string `toml:"address"`
	WalletIntegrationID string `toml:"wallet_integration_id"`

	// Workflow ids — used by execute_workflow path. Retained for tooling
	// (update_workflow / unlist_workflow) that operates by id.
	CommitWorkflowID string `toml:"commit_workflow_id"`
	RevealWorkflowID string `toml:"reveal_workflow_id"`
	SettleWorkflowID string `toml:"settle_workflow_id"`

	// Listed slugs — what call_workflow consumes. Populated by
	// scripts/list-keeperhub-workflows.sh after `update_workflow_listing`.
	CommitWorkflowSlug string `toml:"commit_workflow_slug"`
	RevealWorkflowSlug string `toml:"reveal_workflow_slug"`
	SettleWorkflowSlug string `toml:"settle_workflow_slug"`
}

// LoadConfig reads + parses configs/keeperhub.toml.
func LoadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if !strings.HasPrefix(cfg.AegisContract, "0x") {
		return nil, fmt.Errorf("aegis_contract missing or not 0x-prefixed")
	}
	if cfg.ChainID == 0 {
		return nil, fmt.Errorf("chain_id must be set")
	}
	return &cfg, nil
}

// VerifierByIndex returns the per-verifier slice for the given 1-based
// index (1, 2, or 3).
func (c *Config) VerifierByIndex(idx int) (VerifierConfig, error) {
	switch idx {
	case 1:
		return c.Verifier1, nil
	case 2:
		return c.Verifier2, nil
	case 3:
		return c.Verifier3, nil
	default:
		return VerifierConfig{}, fmt.Errorf("verifier index %d out of range (want 1-3)", idx)
	}
}

// WorkflowIDs returns the Purpose→id map (legacy execute_workflow path).
func (v VerifierConfig) WorkflowIDs() map[Purpose]string {
	return map[Purpose]string{
		PurposeCommit: v.CommitWorkflowID,
		PurposeReveal: v.RevealWorkflowID,
		PurposeSettle: v.SettleWorkflowID,
	}
}

// WorkflowSlugs returns the Purpose→listedSlug map — what
// LiveClientConfig.WorkflowSlugs wants for the call_workflow path.
func (v VerifierConfig) WorkflowSlugs() map[Purpose]string {
	return map[Purpose]string{
		PurposeCommit: v.CommitWorkflowSlug,
		PurposeReveal: v.RevealWorkflowSlug,
		PurposeSettle: v.SettleWorkflowSlug,
	}
}
