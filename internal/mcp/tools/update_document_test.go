package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

const testDoc = "---\ntitle: Original Title\nstatus: draft\n---\n\n## Context\nOriginal body."

func TestHandleUpdateDocument_TitleOnly(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":  ".archcore/knowledge/my-adr.adr.md",
		"title": "New Title",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "knowledge", "my-adr.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "New Title"`) {
		t.Error("title not updated")
	}
	if !strings.Contains(content, "status: draft") {
		t.Error("status should be preserved")
	}
	if !strings.Contains(content, "Original body.") {
		t.Error("body should be preserved")
	}
}

func TestHandleUpdateDocument_StatusOnly(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   ".archcore/knowledge/my-adr.adr.md",
		"status": "accepted",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "knowledge", "my-adr.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "Original Title"`) {
		t.Error("title should be preserved")
	}
	if !strings.Contains(content, "status: accepted") {
		t.Error("status not updated")
	}
	if !strings.Contains(content, "Original body.") {
		t.Error("body should be preserved")
	}
}

func TestHandleUpdateDocument_ContentOnly(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":    ".archcore/knowledge/my-adr.adr.md",
		"content": "## Updated\nNew body here.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "knowledge", "my-adr.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "Original Title"`) {
		t.Error("title should be preserved")
	}
	if !strings.Contains(content, "status: draft") {
		t.Error("status should be preserved")
	}
	if !strings.Contains(content, "New body here.") {
		t.Error("content not updated")
	}
	if strings.Contains(content, "Original body.") {
		t.Error("old body should be replaced")
	}
}

func TestHandleUpdateDocument_AllFields(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":    ".archcore/knowledge/my-adr.adr.md",
		"title":   "All New",
		"status":  "accepted",
		"content": "Completely new content.",
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
	if info["title"] != "All New" {
		t.Errorf("title = %q, want %q", info["title"], "All New")
	}
	if info["status"] != "accepted" {
		t.Errorf("status = %q, want %q", info["status"], "accepted")
	}
	if info["category"] != "knowledge" {
		t.Errorf("category = %q, want %q", info["category"], "knowledge")
	}
	if info["type"] != "adr" {
		t.Errorf("type = %q, want %q", info["type"], "adr")
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "knowledge", "my-adr.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "All New"`) {
		t.Error("title not updated in file")
	}
	if !strings.Contains(content, "status: accepted") {
		t.Error("status not updated in file")
	}
	if !strings.Contains(content, "Completely new content.") {
		t.Error("content not updated in file")
	}
}

func TestHandleUpdateDocument_InvalidStatus(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   ".archcore/knowledge/my-adr.adr.md",
		"status": "proposed",
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

func TestHandleUpdateDocument_PathTraversal(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   "../etc/passwd",
		"status": "hacked",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for path traversal")
	}
}

func TestHandleUpdateDocument_NonArchcorePath(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   "src/main.go",
		"status": "draft",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for non-.archcore path")
	}
}

func TestHandleUpdateDocument_MissingFile(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   ".archcore/knowledge/nonexistent.adr.md",
		"status": "accepted",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing file")
	}
}

func TestHandleUpdateDocument_NoFieldsProvided(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path": ".archcore/knowledge/my-adr.adr.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error when no update fields provided")
	}
}

func TestHandleUpdateDocument_ContentWithFrontmatter(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "my-adr.adr.md", testDoc)

	// Simulate an AI agent passing content that already includes frontmatter.
	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":    ".archcore/knowledge/my-adr.adr.md",
		"content": "---\ntitle: Ignored Title\nstatus: accepted\n---\n\n## Updated\nNew body here.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	data, err := os.ReadFile(filepath.Join(base, ".archcore", "knowledge", "my-adr.adr.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Must NOT contain duplicate frontmatter.
	if strings.Count(content, "---\n") > 2 {
		t.Errorf("duplicate frontmatter detected:\n%s", content)
	}
	// Original title/status should be preserved since only content was updated.
	if !strings.Contains(content, `title: "Original Title"`) {
		t.Error("title should be preserved when only content is updated")
	}
	if !strings.Contains(content, "status: draft") {
		t.Error("status should be preserved when only content is updated")
	}
	if !strings.Contains(content, "New body here.") {
		t.Error("body content not updated")
	}
}

func TestHandleUpdateDocument_CategoryDerivedFromType(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	// Place a PRD (vision type) in a custom "auth" directory — category must be "vision", not "auth".
	writeDoc(t, base, "auth", "oauth.prd.md", "---\ntitle: OAuth PRD\nstatus: draft\n---\n\nBody.")

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":  ".archcore/auth/oauth.prd.md",
		"title": "Updated OAuth PRD",
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
	if info["category"] != "vision" {
		t.Errorf("category = %q, want %q (should be derived from type, not directory)", info["category"], "vision")
	}
	if info["type"] != "prd" {
		t.Errorf("type = %q, want %q", info["type"], "prd")
	}
}

func TestHandleUpdateDocument_RootLevelDoc(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)

	// Doc at .archcore/ root (no subdirectory).
	writeDoc(t, base, "", "my-idea.idea.md", "---\ntitle: My Idea\nstatus: draft\n---\n\nBody.")

	result, err := callTool(HandleUpdateDocument(base), map[string]any{
		"path":   ".archcore/my-idea.idea.md",
		"status": "accepted",
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
	if info["category"] != "vision" {
		t.Errorf("category = %q, want %q", info["category"], "vision")
	}
	if info["status"] != "accepted" {
		t.Errorf("status = %q, want %q", info["status"], "accepted")
	}
}
