package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"archcore-cli/internal/sync"
	"archcore-cli/templates"

	"github.com/mark3labs/mcp-go/mcp"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// NewCreateDocumentTool returns the tool definition for create_document.
func NewCreateDocumentTool() mcp.Tool {
	return mcp.NewTool("create_document",
		mcp.WithDescription(`Create a new structured document in the .archcore/ knowledge base.

BEFORE calling this tool: call list_documents to confirm no equivalent document already exists. Do not create duplicates.

Use this tool to permanently capture distilled, high-value knowledge — not temporary notes, chat logs, or speculative content.

Document types and when to use each:
  adr       — A decision that has been made, with context and consequences
                § required sections: Context, Decision, Alternatives Considered, Consequences
  rfc       — A proposal open for team review before a decision is made
                § required sections: Summary, Motivation, Detailed Design, Drawbacks, Alternatives
  rule      — A mandatory team standard or required behavior
                § required sections: Rule (imperative statements), Rationale, Examples (Good/Bad), Enforcement
  guide     — Step-by-step instructions for completing a task
                § required sections: Prerequisites, Steps (numbered), Verification, Common Issues
  doc       — General reference documentation (use when no other type fits)
                § required sections: Overview, Content (sections/tables), Examples
  prd       — Product requirements with goals, scope, and acceptance criteria
                § required sections: Vision, Problem Statement, Goals and Success Metrics, Requirements
  idea      — A product or technical concept worth exploring
                § required sections: Idea, Value, Possible Implementation, Risks and Constraints
  plan      — A concrete implementation plan with defined tasks
                § required sections: Goal, Tasks (phased), Acceptance Criteria, Dependencies
  task-type — A proven pattern for a typical recurring implementation task
                § required sections: What, When to Use, Steps, Example, Things to Watch Out For
  cpat      — A code pattern change: documents how and why a convention or approach changed
                § required sections: What Changed, Why, Before, After, Scope

TYPE DISAMBIGUATION:
- rule vs doc: rule prescribes behavior ("Always do X") with good/bad examples and enforcement. doc describes what exists (tables, registries, explanations). Descriptive content → doc.
- adr vs rfc: adr = decision already final. rfc = proposal open for feedback.
- guide vs doc: guide = sequential steps to follow. doc = non-sequential reference to look things up.

Returns: JSON with path, type, category, title, status, and optionally nearby_documents — paths of other documents in the same directory that may warrant adding a relation.`),
		mcp.WithString("type",
			mcp.Description("Document type. Choose based on the nature of the content, not the topic. If uncertain between adr and rfc: use adr only if the decision is already final."),
			mcp.Required(),
			mcp.Enum(templates.ValidTypes()...),
		),
		mcp.WithString("filename",
			mcp.Description("URL-safe slug for the filename. Use lowercase letters and hyphens only — no spaces, underscores, or special characters. Do not include the file extension. Example: \"use-postgres\", \"rate-limiting-strategy\"."),
			mcp.Required(),
		),
		mcp.WithString("title",
			mcp.Description("Human-readable document title. Write as a short descriptive phrase, not a slug. Example: \"Use PostgreSQL for primary persistence\". If omitted, derived from filename."),
		),
		mcp.WithString("status",
			mcp.Description("Initial document status. Valid values: draft, accepted, rejected. Defaults to \"draft\"."),
			mcp.Enum("draft", "accepted", "rejected"),
		),
		mcp.WithString("content",
			mcp.Description(`Markdown body of the document. RECOMMENDED: omit this parameter to get the standard template for the document type — it contains all required sections with guidance placeholders.

If you provide content, you MUST include the required sections for the chosen type (see § required sections above). Do not invent a structure — follow the template's section layout. Do not include a top-level heading — the title is stored in frontmatter separately.`),
		),
		mcp.WithString("directory",
			mcp.Description(`Optional subdirectory inside .archcore/ where the file should be created. Use to organize documents by domain, feature, or team (e.g. "auth", "payments", "infrastructure/k8s"). If omitted, the file is created in the .archcore/ root. Must not contain ".." or start with "/".`),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Create Document",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(false),
		}),
	)
}

// HandleCreateDocument handles the create_document tool call.
func HandleCreateDocument(baseDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		docType, err := request.RequireString("type")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		filename, err := request.RequireString("filename")
		if err != nil {
			return errorResult(err.Error()), nil
		}

		if !templates.IsValidType(docType) {
			return errorResult(fmt.Sprintf("invalid document type %q (valid: %s)", docType, strings.Join(templates.ValidTypes(), ", "))), nil
		}

		filename = strings.TrimSpace(filename)
		if filename == "" {
			return errorResult("filename is required"), nil
		}
		if strings.ContainsAny(filename, "/\\") {
			return errorResult(fmt.Sprintf("invalid filename %q: must not contain path separators", filename)), nil
		}
		if !slugRe.MatchString(filename) {
			return errorResult(fmt.Sprintf("invalid filename %q: must be lowercase alphanumeric with hyphens (e.g. \"use-postgres\")", filename)), nil
		}

		title := request.GetString("title", "")
		if title == "" {
			title = filenameToTitle(filename)
		}

		status := request.GetString("status", "draft")
		if !templates.IsValidStatus(status) {
			return errorResult(fmt.Sprintf("invalid status %q (valid: %s)", status, strings.Join(templates.ValidStatuses(), ", "))), nil
		}
		content := request.GetString("content", "")
		directory := request.GetString("directory", "")

		// Validate directory parameter.
		if directory != "" {
			directory = strings.TrimSpace(directory)
			cleaned := filepath.Clean(directory)
			if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
				return errorResult(fmt.Sprintf("invalid directory %q: must be relative and within .archcore/", directory)), nil
			}
			directory = filepath.ToSlash(cleaned)
		}

		category := templates.CategoryForType(templates.DocumentType(docType))

		// Build target directory: .archcore/<directory>/ or .archcore/ root.
		var dir string
		if directory != "" {
			dir = filepath.Join(baseDir, ".archcore", directory)
		} else {
			dir = filepath.Join(baseDir, ".archcore")
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return errorResult(fmt.Sprintf("creating directory %q: %v", directory, err)), nil
		}

		outputFile := filepath.Join(dir, filename+"."+docType+".md")

		relPath, err := filepath.Rel(baseDir, outputFile)
		if err != nil {
			relPath = filepath.ToSlash(outputFile)
		}
		relPath = filepath.ToSlash(relPath)

		if _, err := os.Stat(outputFile); err == nil {
			return errorResult(fmt.Sprintf("file already exists: %s", relPath)), nil
		}

		body := content
		if body == "" {
			body = templates.GenerateTemplate(templates.DocumentType(docType))
		} else {
			body = stripFrontmatter(body)
		}

		fileContent := buildDocumentFile(title, status, body)

		if err := os.WriteFile(outputFile, []byte(fileContent), 0o644); err != nil {
			return errorResult(fmt.Sprintf("writing %s: failed to write file", relPath)), nil
		}

		result := map[string]any{
			"path":     relPath,
			"category": category,
			"type":     docType,
			"title":    title,
			"status":   status,
		}

		// Scan for nearby documents in the same directory and auto-create relations.
		if allDocs, scanErr := ScanDocuments(baseDir); scanErr == nil {
			createdDir := filepath.Dir(relPath)
			var nearby []string
			for _, d := range allDocs {
				if d.Path != relPath && filepath.Dir(d.Path) == createdDir {
					nearby = append(nearby, d.Path)
				}
			}
			if len(nearby) > 0 {
				result["nearby_documents"] = nearby

				// Auto-add "related" relations between the new doc and nearby docs.
				if m, loadErr := sync.LoadManifest(baseDir); loadErr == nil {
					// relPath is ".archcore/dir/file.md", normalize to "dir/file.md"
					newDocRel := strings.TrimPrefix(relPath, ".archcore/")
					var added []string
					for _, np := range nearby {
						nearbyRel := strings.TrimPrefix(np, ".archcore/")
						if m.AddRelation(newDocRel, nearbyRel, sync.RelRelated) {
							added = append(added, np)
						}
					}
					if len(added) > 0 {
						if saveErr := sync.SaveManifest(baseDir, m); saveErr == nil {
							result["auto_relations_added"] = added
						}
					}
				}
			}
		}

		data, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("marshaling result: %v", err)), nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(string(data))},
		}, nil
	}
}

// filenameToTitle converts a slug like "oauth-tokens" to "Oauth Tokens".
func filenameToTitle(filename string) string {
	parts := strings.Split(filename, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
