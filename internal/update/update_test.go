package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNeedsUpdate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"same version", "v1.0.0", "v1.0.0", false},
		{"patch update", "v1.0.0", "v1.0.1", true},
		{"minor update", "v1.0.0", "v1.1.0", true},
		{"major update", "v1.0.0", "v2.0.0", true},
		{"current newer patch", "v1.0.2", "v1.0.1", false},
		{"current newer minor", "v1.2.0", "v1.1.0", false},
		{"current newer major", "v2.0.0", "v1.9.9", false},
		{"dev always updates", "dev", "v1.0.0", true},
		{"dev with v prefix", "vdev", "v1.0.0", true},
		{"no v prefix current", "1.0.0", "v1.1.0", true},
		{"no v prefix latest", "v1.0.0", "1.1.0", true},
		{"no v prefix both", "1.0.0", "1.1.0", true},
		{"pre-release to release", "v1.0.0-alpha.1", "v1.0.0", true},
		{"pre-release major update", "v0.9.0-beta.1", "v1.0.0", true},
		{"same pre-release", "v0.0.1-alpha.7", "v0.0.1-alpha.7", false},
		{"pre-release bump", "v0.0.1-alpha.7", "v0.0.1-alpha.8", true},
		{"pre-release newer current", "v0.0.1-alpha.8", "v0.0.1-alpha.7", false},
		{"release to pre-release", "v1.0.0", "v1.0.0-alpha.1", false},
		{"alpha to beta", "v1.0.0-alpha.1", "v1.0.0-beta.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsUpdate(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("NeedsUpdate(%q, %q) = %v, want %v",
					tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCheckLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/archcore-ai/cli/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releaseResponse{TagName: "v1.2.3"})
	}))
	defer srv.Close()

	u := NewUpdater("v1.0.0", "archcore-ai/cli", "archcore")
	u.HTTPClient = srv.Client()

	// Override the GitHub API URL by wrapping the transport.
	u.HTTPClient.Transport = &rewriteTransport{
		base:   srv.Client().Transport,
		target: srv.URL,
	}

	got, err := u.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("CheckLatest() error: %v", err)
	}
	if got != "v1.2.3" {
		t.Errorf("CheckLatest() = %q, want %q", got, "v1.2.3")
	}
}

func TestCheckLatestError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"server error", http.StatusInternalServerError, "internal error"},
		{"not found", http.StatusNotFound, "not found"},
		{"rate limited", http.StatusForbidden, "rate limit exceeded"},
		{"empty tag_name", http.StatusOK, `{"tag_name": ""}`},
		{"malformed JSON", http.StatusOK, "not json at all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			u := NewUpdater("v1.0.0", "archcore-ai/cli", "archcore")
			u.HTTPClient = &http.Client{
				Transport: &rewriteTransport{
					base:   http.DefaultTransport,
					target: srv.URL,
				},
			}

			_, err := u.CheckLatest(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	validHash := fmt.Sprintf("%x", hash)

	tests := []struct {
		name      string
		data      []byte
		checksums string
		filename  string
		wantErr   bool
	}{
		{
			name:      "valid checksum",
			data:      data,
			checksums: fmt.Sprintf("%s  archcore_linux_amd64.tar.gz\n", validHash),
			filename:  "archcore_linux_amd64.tar.gz",
			wantErr:   false,
		},
		{
			name:      "valid with multiple entries",
			data:      data,
			checksums: fmt.Sprintf("abc123  archcore_darwin_arm64.tar.gz\n%s  archcore_linux_amd64.tar.gz\n", validHash),
			filename:  "archcore_linux_amd64.tar.gz",
			wantErr:   false,
		},
		{
			name:      "checksum mismatch",
			data:      data,
			checksums: "0000000000000000000000000000000000000000000000000000000000000000  archcore_linux_amd64.tar.gz\n",
			filename:  "archcore_linux_amd64.tar.gz",
			wantErr:   true,
		},
		{
			name:      "file not in checksums",
			data:      data,
			checksums: fmt.Sprintf("%s  archcore_darwin_arm64.tar.gz\n", validHash),
			filename:  "archcore_linux_amd64.tar.gz",
			wantErr:   true,
		},
		{
			name:      "empty checksums",
			data:      data,
			checksums: "",
			filename:  "archcore_linux_amd64.tar.gz",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyChecksum(tt.data, []byte(tt.checksums), tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractBinary(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho archcore")

	tests := []struct {
		name       string
		files      map[string][]byte // filename -> content
		candidates []string
		wantErr    bool
		wantData   []byte
	}{
		{
			name:       "extract by primary name",
			files:      map[string][]byte{"archcore": binaryContent},
			candidates: []string{"archcore", "cli"},
			wantErr:    false,
			wantData:   binaryContent,
		},
		{
			name:       "fallback to secondary name",
			files:      map[string][]byte{"cli": binaryContent},
			candidates: []string{"archcore", "cli"},
			wantErr:    false,
			wantData:   binaryContent,
		},
		{
			name:       "binary in subdirectory",
			files:      map[string][]byte{"archcore_v1.0.0_linux_amd64/archcore": binaryContent},
			candidates: []string{"archcore"},
			wantErr:    false,
			wantData:   binaryContent,
		},
		{
			name:       "binary not found",
			files:      map[string][]byte{"README.md": []byte("readme")},
			candidates: []string{"archcore", "cli"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := createTarGz(t, tt.files)

			got, err := ExtractBinary(archive, tt.candidates...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !bytes.Equal(got, tt.wantData) {
				t.Errorf("ExtractBinary() data mismatch: got %q, want %q", got, tt.wantData)
			}
		})
	}
}

func TestArchiveName(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		goos       string
		goarch     string
		want       string
	}{
		{"darwin arm64", "archcore", "darwin", "arm64", "archcore_darwin_arm64.tar.gz"},
		{"linux amd64", "archcore", "linux", "amd64", "archcore_linux_amd64.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArchiveName(tt.binaryName, tt.goos, tt.goarch)
			if got != tt.want {
				t.Errorf("ArchiveName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input   string
		want    []int
		wantPre string
	}{
		{"1.2.3", []int{1, 2, 3}, ""},
		{"0.0.1", []int{0, 0, 1}, ""},
		{"10.20.30", []int{10, 20, 30}, ""},
		{"1.2.3-alpha.1", []int{1, 2, 3}, "alpha.1"},
		{"0.0.1-beta.2", []int{0, 0, 1}, "beta.2"},
		{"invalid", nil, ""},
		{"1.2", nil, ""},
		{"1.2.x", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, gotPre := parseSemver(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseSemver(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseSemver(%q) = nil, want %v", tt.input, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("parseSemver(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
				}
			}
			if gotPre != tt.wantPre {
				t.Errorf("parseSemver(%q) pre-release = %q, want %q", tt.input, gotPre, tt.wantPre)
			}
		})
	}
}

func TestComparePreRelease(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "alpha.1", "alpha.1", 0},
		{"numeric bump", "alpha.7", "alpha.8", -1},
		{"numeric reverse", "alpha.8", "alpha.7", 1},
		{"alpha vs beta", "alpha.1", "beta.1", -1},
		{"numeric vs alpha", "1", "alpha", -1},
		{"shorter is less", "alpha", "alpha.1", -1},
		{"longer is more", "alpha.1", "alpha", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := comparePreRelease(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("comparePreRelease(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// createTarGz builds a tar.gz archive in memory from a map of filename -> content.
func createTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("writing tar header for %s: %v", name, err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("writing tar content for %s: %v", name, err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("closing gzip writer: %v", err)
	}

	return buf.Bytes()
}

// rewriteTransport rewrites all request URLs to point at a test server.
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	parsed, _ := url.Parse(t.target)
	req.URL.Scheme = parsed.Scheme
	req.URL.Host = parsed.Host
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

func TestAtomicReplace(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "binary")

	// Create initial file.
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("creating initial file: %v", err)
	}

	// Replace it.
	newData := []byte("new content")
	if err := atomicReplace(target, newData); err != nil {
		t.Fatalf("atomicReplace() error: %v", err)
	}

	// Verify content.
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading replaced file: %v", err)
	}
	if !bytes.Equal(got, newData) {
		t.Errorf("file content = %q, want %q", got, newData)
	}

	// Verify permissions.
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat replaced file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o755 {
		t.Errorf("file permissions = %o, want 755", perm)
	}
}

func TestAtomicReplace_NonexistentDir(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nonexistent", "binary")

	err := atomicReplace(target, []byte("data"))
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

func TestExtractBinary_CorruptArchive(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"not gzip", []byte("not a gzip archive")},
		{"truncated gzip header", []byte{0x1f, 0x8b}},
		{"empty input", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExtractBinary(tt.data, "archcore")
			if err == nil {
				t.Fatal("expected error for corrupt archive, got nil")
			}
		})
	}
}

func TestApply(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho archcore v2.0.0")
	archiveName := ArchiveName("archcore", runtime.GOOS, runtime.GOARCH)
	archiveData := createTarGz(t, map[string][]byte{"archcore": binaryContent})
	checksum := sha256.Sum256(archiveData)
	checksumLine := fmt.Sprintf("%x  %s\n", checksum, archiveName)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, archiveName):
			w.Write(archiveData)
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			w.Write([]byte(checksumLine))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "archcore")
	if err := os.WriteFile(fakeBinary, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("creating fake binary: %v", err)
	}

	u := &Updater{
		CurrentVersion: "v1.0.0",
		GitHubRepo:     "archcore-ai/cli",
		BinaryName:     "archcore",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{target: srv.URL},
		},
		ExecPath: fakeBinary,
	}

	if err := u.Apply(context.Background(), "v2.0.0"); err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	got, err := os.ReadFile(fakeBinary)
	if err != nil {
		t.Fatalf("reading replaced binary: %v", err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("binary content mismatch: got %d bytes, want %d bytes", len(got), len(binaryContent))
	}

	info, err := os.Stat(fakeBinary)
	if err != nil {
		t.Fatalf("stat binary: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o755 {
		t.Errorf("binary permissions = %o, want 755", perm)
	}
}

func TestApply_ChecksumMismatch(t *testing.T) {
	archiveName := ArchiveName("archcore", runtime.GOOS, runtime.GOARCH)
	archiveData := createTarGz(t, map[string][]byte{"archcore": []byte("binary")})
	badChecksumLine := fmt.Sprintf("%s  %s\n",
		"0000000000000000000000000000000000000000000000000000000000000000", archiveName)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, archiveName):
			w.Write(archiveData)
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			w.Write([]byte(badChecksumLine))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "archcore")
	os.WriteFile(fakeBinary, []byte("old"), 0o755)

	u := &Updater{
		CurrentVersion: "v1.0.0",
		GitHubRepo:     "archcore-ai/cli",
		BinaryName:     "archcore",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{target: srv.URL},
		},
		ExecPath: fakeBinary,
	}

	err := u.Apply(context.Background(), "v2.0.0")
	if err == nil {
		t.Fatal("expected checksum mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected checksum mismatch error, got: %v", err)
	}
}

func TestApply_DownloadFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "archcore")
	os.WriteFile(fakeBinary, []byte("old"), 0o755)

	u := &Updater{
		CurrentVersion: "v1.0.0",
		GitHubRepo:     "archcore-ai/cli",
		BinaryName:     "archcore",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{target: srv.URL},
		},
		ExecPath: fakeBinary,
	}

	err := u.Apply(context.Background(), "v2.0.0")
	if err == nil {
		t.Fatal("expected download error, got nil")
	}
	if !strings.Contains(err.Error(), "downloading") {
		t.Errorf("expected downloading error, got: %v", err)
	}
}

func TestDownload_SizeLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := bytes.Repeat([]byte("a"), maxChecksumsSize+1)
		w.Write(data)
	}))
	defer srv.Close()

	u := &Updater{
		GitHubRepo: "test/repo",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{target: srv.URL},
		},
	}

	_, err := u.download(context.Background(), "v1.0.0", "checksums.txt")
	if err == nil {
		t.Fatal("expected size limit error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds size limit") {
		t.Errorf("expected size limit error, got: %v", err)
	}
}
