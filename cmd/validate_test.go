package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"archcore-cli/internal/config"
	"archcore-cli/internal/sync"
)

// runValidateInDir executes "archcore validate" in dir and returns stdout + error.
func runValidateInDir(t *testing.T, dir string) (string, error) {
	t.Helper()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root := NewRootCmd("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"validate"})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	execErr := root.Execute()
	w.Close()
	os.Stdout = oldStdout

	var out bytes.Buffer
	out.ReadFrom(r)
	return out.String(), execErr
}

// writeDoc creates a .md file with given content inside .archcore/<subdir>/.
// Creates the subdirectory if it doesn't exist.
func writeDoc(t *testing.T, dir, subdir, filename, content string) {
	t.Helper()
	d := filepath.Join(dir, ".archcore", subdir)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(d, filename)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const validFrontmatter = "---\ntitle: Test Doc\nstatus: draft\n---\n\nBody.\n"

func initValidDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := config.InitDir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestValidate_NoArchcoreDir(t *testing.T) {
	dir := t.TempDir()
	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in output, got: %s", out)
	}
}

func TestValidate_CustomDirectoryAllowed(t *testing.T) {
	dir := initValidDir(t)
	os.MkdirAll(filepath.Join(dir, ".archcore", "auth"), 0o755)
	writeDoc(t, dir, "auth", "jwt.adr.md", validFrontmatter)

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error for custom directory, got: %v", err)
	}
}

func TestValidate_ValidStructureAndFiles(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "use-postgres.adr.md", validFrontmatter)

	out, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "exists") {
		t.Errorf("expected 'exists' checks in output, got: %s", out)
	}
}

func TestValidate_BadFilename_NoTypeSegment(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "readme.md", validFrontmatter)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "<slug>.<type>.md") {
		t.Errorf("expected '<slug>.<type>.md' hint, got: %s", out)
	}
}

func TestValidate_BadSlug_Uppercase(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "MyFeature.adr.md", validFrontmatter)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "lowercase") {
		t.Errorf("expected 'lowercase' in output, got: %s", out)
	}
}

func TestValidate_UnknownDocumentType(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-feature.banana.md", validFrontmatter)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "unknown") {
		t.Errorf("expected 'unknown' in output, got: %s", out)
	}
}

func TestValidate_AnyDirectoryForAnyType(t *testing.T) {
	dir := initValidDir(t)
	// task-type in any directory is fine — categories are virtual.
	writeDoc(t, dir, "vision", "my-task.task-type.md", validFrontmatter)

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_MissingFrontmatter(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-doc.adr.md", "# No frontmatter\n")

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "missing YAML frontmatter") {
		t.Errorf("expected 'missing YAML frontmatter', got: %s", out)
	}
}

func TestValidate_MissingTitle(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-doc.adr.md", "---\nstatus: draft\n---\n")

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "missing required field") {
		t.Errorf("expected 'missing required field', got: %s", out)
	}
}

func TestValidate_MissingStatus(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-doc.adr.md", "---\ntitle: Hello\n---\n")

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "missing required field") {
		t.Errorf("expected 'missing required field', got: %s", out)
	}
}

func TestValidate_MetaAsScalar(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-doc.adr.md", "---\ntitle: Hello\nstatus: draft\nmeta: not-a-map\n---\n")

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "must be an object") {
		t.Errorf("expected 'must be an object', got: %s", out)
	}
}

func TestValidate_ValidMeta(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "knowledge", "my-doc.adr.md", "---\ntitle: Hello\nstatus: draft\nmeta:\n  key: value\n---\n")

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_NoManifestFile(t *testing.T) {
	dir := initValidDir(t)
	out, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "No sync manifest") {
		t.Errorf("expected 'No sync manifest' in output, got: %s", out)
	}
}

func TestValidate_ValidManifest(t *testing.T) {
	dir := initValidDir(t)
	data := `{"version":1,"files":{"vision/test.md":"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"}}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Sync manifest valid") {
		t.Errorf("expected 'Sync manifest valid' in output, got: %s", out)
	}
	if !strings.Contains(out, "1 file(s) tracked, 0 relation(s)") {
		t.Errorf("expected '1 file(s) tracked, 0 relation(s)' in output, got: %s", out)
	}
}

func TestValidate_CorruptManifest(t *testing.T) {
	dir := initValidDir(t)
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte("{truncated"), 0o644)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in output, got: %s", out)
	}
	if !strings.Contains(out, "Delete .archcore/.sync-state.json") {
		t.Errorf("expected delete hint in output, got: %s", out)
	}
}

func TestValidate_InvalidHashInManifest(t *testing.T) {
	dir := initValidDir(t)
	data := `{"version":1,"files":{"vision/test.md":"short"}}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "not valid SHA-256") {
		t.Errorf("expected 'not valid SHA-256' in output, got: %s", out)
	}
}

func TestValidate_DeeplyNestedFiles(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "infrastructure/k8s/prod", "migration.adr.md", validFrontmatter)

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error for deeply nested file, got: %v", err)
	}
}

func TestValidate_FileInRoot(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "", "root-doc.adr.md", validFrontmatter)

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error for file in .archcore/ root, got: %v", err)
	}
}

func TestValidate_NonMdFilesIgnored(t *testing.T) {
	dir := initValidDir(t)
	// Write a non-.md file that should be silently skipped.
	d := filepath.Join(dir, ".archcore", "knowledge")
	os.MkdirAll(d, 0o755)
	os.WriteFile(filepath.Join(d, "notes.txt"), []byte("not a doc"), 0o644)
	// Also write a valid doc so we get meaningful output.
	writeDoc(t, dir, "knowledge", "real.adr.md", validFrontmatter)

	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error (non-.md files ignored), got: %v", err)
	}
}

func TestValidate_EmptyArchcoreDir(t *testing.T) {
	dir := initValidDir(t)
	// .archcore/ exists but has no documents.
	_, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error for empty .archcore/ dir, got: %v", err)
	}
}

func TestValidate_ManifestWithCustomDirPath(t *testing.T) {
	dir := initValidDir(t)
	// Manifest with a custom directory path (not vision/knowledge/experience).
	data := `{"version":1,"files":{"auth/jwt-strategy.adr.md":"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"}}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error for custom-dir manifest path, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Sync manifest valid") {
		t.Errorf("expected 'Sync manifest valid', got: %s", out)
	}
}

func TestValidate_ManifestWithValidRelations(t *testing.T) {
	dir := initValidDir(t)
	// Create actual documents referenced by the relation.
	writeDoc(t, dir, "", "a.adr.md", validFrontmatter)
	writeDoc(t, dir, "", "b.prd.md", validFrontmatter)
	data := `{"version":1,"files":{},"relations":[{"source":"a.adr.md","target":"b.prd.md","type":"implements"}]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "1 relation(s)") {
		t.Errorf("expected '1 relation(s)' in output, got: %s", out)
	}
}

func TestValidate_ManifestWithDanglingRelation(t *testing.T) {
	dir := initValidDir(t)
	// Only create source, not target.
	writeDoc(t, dir, "", "a.adr.md", validFrontmatter)
	data := `{"version":1,"files":{},"relations":[{"source":"a.adr.md","target":"nonexistent.prd.md","type":"related"}]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error for dangling relation, got nil")
	}
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected 'does not exist' in output, got: %s", out)
	}
}

func TestValidate_ManifestWithInvalidRelationType(t *testing.T) {
	dir := initValidDir(t)
	data := `{"version":1,"files":{},"relations":[{"source":"a.adr.md","target":"b.prd.md","type":"blocks"}]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error for invalid relation type, got nil")
	}
	if !strings.Contains(out, "invalid type") {
		t.Errorf("expected 'invalid type' in output, got: %s", out)
	}
}

func runValidateFixInDir(t *testing.T, dir string) (string, error) {
	t.Helper()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root := NewRootCmd("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"validate", "--fix"})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	execErr := root.Execute()
	w.Close()
	os.Stdout = oldStdout

	var out bytes.Buffer
	out.ReadFrom(r)
	return out.String(), execErr
}

func TestValidate_FixRemovesDanglingRelations(t *testing.T) {
	dir := initValidDir(t)
	// Create only source file, not target — so the relation is dangling.
	writeDoc(t, dir, "", "a.adr.md", validFrontmatter)
	data := `{"version":1,"files":{},"relations":[
		{"source":"a.adr.md","target":"nonexistent.prd.md","type":"related"},
		{"source":"a.adr.md","target":"also-gone.rfc.md","type":"implements"}
	]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateFixInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error with --fix, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Removed 2 orphaned relation(s)") {
		t.Errorf("expected 'Removed 2 orphaned relation(s)' in output, got: %s", out)
	}
	if !strings.Contains(out, "Sync manifest valid") {
		t.Errorf("expected 'Sync manifest valid' in output, got: %s", out)
	}

	// Verify the manifest on disk has no relations left.
	m, loadErr := sync.LoadManifest(dir)
	if loadErr != nil {
		t.Fatalf("LoadManifest after fix: %v", loadErr)
	}
	if len(m.Relations) != 0 {
		t.Errorf("expected 0 relations after fix, got %d", len(m.Relations))
	}
}

func TestValidate_FixKeepsValidRelations(t *testing.T) {
	dir := initValidDir(t)
	writeDoc(t, dir, "", "a.adr.md", validFrontmatter)
	writeDoc(t, dir, "", "b.prd.md", validFrontmatter)
	data := `{"version":1,"files":{},"relations":[
		{"source":"a.adr.md","target":"b.prd.md","type":"implements"},
		{"source":"a.adr.md","target":"gone.rfc.md","type":"related"}
	]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateFixInDir(t, dir)
	if err != nil {
		t.Fatalf("expected no error with --fix, got: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Removed 1 orphaned relation(s)") {
		t.Errorf("expected 'Removed 1 orphaned relation(s)' in output, got: %s", out)
	}

	// Verify the valid relation is still there.
	m, loadErr := sync.LoadManifest(dir)
	if loadErr != nil {
		t.Fatalf("LoadManifest after fix: %v", loadErr)
	}
	if len(m.Relations) != 1 {
		t.Fatalf("expected 1 relation after fix, got %d", len(m.Relations))
	}
	if m.Relations[0].Target != "b.prd.md" {
		t.Errorf("expected kept relation target b.prd.md, got %s", m.Relations[0].Target)
	}
}

func TestValidate_ManifestWithDuplicateRelation(t *testing.T) {
	dir := initValidDir(t)
	data := `{"version":1,"files":{},"relations":[
		{"source":"a.adr.md","target":"b.prd.md","type":"related"},
		{"source":"a.adr.md","target":"b.prd.md","type":"related"}
	]}`
	os.WriteFile(filepath.Join(dir, ".archcore", sync.ManifestFile), []byte(data), 0o644)

	out, err := runValidateInDir(t, dir)
	if err == nil {
		t.Fatal("expected error for duplicate relation, got nil")
	}
	if !strings.Contains(out, "duplicate") {
		t.Errorf("expected 'duplicate' in output, got: %s", out)
	}
}
