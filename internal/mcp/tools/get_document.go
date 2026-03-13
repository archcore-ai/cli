package tools

import (
	"context"
	"encoding/json"

	"archcore-cli/internal/sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewGetDocumentTool returns the tool definition for get_document.
func NewGetDocumentTool() mcp.Tool {
	return mcp.NewTool("get_document",
		mcp.WithDescription(`Read the full content of a single .archcore/ document by its file path.

Call this tool AFTER list_documents has returned a valid path. Do not guess or construct paths — only use paths returned by list_documents.

Returns: the document's YAML frontmatter (title, type, status), its full Markdown body, and any outgoing_relations and incoming_relations from the knowledge graph.

Use this tool when you need to:
- Read the reasoning or content of a specific document
- Verify what a document says before creating a related one
- Retrieve a document to summarize or reference in a response`),
		mcp.WithString("path",
			mcp.Description(`Relative path to the document from the project root. Must be obtained from list_documents — do not construct this manually. Example: ".archcore/knowledge/use-postgres.adr.md"`),
			mcp.Required(),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Document",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

// HandleGetDocument handles the get_document tool call.
func HandleGetDocument(baseDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return errorResult(err.Error()), nil
		}

		// Validate path safety.
		path, errMsg := validateArchcorePath(path)
		if errMsg != "" {
			return errorResult(errMsg), nil
		}

		doc, err := ReadDocumentContent(baseDir, path)
		if err != nil {
			return errorResult("document not found: " + path), nil
		}

		enriched := EnrichedDocument{LocalDocument: doc}

		// Try to load relations from manifest.
		relPath := normalizeRelPath(path)
		if m, mErr := sync.LoadManifest(baseDir); mErr == nil {
			outgoing, incoming := m.RelationsFor(relPath)
			for _, r := range outgoing {
				enriched.OutgoingRelations = append(enriched.OutgoingRelations, DocumentRelation{
					Path: ".archcore/" + r.Target,
					Type: string(r.Type),
				})
			}
			for _, r := range incoming {
				enriched.IncomingRelations = append(enriched.IncomingRelations, DocumentRelation{
					Path: ".archcore/" + r.Source,
					Type: string(r.Type),
				})
			}
		}

		data, err := json.Marshal(enriched)
		if err != nil {
			return errorResult("marshaling result: " + err.Error()), nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(string(data))},
		}, nil
	}
}
