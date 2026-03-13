package cmd

import (
	"fmt"
	"strings"

	"archcore-cli/internal/mcp/tools"
	"archcore-cli/internal/sync"
)

// buildSessionContext generates the session-start context string
// that is injected into agents at session start.
func buildSessionContext(baseDir string) (string, int) {
	docs, err := tools.ScanDocuments(baseDir)
	if err != nil {
		docs = nil // proceed with empty list on error
	}

	var b strings.Builder
	b.WriteString("[Archcore — System Context Platform]\n")
	b.WriteString("Keeps humans and AI in sync with your system.\n")
	b.WriteString("You have MCP tools available: list_documents, get_document, create_document, update_document, add_relation, remove_relation, list_relations.\n")

	// Pre-group documents by category.
	docsByCategory := make(map[string][]tools.LocalDocument, 3)
	for _, doc := range docs {
		docsByCategory[doc.Category] = append(docsByCategory[doc.Category], doc)
	}

	// Existing documents by category.
	b.WriteString("\nEXISTING DOCUMENTS:\n")
	for _, cat := range []string{"knowledge", "vision", "experience"} {
		fmt.Fprintf(&b, "  [%s]\n", cat)
		catDocs := docsByCategory[cat]
		if len(catDocs) == 0 {
			b.WriteString("    (none)\n")
			continue
		}
		for _, doc := range catDocs {
			titlePart := ""
			if doc.Title != "" {
				titlePart = fmt.Sprintf(" — %q", doc.Title)
			}
			fmt.Fprintf(&b, "    - %s%s\n", doc.Filename, titlePart)
		}
	}

	// Document relations summary.
	if m, mErr := sync.LoadManifest(baseDir); mErr == nil && len(m.Relations) > 0 {
		fmt.Fprintf(&b, "\nDOCUMENT RELATIONS: %d relation(s) stored.\n", len(m.Relations))
		b.WriteString("  Use list_relations, add_relation, remove_relation MCP tools to manage.\n")
	}

	b.WriteString("\nRefer to MCP server instructions for document types, workflow rules, and usage guidance.\n")

	return b.String(), len(docs)
}
