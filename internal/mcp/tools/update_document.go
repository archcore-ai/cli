package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"archcore-cli/templates"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewUpdateDocumentTool returns the tool definition for update_document.
func NewUpdateDocumentTool() mcp.Tool {
	return mcp.NewTool("update_document",
		mcp.WithDescription(`Update an existing document in the .archcore/ knowledge base.

Before updating, confirm the intended changes with the user if possible.

Call list_documents first to get the document's path, then pass that path to get_document to review current content before updating.

You can update any combination of: title (frontmatter), status (frontmatter), content (markdown body). Fields not provided are left unchanged.

Returns: JSON with the path of the updated file, its type, category, title, and status.`),
		mcp.WithString("path",
			mcp.Description("Relative path to the document from the project root. Must be obtained from list_documents — do not construct this manually. Example: \".archcore/knowledge/use-postgres.adr.md\""),
			mcp.Required(),
		),
		mcp.WithString("title",
			mcp.Description("New title for the document frontmatter. If omitted, the existing title is preserved."),
		),
		mcp.WithString("status",
			mcp.Description("New document status. Valid values: draft, accepted, rejected."),
			mcp.Enum("draft", "accepted", "rejected"),
		),
		mcp.WithString("content",
			mcp.Description("New markdown body for the document. Replaces everything after the frontmatter. If omitted, the existing body is preserved."),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Update Document",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(false),
		}),
	)
}

// HandleUpdateDocument handles the update_document tool call.
func HandleUpdateDocument(baseDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		relPath, err := request.RequireString("path")
		if err != nil {
			return errorResult(err.Error()), nil
		}

		// Validate path.
		relPath, errMsg := validateArchcorePath(relPath)
		if errMsg != "" {
			return errorResult(errMsg), nil
		}

		// Require at least one update field.
		newTitle := request.GetString("title", "")
		newStatus := request.GetString("status", "")
		newContent := request.GetString("content", "")
		if newTitle == "" && newStatus == "" && newContent == "" {
			return errorResult("at least one of title, status, or content must be provided"), nil
		}

		if newStatus != "" && !templates.IsValidStatus(newStatus) {
			return errorResult(fmt.Sprintf("invalid status %q (valid: %s)", newStatus, strings.Join(templates.ValidStatuses(), ", "))), nil
		}

		// Read existing file.
		absPath := filepath.Join(baseDir, relPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			return errorResult(fmt.Sprintf("document not found: %s", relPath)), nil
		}

		// Parse existing document.
		existingTitle, existingStatus, existingBody := templates.SplitDocument(data)

		// Apply updates.
		title := existingTitle
		if newTitle != "" {
			title = newTitle
		}
		status := existingStatus
		if newStatus != "" {
			status = newStatus
		}
		body := existingBody
		if newContent != "" {
			body = stripFrontmatter(newContent)
		}

		// Reconstruct the file.
		fileContent := buildDocumentFile(title, status, body)

		if err := os.WriteFile(absPath, []byte(fileContent), 0o644); err != nil {
			return errorResult(fmt.Sprintf("writing file: %v", err)), nil
		}

		// Derive category from document type, not directory.
		filename := filepath.Base(relPath)
		docType := templates.ExtractDocType(filename)
		category := templates.CategoryForType(templates.DocumentType(docType))

		result := map[string]any{
			"path":     relPath,
			"category": category,
			"type":     docType,
			"title":    title,
			"status":   status,
		}
		jsonData, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("marshaling result: %v", err)), nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(string(jsonData))},
		}, nil
	}
}
