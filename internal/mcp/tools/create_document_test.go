package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"archcore-cli/internal/sync"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleCreateDocument_Success(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "use-postgres",
		"title":    "Use PostgreSQL",
		"status":   "accepted",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	if info["path"] != ".archcore/use-postgres.adr.md" {
		t.Errorf("path = %q", info["path"])
	}
	if info["category"] != "knowledge" {
		t.Errorf("category = %q", info["category"])
	}

	// Verify file exists with correct content.
	data, err := os.ReadFile(filepath.Join(base, ".archcore", "use-postgres.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "Use PostgreSQL"`) {
		t.Error("missing title in frontmatter")
	}
	if !strings.Contains(content, "status: accepted") {
		t.Error("missing status in frontmatter")
	}
	if !strings.Contains(content, "## Decision") {
		t.Error("missing ADR template body")
	}
}

func TestHandleCreateDocument_DefaultTitleAndStatus(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "rfc",
		"filename": "oauth-tokens",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "oauth-tokens.rfc.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "Oauth Tokens"`) {
		t.Error("missing default title")
	}
	if !strings.Contains(content, "status: draft") {
		t.Error("missing default status")
	}
}

func TestHandleCreateDocument_CustomContent(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "custom",
		"content":  "## My Custom Content\n\nCustom body here.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "custom.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "## My Custom Content") {
		t.Error("missing custom content")
	}
	if strings.Contains(content, "## Decision") {
		t.Error("should not contain default template when custom content provided")
	}
}

func TestHandleCreateDocument_ContentWithFrontmatter(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	// Simulate an AI agent passing content that already includes frontmatter.
	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "frontmatter-dup",
		"title":    "Real Title",
		"status":   "accepted",
		"content":  "---\ntitle: Ignored Title\nstatus: draft\n---\n\n## Decision\nUse PostgreSQL.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "frontmatter-dup.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Must NOT contain duplicate frontmatter.
	if strings.Count(content, "---\n") > 2 {
		t.Errorf("duplicate frontmatter detected:\n%s", content)
	}
	// The title/status from the tool parameters should be used, not the ones in content.
	if !strings.Contains(content, `title: "Real Title"`) {
		t.Error("expected title from tool parameter, not from content frontmatter")
	}
	if !strings.Contains(content, "status: accepted") {
		t.Error("expected status from tool parameter, not from content frontmatter")
	}
	if !strings.Contains(content, "Use PostgreSQL.") {
		t.Error("body content missing")
	}
}

func TestHandleCreateDocument_DuplicatePrevented(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	_, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "dup-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "dup-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for duplicate file")
	}
}

func TestHandleCreateDocument_InvalidType(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "invalid",
		"filename": "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid type")
	}
}

func TestHandleCreateDocument_EmptyFilename(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for empty filename")
	}
}

func TestHandleCreateDocument_PathSeparator(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "../etc/passwd",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path separator in filename")
	}
}

func TestHandleCreateDocument_InvalidStatus(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "adr",
		"filename": "test",
		"status":   "proposed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid status")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "invalid status") {
		t.Errorf("error message = %q, want it to mention invalid status", text)
	}
}

func TestHandleCreateDocument_AllTypes(t *testing.T) {
	t.Parallel()

	types := []struct {
		docType      string
		category     string
		bodyContains string
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

	for _, tt := range types {
		t.Run(tt.docType, func(t *testing.T) {
			t.Parallel()
			base := setupTestArchcore(t)

			result, err := callTool(HandleCreateDocument(base), map[string]any{
				"type":     tt.docType,
				"filename": "test-doc",
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError {
				t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
			}

			// Without directory param, files go to .archcore/ root.
			path := filepath.Join(base, ".archcore", "test-doc."+tt.docType+".md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("file not at expected path: %v", err)
			}
			if !strings.Contains(string(data), tt.bodyContains) {
				t.Errorf("missing expected body %q for type %s", tt.bodyContains, tt.docType)
			}

			// Verify virtual category in response.
			var info map[string]any
			if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
				t.Fatal(err)
			}
			if info["category"] != tt.category {
				t.Errorf("category = %q, want %q", info["category"], tt.category)
			}
		})
	}
}

func TestHandleCreateDocument_WithDirectory(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "jwt-strategy",
		"title":     "JWT Strategy",
		"directory": "auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	if info["path"] != ".archcore/auth/jwt-strategy.adr.md" {
		t.Errorf("path = %q", info["path"])
	}
	if info["category"] != "knowledge" {
		t.Errorf("category = %q", info["category"])
	}

	// Verify file exists.
	if _, err := os.Stat(filepath.Join(base, ".archcore", "auth", "jwt-strategy.adr.md")); err != nil {
		t.Fatalf("file not found: %v", err)
	}
}

func TestHandleCreateDocument_DirectoryTraversal(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "evil",
		"directory": "../etc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for directory traversal")
	}
}

func TestHandleCreateDocument_NestedDirectory(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "k8s-migration",
		"title":     "K8s Migration",
		"directory": "infrastructure/k8s",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	if info["path"] != ".archcore/infrastructure/k8s/k8s-migration.adr.md" {
		t.Errorf("path = %q", info["path"])
	}

	// Verify file exists on disk.
	if _, err := os.Stat(filepath.Join(base, ".archcore", "infrastructure", "k8s", "k8s-migration.adr.md")); err != nil {
		t.Fatalf("file not found: %v", err)
	}
}

func TestHandleCreateDocument_AbsoluteDirectory(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "evil",
		"directory": "/etc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for absolute directory path")
	}
}

func TestHandleCreateDocument_DirectoryWithWhitespace(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "trimmed",
		"directory": "  auth  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	// Should be trimmed to "auth".
	if info["path"] != ".archcore/auth/trimmed.adr.md" {
		t.Errorf("path = %q, want .archcore/auth/trimmed.adr.md", info["path"])
	}
}

func TestHandleCreateDocument_NearbyDocuments(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	// Create an existing doc in the "auth" directory.
	writeDoc(t, base, "auth", "existing.adr.md", "---\ntitle: Existing\nstatus: draft\n---\n\nbody")

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "rule",
		"filename":  "new-rule",
		"directory": "auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	nearby, ok := info["nearby_documents"]
	if !ok {
		t.Fatal("expected nearby_documents in response")
	}
	arr, ok := nearby.([]any)
	if !ok || len(arr) == 0 {
		t.Fatal("expected non-empty nearby_documents array")
	}
	found := false
	for _, v := range arr {
		if v == ".archcore/auth/existing.adr.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("nearby_documents = %v, want to contain existing.adr.md", arr)
	}

	// Verify auto-created relations.
	autoRels, ok := info["auto_relations_added"]
	if !ok {
		t.Fatal("expected auto_relations_added in response")
	}
	relArr := autoRels.([]any)
	if len(relArr) != 1 || relArr[0] != ".archcore/auth/existing.adr.md" {
		t.Errorf("auto_relations_added = %v, want [.archcore/auth/existing.adr.md]", relArr)
	}

	// Verify relation persisted in manifest.
	m, err := sync.LoadManifest(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(m.Relations))
	}
	rel := m.Relations[0]
	if rel.Source != "auth/new-rule.rule.md" || rel.Target != "auth/existing.adr.md" || rel.Type != sync.RelRelated {
		t.Errorf("relation = %+v, want auth/new-rule.rule.md -> auth/existing.adr.md (related)", rel)
	}
}

func TestHandleCreateDocument_NearbyDocuments_Empty(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":      "adr",
		"filename":  "lonely",
		"directory": "empty-dir",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	if _, ok := info["nearby_documents"]; ok {
		t.Error("expected no nearby_documents when directory is empty")
	}
}

func TestFilenameToTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"oauth-tokens", "Oauth Tokens"},
		{"a-b-c", "A B C"},
		{"3-tier-arch", "3 Tier Arch"},
		{"single", "Single"},
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

func TestHandleCreateDocument_NearbyDocuments_SameRoot(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	// Create an existing doc at root.
	writeDoc(t, base, "", "root-doc.adr.md", "---\ntitle: Root\nstatus: draft\n---\n\nbody")

	result, err := callTool(HandleCreateDocument(base), map[string]any{
		"type":     "rule",
		"filename": "another-root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var info map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &info); err != nil {
		t.Fatal(err)
	}
	nearby, ok := info["nearby_documents"]
	if !ok {
		t.Fatal("expected nearby_documents for root-level doc")
	}
	arr := nearby.([]any)
	found := false
	for _, v := range arr {
		if v == ".archcore/root-doc.adr.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("nearby_documents = %v, want to contain root-doc.adr.md", arr)
	}
}
