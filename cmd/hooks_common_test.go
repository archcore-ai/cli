package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"archcore-cli/internal/sync"
)

func TestBuildSessionContext_Empty(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	ctx, _ := buildSessionContext(base)

	if !strings.Contains(ctx, "Archcore") {
		t.Error("missing Archcore header")
	}
	for _, cat := range []string{"knowledge", "vision", "experience"} {
		if !strings.Contains(ctx, "  ["+cat+"]\n    (none)") {
			t.Errorf("category %s should show (none)", cat)
		}
	}
}

func TestBuildSessionContext_WithDocs(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	knowledgeDoc := filepath.Join(base, ".archcore", "knowledge", "use-postgres.adr.md")
	if err := os.WriteFile(knowledgeDoc, []byte("---\ntitle: Use PostgreSQL\nstatus: accepted\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	visionDoc := filepath.Join(base, ".archcore", "vision", "mvp.plan.md")
	if err := os.WriteFile(visionDoc, []byte("---\ntitle: MVP Plan\nstatus: draft\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ctx, _ := buildSessionContext(base)

	if !strings.Contains(ctx, "use-postgres.adr.md") {
		t.Error("context missing knowledge doc")
	}
	if !strings.Contains(ctx, "mvp.plan.md") {
		t.Error("context missing vision doc")
	}
	if !strings.Contains(ctx, "Refer to MCP server instructions") {
		t.Error("context missing MCP referral line")
	}
}

func TestBuildSessionContext_WithRelations(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	m := sync.NewManifest()
	m.AddRelation("knowledge/a.adr.md", "vision/b.prd.md", sync.RelImplements)
	m.AddRelation("knowledge/a.adr.md", "knowledge/c.rfc.md", sync.RelRelated)
	if err := sync.SaveManifest(base, m); err != nil {
		t.Fatal(err)
	}

	ctx, _ := buildSessionContext(base)
	if !strings.Contains(ctx, "DOCUMENT RELATIONS") {
		t.Error("expected DOCUMENT RELATIONS section")
	}
	if !strings.Contains(ctx, "2 relation(s)") {
		t.Error("expected '2 relation(s)' count")
	}
}

func TestBuildSessionContext_MentionsRelationTools(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	ctx, _ := buildSessionContext(base)
	if !strings.Contains(ctx, "add_relation") {
		t.Error("expected 'add_relation' in context")
	}
	if !strings.Contains(ctx, "list_relations") {
		t.Error("expected 'list_relations' in context")
	}
	if !strings.Contains(ctx, "remove_relation") {
		t.Error("expected 'remove_relation' in context")
	}
}

func TestBuildSessionContext_WithTags(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	doc := filepath.Join(base, ".archcore", "knowledge", "tagged.adr.md")
	if err := os.WriteFile(doc, []byte("---\ntitle: Tagged Doc\nstatus: draft\ntags:\n  - frontend\n  - auth\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc2 := filepath.Join(base, ".archcore", "vision", "tagged2.prd.md")
	if err := os.WriteFile(doc2, []byte("---\ntitle: Another\nstatus: draft\ntags:\n  - frontend\n  - backend\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, _ := buildSessionContext(base)
	if !strings.Contains(ctx, "EXISTING TAGS:") {
		t.Error("expected EXISTING TAGS section")
	}
	if !strings.Contains(ctx, "frontend") {
		t.Error("expected 'frontend' tag in context")
	}
	if !strings.Contains(ctx, "auth") {
		t.Error("expected 'auth' tag in context")
	}
	if !strings.Contains(ctx, "backend") {
		t.Error("expected 'backend' tag in context")
	}
}

func TestBuildSessionContext_NoTags(t *testing.T) {
	t.Parallel()
	base := setupArchcoreDir(t)

	doc := filepath.Join(base, ".archcore", "knowledge", "untagged.adr.md")
	if err := os.WriteFile(doc, []byte("---\ntitle: Untagged\nstatus: draft\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, _ := buildSessionContext(base)
	if strings.Contains(ctx, "EXISTING TAGS:") {
		t.Error("should not show EXISTING TAGS when no tags exist")
	}
}
