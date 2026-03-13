package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func callTool(handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) (*mcp.CallToolResult, error) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return handler(context.Background(), req)
}

func TestHandleListDocuments_Empty(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	result, err := callTool(HandleListDocuments(base), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("unexpected error result")
	}

	var docs []LocalDocument
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &docs); err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("expected 0, got %d", len(docs))
	}
}

func TestHandleListDocuments_AllDocs(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "a.adr.md", "---\ntitle: A\nstatus: draft\n---\n")
	writeDoc(t, base, "vision", "b.prd.md", "---\ntitle: B\nstatus: accepted\n---\n")

	result, err := callTool(HandleListDocuments(base), nil)
	if err != nil {
		t.Fatal(err)
	}

	var docs []LocalDocument
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &docs); err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Errorf("expected 2, got %d", len(docs))
	}
}

func TestHandleListDocuments_FilterByType(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "a.adr.md", "---\ntitle: A\nstatus: draft\n---\n")
	writeDoc(t, base, "knowledge", "b.rfc.md", "---\ntitle: B\nstatus: draft\n---\n")

	result, err := callTool(HandleListDocuments(base), map[string]any{
		"types": []any{"adr"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var docs []LocalDocument
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &docs); err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1, got %d", len(docs))
	}
	if docs[0].Type != "adr" {
		t.Errorf("type = %q, want adr", docs[0].Type)
	}
}

func TestHandleListDocuments_FilterByCategory(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "a.adr.md", "---\ntitle: A\nstatus: draft\n---\n")
	writeDoc(t, base, "vision", "b.prd.md", "---\ntitle: B\nstatus: draft\n---\n")

	result, err := callTool(HandleListDocuments(base), map[string]any{
		"category": "vision",
	})
	if err != nil {
		t.Fatal(err)
	}

	var docs []LocalDocument
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &docs); err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1, got %d", len(docs))
	}
	if docs[0].Category != "vision" {
		t.Errorf("category = %q, want vision", docs[0].Category)
	}
}

func TestHandleListDocuments_FilterByStatus(t *testing.T) {
	t.Parallel()
	base := setupTestArchcore(t)
	writeDoc(t, base, "knowledge", "a.adr.md", "---\ntitle: A\nstatus: draft\n---\n")
	writeDoc(t, base, "knowledge", "b.rfc.md", "---\ntitle: B\nstatus: accepted\n---\n")

	result, err := callTool(HandleListDocuments(base), map[string]any{
		"status": "accepted",
	})
	if err != nil {
		t.Fatal(err)
	}

	var docs []LocalDocument
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &docs); err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1, got %d", len(docs))
	}
	if docs[0].Status != "accepted" {
		t.Errorf("status = %q, want accepted", docs[0].Status)
	}
}
