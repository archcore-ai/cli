package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"archcore-cli/internal/agents"
)

func TestRunMCPInstall_NoArchcoreDir(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	err := runMCPInstall(base)
	if err == nil {
		t.Fatal("expected error without .archcore/")
	}
	if got := err.Error(); got != ".archcore/ not found — run 'archcore init' first" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestRunMCPInstall_FreshInstall(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	if err := runMCPInstall(base); err != nil {
		t.Fatalf("runMCPInstall: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(base, ".mcp.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	var servers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &servers); err != nil {
		t.Fatalf("Unmarshal mcpServers: %v", err)
	}

	if _, ok := servers["archcore"]; !ok {
		t.Error("missing 'archcore' in mcpServers")
	}

	var entry struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(servers["archcore"], &entry); err != nil {
		t.Fatalf("Unmarshal entry: %v", err)
	}
	if entry.Command != "archcore" {
		t.Errorf("command = %q, want %q", entry.Command, "archcore")
	}
	if len(entry.Args) != 1 || entry.Args[0] != "mcp" {
		t.Errorf("args = %v, want [mcp]", entry.Args)
	}
}

func TestRunMCPInstall_Idempotent(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	if err := runMCPInstall(base); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := runMCPInstall(base); err != nil {
		t.Fatalf("second install: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(base, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	var servers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &servers); err != nil {
		t.Fatal(err)
	}

	if len(servers) != 1 {
		t.Errorf("expected 1 server entry, got %d", len(servers))
	}
}

func TestRunMCPInstall_MergesExisting(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	// Pre-populate .mcp.json with an existing server.
	existing := map[string]any{
		"mcpServers": map[string]any{
			"other-tool": map[string]any{
				"command": "other-tool",
				"args":    []string{"serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(filepath.Join(base, ".mcp.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runMCPInstall(base); err != nil {
		t.Fatalf("runMCPInstall: %v", err)
	}

	resultData, err := os.ReadFile(filepath.Join(base, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resultData, &raw); err != nil {
		t.Fatal(err)
	}

	var servers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &servers); err != nil {
		t.Fatal(err)
	}

	if _, ok := servers["other-tool"]; !ok {
		t.Error("existing 'other-tool' was lost during merge")
	}
	if _, ok := servers["archcore"]; !ok {
		t.Error("missing 'archcore' after install")
	}
}

func TestMCPInstall_AgentFlag_Valid(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	if err := runMCPInstallForAgent(base, agents.Cursor); err != nil {
		t.Fatalf("runMCPInstallForAgent: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(base, ".cursor", "mcp.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["mcpServers"]; !ok {
		t.Error("missing mcpServers in .cursor/mcp.json")
	}
}

func TestMCPInstall_AgentFlag_Invalid(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	err := runMCPInstallForAgent(base, "nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid agent")
	}
}

func TestMCPInstall_AutoDetect(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)
	os.MkdirAll(filepath.Join(base, ".cursor"), 0o755)
	os.MkdirAll(filepath.Join(base, ".roo"), 0o755)

	if err := runMCPInstallAutoDetect(base); err != nil {
		t.Fatalf("runMCPInstallAutoDetect: %v", err)
	}

	// Both .cursor/mcp.json and .roo/mcp.json should exist.
	for _, path := range []string{".cursor/mcp.json", ".roo/mcp.json"} {
		if _, err := os.Stat(filepath.Join(base, path)); err != nil {
			t.Errorf("expected %s to exist: %v", path, err)
		}
	}
}

func TestMCPInstall_NoAgents_DefaultClaudeCode(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	if err := runMCPInstallAutoDetect(base); err != nil {
		t.Fatalf("runMCPInstallAutoDetect: %v", err)
	}

	// .mcp.json should exist (Claude Code fallback).
	if _, err := os.Stat(filepath.Join(base, ".mcp.json")); err != nil {
		t.Error("expected .mcp.json to exist (Claude Code fallback)")
	}
}

func TestMCPInstall_MultipleAgents(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)
	os.MkdirAll(filepath.Join(base, ".claude"), 0o755)
	os.MkdirAll(filepath.Join(base, ".gemini"), 0o755)
	os.MkdirAll(filepath.Join(base, ".codex"), 0o755)

	if err := runMCPInstallAutoDetect(base); err != nil {
		t.Fatalf("runMCPInstallAutoDetect: %v", err)
	}

	// Claude Code .mcp.json.
	if _, err := os.Stat(filepath.Join(base, ".mcp.json")); err != nil {
		t.Error("expected .mcp.json")
	}
	// Gemini CLI .gemini/settings.json.
	if _, err := os.Stat(filepath.Join(base, ".gemini", "settings.json")); err != nil {
		t.Error("expected .gemini/settings.json")
	}
	// Codex CLI .codex/config.toml.
	if _, err := os.Stat(filepath.Join(base, ".codex", "config.toml")); err != nil {
		t.Error("expected .codex/config.toml")
	}
}
