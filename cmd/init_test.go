package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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

	// Verify exact JSON format — should have project_id but not archcore_url.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["project_id"]; !ok {
		t.Error("cloud settings should have project_id field")
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

	// Verify exact JSON format — should have both project_id and archcore_url.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["project_id"]; !ok {
		t.Error("on-prem settings should have project_id field")
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
	if result == nil {
		t.Fatal("expected non-nil result even on server error")
	}
	// Dirs should still be created even though server is unreachable.
	if !config.DirExists(base) {
		t.Error(".archcore/ directory should be created even when server is unreachable")
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
