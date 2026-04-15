package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"archcore-cli/internal/agents"
	"archcore-cli/internal/config"
)

func healthyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ready":true}`))
	}
}

func TestRunInit_SyncNone(t *testing.T) {
	base := t.TempDir()
	settings := config.NewNoneSettings()
	result, err := runInit(context.Background(), base, settings)
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if result.serverReachable {
		t.Error("serverReachable should be false for sync none")
	}
	if !config.DirExists(base) {
		t.Error(".archcore/ directory not created")
	}
	s, err := config.Load(base)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Sync != config.SyncTypeNone {
		t.Errorf("Sync = %q, want %q", s.Sync, config.SyncTypeNone)
	}

	// Verify exact JSON format.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["project_id"]; ok {
		t.Error("none settings should not have project_id field")
	}
	if _, ok := raw["archcore_url"]; ok {
		t.Error("none settings should not have archcore_url field")
	}
}

func TestRunInit_SyncCloud(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	defer srv.Close()

	// Override CloudServerURL for test.
	orig := config.CloudServerURL
	config.CloudServerURL = srv.URL
	defer func() { config.CloudServerURL = orig }()

	base := t.TempDir()
	settings := config.NewCloudSettings()
	result, err := runInit(context.Background(), base, settings)
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if !result.serverReachable {
		t.Error("serverReachable should be true")
	}
	s, err := config.Load(base)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Sync != config.SyncTypeCloud {
		t.Errorf("Sync = %q, want %q", s.Sync, config.SyncTypeCloud)
	}

	// Verify exact JSON format — should not have project_id (nil) or archcore_url.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["project_id"]; ok {
		t.Error("cloud settings should not have project_id field when nil")
	}
	if _, ok := raw["archcore_url"]; ok {
		t.Error("cloud settings should not have archcore_url field")
	}
}

func TestRunInit_SyncOnPrem(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	defer srv.Close()

	base := t.TempDir()
	settings := config.NewOnPremSettings(srv.URL)
	result, err := runInit(context.Background(), base, settings)
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if !result.serverReachable {
		t.Error("serverReachable should be true")
	}
	s, err := config.Load(base)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Sync != config.SyncTypeOnPrem {
		t.Errorf("Sync = %q, want %q", s.Sync, config.SyncTypeOnPrem)
	}
	if s.ArchcoreURL != srv.URL {
		t.Errorf("ArchcoreURL = %q, want %q", s.ArchcoreURL, srv.URL)
	}

	// Verify exact JSON format — should not have project_id (nil), should have archcore_url.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["project_id"]; ok {
		t.Error("on-prem settings should not have project_id field when nil")
	}
	if _, ok := raw["archcore_url"]; !ok {
		t.Error("on-prem settings should have archcore_url field")
	}
}

func TestRunInit_ServerUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close() // close immediately

	base := t.TempDir()
	settings := config.NewOnPremSettings(srv.URL)
	result, err := runInit(context.Background(), base, settings)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if result != nil {
		t.Fatal("expected nil result on server error")
	}
	if !errors.Is(err, ErrServerUnreachable) {
		t.Fatalf("expected ErrServerUnreachable, got: %v", err)
	}
	// Dirs should still be created even though server is unreachable.
	if !config.DirExists(base) {
		t.Error(".archcore/ directory should be created even when server is unreachable")
	}
}

func TestRunInit_InstallsHooksAndMCP(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	settings := config.NewNoneSettings()
	if _, err := runInit(context.Background(), base, settings); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if err := runHooksInstallForAgent(base, agents.ClaudeCode); err != nil {
		t.Fatalf("runHooksInstallForAgent: %v", err)
	}

	// Verify .claude/settings.json has all 3 hook events.
	data, err := os.ReadFile(filepath.Join(base, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile .claude/settings.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	var hooks map[string][]hookMatcher
	if err := json.Unmarshal(raw["hooks"], &hooks); err != nil {
		t.Fatalf("Unmarshal hooks: %v", err)
	}
	for _, event := range []string{"SessionStart"} {
		matchers, ok := hooks[event]
		if !ok {
			t.Errorf("missing hook event %s", event)
			continue
		}
		if len(matchers) != 1 {
			t.Errorf("event %s: want 1 matcher, got %d", event, len(matchers))
		}
	}

	// Verify .mcp.json has the archcore server entry.
	mcpData, err := os.ReadFile(filepath.Join(base, ".mcp.json"))
	if err != nil {
		t.Fatalf("ReadFile .mcp.json: %v", err)
	}
	var mcpRaw map[string]json.RawMessage
	if err := json.Unmarshal(mcpData, &mcpRaw); err != nil {
		t.Fatalf("Unmarshal .mcp.json: %v", err)
	}
	var servers map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw["mcpServers"], &servers); err != nil {
		t.Fatalf("Unmarshal mcpServers: %v", err)
	}
	if _, ok := servers["archcore"]; !ok {
		t.Error("missing archcore entry in .mcp.json mcpServers")
	}
}

func TestRunInit_HooksIdempotent(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	settings := config.NewNoneSettings()
	if _, err := runInit(context.Background(), base, settings); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Run hooks install twice.
	for i := 0; i < 2; i++ {
		if err := runHooksInstallForAgent(base, agents.ClaudeCode); err != nil {
			t.Fatalf("runHooksInstallForAgent call %d: %v", i+1, err)
		}
	}

	// Verify exactly 1 matcher per hook event (no duplicates).
	data, err := os.ReadFile(filepath.Join(base, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	var hooks map[string][]hookMatcher
	if err := json.Unmarshal(raw["hooks"], &hooks); err != nil {
		t.Fatalf("Unmarshal hooks: %v", err)
	}
	for _, event := range []string{"SessionStart"} {
		if len(hooks[event]) != 1 {
			t.Errorf("event %s: want 1 matcher after idempotent install, got %d", event, len(hooks[event]))
		}
	}

	// Verify exactly 1 archcore entry in .mcp.json.
	mcpData, err := os.ReadFile(filepath.Join(base, ".mcp.json"))
	if err != nil {
		t.Fatalf("ReadFile .mcp.json: %v", err)
	}
	var mcpRaw map[string]json.RawMessage
	if err := json.Unmarshal(mcpData, &mcpRaw); err != nil {
		t.Fatalf("Unmarshal .mcp.json: %v", err)
	}
	var servers map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw["mcpServers"], &servers); err != nil {
		t.Fatalf("Unmarshal mcpServers: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("want 1 MCP server entry, got %d", len(servers))
	}
}

func TestRunInit_Idempotent(t *testing.T) {
	base := t.TempDir()
	for i := 0; i < 2; i++ {
		_, err := runInit(context.Background(), base, config.NewNoneSettings())
		if err != nil {
			t.Fatalf("runInit call %d: %v", i+1, err)
		}
	}
	s, err := config.Load(base)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Sync != config.SyncTypeNone {
		t.Errorf("Sync = %q, want %q", s.Sync, config.SyncTypeNone)
	}
}

func TestInit_DetectsMultipleAgents(t *testing.T) {
	t.Parallel()
	base := t.TempDir()

	// Create agent marker directories before init.
	if err := os.MkdirAll(filepath.Join(base, ".cursor"), 0o755); err != nil {
		t.Fatalf("MkdirAll .cursor: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(base, ".gemini"), 0o755); err != nil {
		t.Fatalf("MkdirAll .gemini: %v", err)
	}

	settings := config.NewNoneSettings()
	if _, err := runInit(context.Background(), base, settings); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Now detect and install for detected agents.
	detected := agents.Detect(base)
	for _, agent := range detected {
		switch agent.ID {
		case agents.Cursor:
			runCursorHooksInstall(base)
		case agents.GeminiCLI:
			runGeminiCLIHooksInstall(base)
		}
	}

	// Verify .cursor/hooks.json exists.
	if _, err := os.Stat(filepath.Join(base, ".cursor", "hooks.json")); err != nil {
		t.Error("expected .cursor/hooks.json to exist")
	}
	// Verify .gemini/settings.json has hooks.
	data, err := os.ReadFile(filepath.Join(base, ".gemini", "settings.json"))
	if err != nil {
		t.Fatal("expected .gemini/settings.json to exist")
	}
	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)
	if _, ok := raw["hooks"]; !ok {
		t.Error("expected hooks in .gemini/settings.json")
	}
}

func TestResolveAgents_NoAgents_NonInteractive(t *testing.T) {
	base := t.TempDir()
	origIsInteractive := isInteractive
	isInteractive = func() bool { return false }
	defer func() { isInteractive = origIsInteractive }()

	settings := config.NewNoneSettings()
	if _, err := runInit(context.Background(), base, settings); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	resolved, err := resolveAgents(base)
	if err != nil {
		t.Fatalf("resolveAgents: %v", err)
	}
	if len(resolved) != 0 {
		t.Fatalf("expected no resolved agents, got %d", len(resolved))
	}
}
