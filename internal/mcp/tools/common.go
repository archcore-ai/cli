package tools

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"archcore-cli/templates"
)

// DocumentRelation represents one side of a relation for enriched output.
type DocumentRelation struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// EnrichedDocument extends LocalDocument with relation information.
type EnrichedDocument struct {
	LocalDocument
	OutgoingRelations []DocumentRelation `json:"outgoing_relations,omitempty"`
	IncomingRelations []DocumentRelation `json:"incoming_relations,omitempty"`
}

// LocalDocument represents a document discovered in .archcore/.
type LocalDocument struct {
	Path     string `json:"path"`               // relative: ".archcore/auth/jwt-strategy.adr.md"
	Category string `json:"category"`            // virtual: vision, knowledge, experience (derived from type)
	Type     string `json:"type"`                // adr, rfc, rule...
	Filename string `json:"filename"`            // "jwt-strategy.adr.md"
	Slug     string `json:"slug"`                // "jwt-strategy"
	Title    string `json:"title,omitempty"`      // from frontmatter
	Status   string `json:"status,omitempty"`     // from frontmatter
	Content  string `json:"content,omitempty"`    // full markdown (optional)
}

// ScanDocuments discovers all .md files recursively inside .archcore/.
func ScanDocuments(baseDir string) ([]LocalDocument, error) {
	archcoreDir := filepath.Join(baseDir, ".archcore")
	var docs []LocalDocument

	err := templates.WalkArchcoreFiles(archcoreDir, func(path string, d fs.DirEntry) error {
		name := d.Name()

		docType := templates.ExtractDocType(name)
		category := templates.CategoryForType(templates.DocumentType(docType))
		title, status := extractFrontmatter(path)
		slug := templates.ExtractSlug(name)

		relPath, _ := filepath.Rel(baseDir, path)
		relPath = filepath.ToSlash(relPath)

		docs = append(docs, LocalDocument{
			Path:     relPath,
			Category: category,
			Type:     docType,
			Filename: name,
			Slug:     slug,
			Title:    title,
			Status:   status,
		})
		return nil
	})

	if os.IsNotExist(err) {
		return nil, nil
	}
	return docs, err
}

// ReadDocumentContent reads a single document fully from a relative path.
func ReadDocumentContent(baseDir, relPath string) (LocalDocument, error) {
	absPath := filepath.Join(baseDir, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return LocalDocument{}, err
	}

	filename := filepath.Base(relPath)
	docType := templates.ExtractDocType(filename)
	category := templates.CategoryForType(templates.DocumentType(docType))
	title, status, _ := templates.SplitDocument(data)
	slug := templates.ExtractSlug(filename)

	return LocalDocument{
		Path:     relPath,
		Category: category,
		Type:     docType,
		Filename: filename,
		Slug:     slug,
		Title:    title,
		Status:   status,
		Content:  string(data),
	}, nil
}

// buildDocumentFile reconstructs a full document file from frontmatter fields and body.
func buildDocumentFile(title, status, body string) string {
	var buf strings.Builder
	buf.WriteString("---\n")
	fmt.Fprintf(&buf, "title: %q\n", title)
	buf.WriteString("status: " + status + "\n")
	buf.WriteString("---\n\n")
	buf.WriteString(body)
	return buf.String()
}

// stripFrontmatter removes YAML frontmatter from content if present.
// This prevents duplicate frontmatter when callers (e.g. AI agents)
// include frontmatter in the content parameter despite the tool
// description specifying body-only content.
func stripFrontmatter(content string) string {
	s := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return content
	}
	end := strings.Index(s[4:], "\n---\n")
	if end == -1 {
		// Check for frontmatter at the very end (no trailing newline after closing ---).
		end = strings.Index(s[4:], "\n---")
		if end == -1 || end+4+len("\n---") != len(s) {
			return content
		}
		// Frontmatter block ends at EOF with no body.
		return ""
	}
	end += 4 // adjust for the offset from s[4:]
	body := s[end+5:] // skip past "\n---\n"
	body = strings.TrimPrefix(body, "\n")
	return body
}

// validateArchcorePath normalises and validates a document path.
// It returns the cleaned path or an error message if the path is invalid.
func validateArchcorePath(relPath string) (string, string) {
	relPath = filepath.ToSlash(relPath)
	if !strings.HasPrefix(relPath, ".archcore/") {
		return "", "invalid path: must start with \".archcore/\""
	}
	cleaned := filepath.Clean(relPath)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) || !strings.HasPrefix(cleaned, ".archcore/") {
		return "", "invalid path: must be relative and within .archcore/"
	}
	return cleaned, ""
}

// extractFrontmatter reads the YAML frontmatter to extract title and status.
func extractFrontmatter(path string) (title, status string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	// Read up to 1024 bytes for frontmatter extraction.
	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		return "", ""
	}
	t, s, _ := templates.SplitDocument(buf[:n])
	return t, s
}
