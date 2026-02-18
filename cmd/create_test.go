package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertFrontmatter validates the complete frontmatter block: opening/closing
// delimiters, title, and status fields.
func assertFrontmatter(t *testing.T, content, expectedTitle, expectedStatus string) {
	t.Helper()

	if !strings.HasPrefix(content, "---\n") {
		t.Error("missing frontmatter opening ---")
		return
	}

	rest := content[4:]
	frontmatter, _, found := strings.Cut(rest, "\n---\n")
	if !found {
		t.Error("missing frontmatter closing ---")
		return
	}
	if !strings.Contains(frontmatter, "title: "+expectedTitle) {
		t.Errorf("frontmatter: want title %q, got:\n%s", expectedTitle, frontmatter)
	}
	if !strings.Contains(frontmatter, "status: "+expectedStatus) {
		t.Errorf("frontmatter: want status %q, got:\n%s", expectedStatus, frontmatter)
	}
}

func TestFilenameToTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"oauth-tokens", "Oauth Tokens"},
		{"use-postgres", "Use Postgres"},
		{"simple", "Simple"},
		{"a-b-c", "A B C"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := filenameToTitle(tt.input)
			if got != tt.want {
				t.Errorf("filenameToTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestRunCreate validates every document type: correct category directory,
// frontmatter structure, default title/status, and that the right template
// body is rendered.
func TestRunCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		docType      string
		category     string
		bodyContains string // unique string from each type's template
	}{
		{"adr", "knowledge", "## Decision"},
		{"rfc", "knowledge", "## Motivation"},
		{"rule", "knowledge", "Rule as imperative statement"},
		{"guide", "knowledge", "## Steps"},
		{"doc", "knowledge", "## Content"},
		{"task-type", "experience", "## When to Use"},
		{"cpat", "experience", "### Classification"},
		{"prd", "vision", "### Product Vision Statement"},
		{"idea", "vision", "### Problem / Opportunity"},
		{"plan", "vision", "## Tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.docType, func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()
			err := runCreate(base, tt.docType, "test-doc", "", "")
			if err != nil {
				t.Fatalf("runCreate(%s): %v", tt.docType, err)
			}

			path := filepath.Join(base, ".archcore", tt.category, "test-doc."+tt.docType+".md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("file not at expected path: %v", err)
			}
			content := string(data)

			assertFrontmatter(t, content, "Test Doc", "draft")

			if !strings.Contains(content, tt.bodyContains) {
				t.Errorf("missing expected template body %q for type %s", tt.bodyContains, tt.docType)
			}
		})
	}
}

func TestRunCreate_TitleOverride(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	err := runCreate(base, "adr", "use-postgres", "Use PostgreSQL for Persistence", "")
	if err != nil {
		t.Fatalf("runCreate: %v", err)
	}

	path := filepath.Join(base, ".archcore", "knowledge", "use-postgres.adr.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	assertFrontmatter(t, string(data), "Use PostgreSQL for Persistence", "draft")
}

func TestRunCreate_StatusOverride(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	err := runCreate(base, "adr", "use-postgres", "", "accepted")
	if err != nil {
		t.Fatalf("runCreate: %v", err)
	}

	path := filepath.Join(base, ".archcore", "knowledge", "use-postgres.adr.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	assertFrontmatter(t, string(data), "Use Postgres", "accepted")
}

func TestRunCreate_PreventOverwrite(t *testing.T) {
	t.Parallel()
	base := t.TempDir()

	if err := runCreate(base, "rfc", "my-rfc", "", ""); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := runCreate(base, "rfc", "my-rfc", "", "")
	if err == nil {
		t.Fatal("expected error on duplicate create")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCreate_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		docType  string
		filename string
		errMsg   string
	}{
		{"invalid type", "invalid", "test", "invalid document type"},
		{"empty type", "", "test", "invalid document type"},
		{"project type excluded", "project", "test", "invalid document type"},
		{"empty filename", "rfc", "", "filename is required"},
		{"whitespace filename", "rfc", "   ", "filename is required"},
		{"path traversal slash", "rfc", "../etc/passwd", "must not contain path separators"},
		{"path traversal backslash", "rfc", "foo\\bar", "must not contain path separators"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()
			err := runCreate(base, tt.docType, tt.filename, "", "")
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestRunCreate_MkdirAllFailure(t *testing.T) {
	t.Parallel()
	base := t.TempDir()

	// Place a regular file where the directory needs to be created.
	blockPath := filepath.Join(base, ".archcore")
	if err := os.WriteFile(blockPath, []byte("block"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runCreate(base, "rfc", "test", "", "")
	if err == nil {
		t.Fatal("expected error when directory creation is blocked")
	}
	if !strings.Contains(err.Error(), "creating directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

// Cobra command integration tests below use os.Chdir (process-global),
// so they must NOT call t.Parallel().

func TestNewCreateCmd_TwoArgs(t *testing.T) {
	base := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	cmd := newCreateCmd()
	cmd.SetArgs([]string{"rfc", "cobra-test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	path := filepath.Join(base, ".archcore", "knowledge", "cobra-test.rfc.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	assertFrontmatter(t, content, "Cobra Test", "draft")
	if !strings.Contains(content, "## Motivation") {
		t.Error("missing RFC template body")
	}
}

func TestNewCreateCmd_WithFlags(t *testing.T) {
	base := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	cmd := newCreateCmd()
	cmd.SetArgs([]string{"adr", "flag-test", "--title", "Custom Title", "--status", "accepted"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	path := filepath.Join(base, ".archcore", "knowledge", "flag-test.adr.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	assertFrontmatter(t, string(data), "Custom Title", "accepted")
}

func TestNewCreateCmd_OneArgInvalidType(t *testing.T) {
	base := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	cmd := newCreateCmd()
	cmd.SetArgs([]string{"invalid"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid document type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCreateCmd_TooManyArgs(t *testing.T) {
	base := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	cmd := newCreateCmd()
	cmd.SetArgs([]string{"rfc", "name", "extra"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}
