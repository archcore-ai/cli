package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewListDocumentsTool returns the tool definition for list_documents.
func NewListDocumentsTool() mcp.Tool {
	return mcp.NewTool("list_documents",
		mcp.WithDescription(`Discover and filter documents in the .archcore/ knowledge base.

Call this tool FIRST before reading or creating any document. Use it to:
- Check whether a document on a given topic already exists (prevents duplicates)
- Get valid file paths required by get_document
- Browse what documentation is available by type, category, or status

Returns: a JSON array of documents, each with path, title, type, category, and status. Returns an empty array if no documents match.

Use the returned paths directly as input to get_document. Do not construct paths manually.`),
		mcp.WithArray("types",
			mcp.Description("Filter by one or more document types. Valid values: adr, rfc, rule, guide, doc, prd, idea, plan, task-type, cpat. Example: [\"adr\", \"rule\"] returns only decision records and standards."),
			mcp.WithStringItems(),
		),
		mcp.WithString("category",
			mcp.Description(`Filter by virtual category (derived from document type, not directory). Use "knowledge" for decisions/standards/guides/docs/proposals, "vision" for requirements/ideas/plans, "experience" for task patterns and code pattern changes.`),
			mcp.Enum("vision", "knowledge", "experience"),
		),
		mcp.WithString("status",
			mcp.Description("Filter by frontmatter status field. Valid values: draft, accepted, rejected."),
			mcp.Enum("draft", "accepted", "rejected"),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Documents",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

// HandleListDocuments handles the list_documents tool call.
func HandleListDocuments(baseDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		docs, err := ScanDocuments(baseDir)
		if err != nil {
			return errorResult("scanning documents: " + err.Error()), nil
		}

		types := request.GetStringSlice("types", nil)
		category := request.GetString("category", "")
		status := request.GetString("status", "")

		var filtered []LocalDocument
		for _, doc := range docs {
			if len(types) > 0 && !containsStr(types, doc.Type) {
				continue
			}
			if category != "" && doc.Category != category {
				continue
			}
			if status != "" && !strings.EqualFold(doc.Status, status) {
				continue
			}
			filtered = append(filtered, doc)
		}

		data, err := json.Marshal(filtered)
		if err != nil {
			return errorResult("marshaling result: " + err.Error()), nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(string(data))},
		}, nil
	}
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func errorResult(msg string) *mcp.CallToolResult {
	return mcp.NewToolResultError(msg)
}
