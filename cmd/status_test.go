package cmd

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"archcore-cli/internal/config"
)

func runStatusCmd(t *testing.T, dir string, jsonFlag bool) (string, error) {
	t.Helper()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	args := []string{"status"}
	if jsonFlag {
		args = append(args, "--json")
	}
	root.SetArgs(args)
	// Capture stdout.
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("os.Pipe: %v", pipeErr)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	err := root.Execute()

	w.Close()
	os.Stdout = oldStdout
	var out bytes.Buffer
	out.ReadFrom(r)
	return out.String(), err
}

func TestStatus_NotInitialized(t *testing.T) {
	dir := t.TempDir()
	out, err := runStatusCmd(t, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Not initialized") {
		t.Errorf("expected 'Not initialized' in output, got: %s", out)
	}
	if !strings.Contains(out, "archcore init") {
		t.Errorf("expected hint to run 'archcore init', got: %s", out)
	}
}

func TestStatus_NotInitialized_JSON(t *testing.T) {
	dir := t.TempDir()
	out, err := runStatusCmd(t, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, out)
	}
	if _, ok := result["error"]; !ok {
		t.Error("expected error field in JSON output")
	}
}

func TestStatus_SyncNone(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewNoneSettings()); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "none") {
		t.Errorf("expected 'none' sync type in output, got: %s", out)
	}
	if strings.Contains(out, "reachable") || strings.Contains(out, "unreachable") {
		t.Errorf("none sync should not show server status, got: %s", out)
	}
}

func TestStatus_SyncNone_JSON(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewNoneSettings()); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
	}
	if result["sync"] != "none" {
		t.Errorf("sync = %v, want none", result["sync"])
	}
	if _, ok := result["connected"]; ok {
		t.Error("none sync should not have connected field")
	}
}

func TestStatus_SyncCloud_JSON(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	defer srv.Close()

	orig := config.CloudServerURL
	config.CloudServerURL = srv.URL
	defer func() { config.CloudServerURL = orig }()

	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewCloudSettings()); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
	}
	if result["sync"] != "cloud" {
		t.Errorf("sync = %v, want cloud", result["sync"])
	}
	if result["connected"] != true {
		t.Errorf("connected = %v, want true", result["connected"])
	}
}

func TestStatus_SyncOnPrem_JSON(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	defer srv.Close()

	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewOnPremSettings(srv.URL)); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
	}
	if result["sync"] != "on-prem" {
		t.Errorf("sync = %v, want on-prem", result["sync"])
	}
	if result["archcore_url"] != srv.URL {
		t.Errorf("archcore_url = %v, want %s", result["archcore_url"], srv.URL)
	}
	if result["connected"] != true {
		t.Errorf("connected = %v, want true", result["connected"])
	}
}

func TestStatus_InvalidSettings(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	// Write invalid settings (unknown field).
	if err := os.WriteFile(dir+"/.archcore/settings.json", []byte(`{"sync":"none","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Invalid") {
		t.Errorf("expected 'Invalid' in output, got: %s", out)
	}
	if !strings.Contains(out, "archcore init") {
		t.Errorf("expected hint to run 'archcore init', got: %s", out)
	}
}

func TestStatus_InvalidSettings_JSON(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/.archcore/settings.json", []byte(`{"sync":"none","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runStatusCmd(t, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, out)
	}
	if _, ok := result["error"]; !ok {
		t.Error("expected error field in JSON output for invalid settings")
	}
}
