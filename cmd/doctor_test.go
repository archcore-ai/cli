package cmd

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"archcore-cli/internal/config"
)

func runDoctorInDir(t *testing.T, dir string) string {
	t.Helper()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root := NewRootCmd("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor"})

	// Capture stdout.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	execErr := root.Execute()
	w.Close()
	os.Stdout = oldStdout

	var out bytes.Buffer
	out.ReadFrom(r)

	if execErr != nil {
		t.Fatalf("doctor command failed: %v", err)
	}
	return out.String()
}

func TestDoctor_NotInitialized(t *testing.T) {
	dir := t.TempDir()
	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in output, got: %s", out)
	}
}

func TestDoctor_FreeFormDirectoryAllowed(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewNoneSettings()); err != nil {
		t.Fatal(err)
	}
	// Custom directories are fine — no required subdirs.
	os.MkdirAll(filepath.Join(dir, ".archcore", "auth"), 0o755)

	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "All checks passed") {
		t.Errorf("expected 'All checks passed', got: %s", out)
	}
}

func TestDoctor_InvalidSettings(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	// Write invalid settings (unknown field triggers validation error).
	if err := os.WriteFile(filepath.Join(dir, ".archcore", "settings.json"), []byte(`{"sync":"none","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "Settings") {
		t.Errorf("expected 'Settings' label in output, got: %s", out)
	}
	if !strings.Contains(out, "not allowed") {
		t.Errorf("expected 'not allowed' validation error in output, got: %s", out)
	}
	if !strings.Contains(out, "issue") {
		t.Errorf("expected 'issue' count in output, got: %s", out)
	}
}

func TestDoctor_SyncNone_NoServerCheck(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewNoneSettings()); err != nil {
		t.Fatal(err)
	}
	out := runDoctorInDir(t, dir)
	if strings.Contains(out, "reachable") || strings.Contains(out, "unreachable") {
		t.Errorf("none sync should not check server, got: %s", out)
	}
	if !strings.Contains(out, "All checks passed") {
		t.Errorf("expected all checks passed, got: %s", out)
	}
}

func TestDoctor_SyncCloud_Reachable(t *testing.T) {
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
	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "reachable") {
		t.Errorf("expected 'reachable' in output, got: %s", out)
	}
	if !strings.Contains(out, "sync: cloud") {
		t.Errorf("expected 'sync: cloud' in output, got: %s", out)
	}
}

func TestDoctor_SyncOnPrem_Unreachable(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	srv.Close() // close immediately

	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewOnPremSettings(srv.URL)); err != nil {
		t.Fatal(err)
	}
	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "unreachable") {
		t.Errorf("expected 'unreachable' in output, got: %s", out)
	}
	if !strings.Contains(out, "issue") {
		t.Errorf("expected 'issue' in output, got: %s", out)
	}
}

func TestDoctor_WithInvalidDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewNoneSettings()); err != nil {
		t.Fatal(err)
	}
	// Write a doc with bad naming (no type segment).
	d := filepath.Join(dir, ".archcore", "knowledge")
	os.MkdirAll(d, 0o755)
	os.WriteFile(filepath.Join(d, "readme.md"), []byte("---\ntitle: T\nstatus: draft\n---\n"), 0o644)

	out := runDoctorInDir(t, dir)
	if strings.Contains(out, "All checks passed") {
		t.Errorf("expected issues, but got 'All checks passed': %s", out)
	}
	if !strings.Contains(out, "issue") {
		t.Errorf("expected 'issue' in output, got: %s", out)
	}
}

// Test the underlying doctor flow directly for completeness.
func TestDoctor_AllChecksPass(t *testing.T) {
	srv := httptest.NewServer(healthyHandler())
	defer srv.Close()

	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(dir, config.NewOnPremSettings(srv.URL)); err != nil {
		t.Fatal(err)
	}

	// Use runInit to also verify integration.
	_, err := runInit(context.Background(), dir, config.NewOnPremSettings(srv.URL))
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}

	out := runDoctorInDir(t, dir)
	if !strings.Contains(out, "All checks passed") {
		t.Errorf("expected 'All checks passed', got: %s", out)
	}
}
