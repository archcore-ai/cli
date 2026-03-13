package mcp

import (
	"archcore-cli/internal/config"
	"archcore-cli/internal/mcp/tools"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

var mcpServerInstructions = `You are working with a project that uses Archcore — System Context Platform. Keeps humans and AI in sync with your system.

The .archcore/ directory contains Markdown files with YAML frontmatter (title, status). The directory structure inside .archcore/ is free-form — you can organize documents by domain, feature, team, or any custom structure. Categories (vision, knowledge, experience) are virtual — derived automatically from the document type in the filename (slug.type.md), not from the physical directory.

Example structures:
  .archcore/auth/jwt-strategy.adr.md         → virtual category: knowledge
  .archcore/auth/auth-redesign.prd.md        → virtual category: vision
  .archcore/payments/stripe.adr.md           → virtual category: knowledge
  .archcore/infrastructure/k8s/migration.adr.md → virtual category: knowledge
  .archcore/my-doc.rule.md                   → virtual category: knowledge (root level)

Document types and their virtual categories:
  knowledge: adr (decisions), rfc (proposals), rule (standards), guide (how-tos), doc (reference), project (project overview)
  vision:    prd (requirements), idea (concepts), plan (action plans)
  experience: task-type (typical task patterns), cpat (code pattern changes)

DOCUMENT RELATIONS:
Documents can be linked with directed relations stored in the sync manifest.
  Relation types:
    related     — general association (e.g., two ADRs on the same topic)
    implements  — source implements what target specifies (e.g., plan implements prd)
    extends     — source builds upon target (e.g., rfc extends an existing adr)
    depends_on  — source requires target to proceed (e.g., plan depends_on adr)

  After creating a document, check the nearby_documents hint in the response.
  Use add_relation to link related documents. Use list_relations to see existing links.

PATH FORMAT: All tool paths use ".archcore/<path>/<slug>.<type>.md" as returned by list_documents. The add_relation and remove_relation tools also accept paths without the ".archcore/" prefix.

WORKFLOW RULES:
1. Before creating any document, call list_documents first to check whether a relevant document already exists. Do not create duplicates.
2. To read a document, call list_documents to get its path, then pass that path to get_document.
3. Only call create_document after confirming no equivalent document exists.
4. Before updating a document, confirm the intended changes with the user when possible.
5. After creating a document, review nearby_documents and consider adding relations with add_relation.
6. When reading a document, check outgoing_relations and incoming_relations for context.
7. Before deleting a document, confirm explicitly with the user. Prefer setting status to "rejected" when historical context is worth keeping.

WHEN TO CREATE:
- A technical decision is made or finalized → adr
- A significant change is being proposed for team review → rfc
- A team standard or required behavior is established → rule
- Step-by-step instructions for completing a task → guide
- Reference information, registries, lookup tables, or general documentation → doc
- A project overview with architecture, components, and getting-started info → project
- A proven workflow for a recurring implementation task is documented → task-type
- A coding pattern, convention, or approach has deliberately changed → cpat
- A product concept or technical idea needs capturing → idea
- An implementation plan with tasks is formed → plan
- Product requirements with goals, scope, and acceptance criteria → prd

WHEN TO UPDATE (use update_document):
- A decision is finalized → change status from "draft" to "accepted"
- A proposal is rejected → change status to "rejected"
- A plan's scope or tasks change → update content
- Do not create a new document when the existing one should be updated.

WHEN TO DELETE (use remove_document):
- A document was created by mistake or is a duplicate → delete it
- A document is entirely irrelevant and has no historical value → delete it
- Prefer update_document with status "rejected" to preserve history when the content was once valid.
- Always confirm with the user before deleting — this is permanent and removes all relations.

TYPE SELECTION RULES (use these to disambiguate):
- rule vs doc: A rule contains imperative statements ("Always do X", "Never do Y") with good/bad code examples and enforcement info. A doc is descriptive reference material (tables, registries, explanations). If the content describes what exists rather than prescribing behavior, use doc.
- adr vs rfc: An adr records a decision already made. An rfc proposes a change open for review. If the decision is final, use adr; if still open for feedback, use rfc.
- guide vs doc: A guide has numbered steps the reader follows sequentially. A doc is non-sequential reference material. If the reader is meant to do something step-by-step, use guide; if they look things up, use doc.

VALID STATUS VALUES:
  draft     — default for new documents; work in progress
  accepted  — finalized or approved; set only when the human confirms
  rejected  — superseded, abandoned, or declined; preserves history

CODE REFERENCES (optional):
Documents may reference source code paths using @-notation (e.g., @cmd/sync.go, @internal/config/).
This is optional but encouraged — it helps agents navigate between documentation and code, and enables future staleness detection.
When writing or updating documents, include relevant code paths where they naturally fit (e.g., in "Implementation Notes", "Key files", "Related" sections).

NEVER create documents for: temporary notes, questions, chat summaries, or speculative content without clear value.
ALWAYS use a descriptive slug (lowercase, hyphens only) and a clear human-readable title.`

// buildInstructions returns MCP server instructions with an optional language directive appended.
func buildInstructions(language string) string {
	if language == "" || language == "en" {
		return mcpServerInstructions
	}
	return mcpServerInstructions + fmt.Sprintf(`

LANGUAGE REQUIREMENT:
All document content (title, body text) MUST be written in %q. YAML frontmatter keys and status values remain in English. Slug must still be lowercase ASCII with hyphens.`, language)
}

// NewServer creates a new MCP server with archcore tools.
func NewServer(baseDir string) *server.MCPServer {
	language := ""
	if settings, err := config.Load(baseDir); err == nil {
		language = settings.Language
	}

	s := server.NewMCPServer(
		"archcore",
		"1.0.0",
		server.WithInstructions(buildInstructions(language)),
	)

	s.AddTool(tools.NewListDocumentsTool(), tools.HandleListDocuments(baseDir))
	s.AddTool(tools.NewGetDocumentTool(), tools.HandleGetDocument(baseDir))
	s.AddTool(tools.NewCreateDocumentTool(), tools.HandleCreateDocument(baseDir))
	s.AddTool(tools.NewUpdateDocumentTool(), tools.HandleUpdateDocument(baseDir))
	s.AddTool(tools.NewRemoveDocumentTool(), tools.HandleRemoveDocument(baseDir))
	s.AddTool(tools.NewAddRelationTool(), tools.HandleAddRelation(baseDir))
	s.AddTool(tools.NewRemoveRelationTool(), tools.HandleRemoveRelation(baseDir))
	s.AddTool(tools.NewListRelationsTool(), tools.HandleListRelations(baseDir))

	return s
}

// RunStdio starts the MCP server on stdin/stdout.
func RunStdio(baseDir string) error {
	s := NewServer(baseDir)
	return server.ServeStdio(s)
}
