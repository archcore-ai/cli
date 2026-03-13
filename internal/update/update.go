package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	maxArchiveSize   = 50 << 20 // 50 MB
	maxChecksumsSize = 1 << 20  // 1 MB
)

// Updater checks for and applies updates from GitHub Releases.
type Updater struct {
	CurrentVersion string
	GitHubRepo     string // e.g. "archcore-ai/cli"
	BinaryName     string // e.g. "archcore"
	HTTPClient     *http.Client
	ExecPath       string // Override for os.Executable(); used in tests.
}

// NewUpdater creates an Updater with sensible defaults.
func NewUpdater(currentVersion, repo, binaryName string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
		GitHubRepo:     repo,
		BinaryName:     binaryName,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// releaseResponse holds the relevant fields from the GitHub Releases API.
type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// CheckLatest queries the GitHub API for the latest release tag.
func (u *Updater) CheckLatest(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.GitHubRepo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("checking latest release: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxChecksumsSize)).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing release response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in release response")
	}

	return release.TagName, nil
}

// NeedsUpdate compares current and latest versions.
// Returns true if latest is newer than current or if current is "dev".
func NeedsUpdate(current, latest string) bool {
	current = stripV(current)
	latest = stripV(latest)

	if current == "dev" {
		return true
	}

	curParts, curPre := parseSemver(current)
	latParts, latPre := parseSemver(latest)

	if curParts == nil || latParts == nil {
		// Fall back to string comparison if parsing fails.
		return current != latest
	}

	for i := range 3 {
		if latParts[i] > curParts[i] {
			return true
		}
		if latParts[i] < curParts[i] {
			return false
		}
	}

	// Major.minor.patch are equal — compare pre-release.
	// Per SemVer: release (no pre-release) > any pre-release.
	if curPre != "" && latPre == "" {
		return true // current is pre-release, latest is release
	}
	if curPre == "" && latPre != "" {
		return false // current is release, latest is pre-release
	}
	if curPre == "" && latPre == "" {
		return false // both are releases, identical
	}

	return comparePreRelease(curPre, latPre) < 0
}

// stripV removes the leading "v" prefix from a version string.
func stripV(v string) string {
	return strings.TrimPrefix(v, "v")
}

// parseSemver splits a version string like "1.2.3" or "1.2.3-beta.1" into
// [major, minor, patch] and an optional pre-release string.
// Returns nil if parsing fails.
func parseSemver(v string) ([]int, string) {
	var preRelease string
	if idx := strings.Index(v, "-"); idx != -1 {
		preRelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil, ""
	}

	result := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, ""
		}
		result[i] = n
	}
	return result, preRelease
}

// comparePreRelease compares two pre-release strings per SemVer 2.0.0 §11.
// Returns -1 if a < b, 0 if a == b, +1 if a > b.
func comparePreRelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	n := len(aParts)
	if len(bParts) < n {
		n = len(bParts)
	}

	for i := range n {
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])

		switch {
		case aErr == nil && bErr == nil:
			// Both numeric: compare as integers.
			if aNum < bNum {
				return -1
			}
			if aNum > bNum {
				return 1
			}
		case aErr == nil && bErr != nil:
			// Numeric < alphanumeric per SemVer.
			return -1
		case aErr != nil && bErr == nil:
			return 1
		default:
			// Both alphanumeric: compare lexically.
			if aParts[i] < bParts[i] {
				return -1
			}
			if aParts[i] > bParts[i] {
				return 1
			}
		}
	}

	// All compared identifiers are equal — shorter set has lower precedence.
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}
	return 0
}

// ArchiveName returns the expected archive filename for the current platform.
func ArchiveName(binaryName, goos, goarch string) string {
	return fmt.Sprintf("%s_%s_%s.tar.gz", binaryName, goos, goarch)
}

// Apply downloads and installs the specified version.
func (u *Updater) Apply(ctx context.Context, version string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	archive := ArchiveName(u.BinaryName, goos, goarch)

	// 1. Download the archive.
	archiveData, err := u.download(ctx, version, archive)
	if err != nil {
		return fmt.Errorf("downloading archive: %w", err)
	}

	// 2. Download and verify checksum.
	checksums, err := u.download(ctx, version, "checksums.txt")
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	if err := VerifyChecksum(archiveData, checksums, archive); err != nil {
		return err
	}

	// 3. Extract binary from the archive.
	// GoReleaser may name the binary either as the configured binary name
	// or as the repo basename. Try the binary name first, fall back to repo basename.
	repoBasename := filepath.Base(u.GitHubRepo)
	binaryData, err := ExtractBinary(archiveData, u.BinaryName, repoBasename)
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// 4. Find current binary path (resolve symlinks).
	execPath := u.ExecPath
	if execPath == "" {
		execPath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("locating current binary: %w", err)
		}
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	// 5. Atomic replace.
	if err := atomicReplace(execPath, binaryData); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

// download fetches a file from a GitHub release.
func (u *Updater) download(ctx context.Context, version, filename string) ([]byte, error) {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		u.GitHubRepo, version, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", filename, resp.StatusCode)
	}

	limit := int64(maxArchiveSize)
	if strings.HasSuffix(filename, ".txt") {
		limit = maxChecksumsSize
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s exceeds size limit (%d bytes)", filename, limit)
	}

	return data, nil
}

// VerifyChecksum validates the SHA-256 checksum of data against the
// checksums file content. The checksums file is expected to have lines
// in the format: "<hash>  <filename>".
func VerifyChecksum(data, checksums []byte, filename string) error {
	expected, err := findChecksum(checksums, filename)
	if err != nil {
		return err
	}

	actual := sha256sum(data)
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filename, expected, actual)
	}

	return nil
}

// findChecksum looks up a filename in checksums.txt content and returns
// the corresponding SHA-256 hash.
func findChecksum(checksums []byte, filename string) (string, error) {
	for line := range strings.SplitSeq(string(checksums), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "<hash>  <filename>" (two spaces between hash and name).
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == filename {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", filename)
}

// sha256sum computes the hex-encoded SHA-256 digest of data.
func sha256sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ExtractBinary extracts a binary from a tar.gz archive.
// It tries each candidate name in order, returning the first match.
func ExtractBinary(archiveData []byte, candidates ...string) ([]byte, error) {
	candidateSet := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		candidateSet[c] = true
	}

	gr, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("opening gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		// The binary may be at the root or inside a directory.
		name := filepath.Base(hdr.Name)
		if hdr.Typeflag == tar.TypeReg && candidateSet[name] {
			data, err := io.ReadAll(io.LimitReader(tr, maxArchiveSize))
			if err != nil {
				return nil, fmt.Errorf("reading %s from archive: %w", name, err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("binary not found in archive (tried: %s)", strings.Join(candidates, ", "))
}

// atomicReplace writes data to a temporary file next to target, then
// renames it over target for an atomic update.
func atomicReplace(target string, data []byte) error {
	dir := filepath.Dir(target)
	tmpPath := filepath.Join(dir, fmt.Sprintf("%s.tmp.%d", filepath.Base(target), os.Getpid()))

	if err := os.WriteFile(tmpPath, data, 0o755); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, target); err != nil {
		// Clean up the temporary file on failure.
		os.Remove(tmpPath)
		return err
	}

	return nil
}
