package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"archcore-cli/internal/config"
	"archcore-cli/internal/display"
	"archcore-cli/internal/sync"
	"archcore-cli/templates"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newValidateCmd() *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate .archcore/ structure and documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			issues := runValidate(cwd, fix)
			if issues > 0 {
				return fmt.Errorf("%d issue(s) found", issues)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Automatically fix issues (e.g., remove orphaned relations)")
	return cmd
}

// runValidate checks .archcore/ structure and documents. Returns issue count.
func runValidate(baseDir string, fix bool) int {
	issues := runValidateChecks(baseDir, fix)
	if issues > 0 {
		fmt.Println(display.Warn.Render(fmt.Sprintf("  %d issue(s) found", issues)))
	}
	return issues
}

// runValidateChecks performs all validation checks without printing a summary.
// It is used by both the validate and doctor commands.
func runValidateChecks(baseDir string, fix bool) int {
	if !config.DirExists(baseDir) {
		fmt.Println(display.FailLine(".archcore/ directory not found"))
		fmt.Println(display.HintLine("Run 'archcore init' to set up"))
		return 1
	}

	issues := 0
	issues += checkStructure(baseDir)
	issues += checkFiles(baseDir)
	issues += checkManifest(baseDir, fix)
	return issues
}

func checkStructure(baseDir string) int {
	fmt.Println(display.CheckLine(".archcore/ exists"))
	return 0
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)


func checkFiles(baseDir string) int {
	issues := 0
	archcoreDir := filepath.Join(baseDir, ".archcore")

	walkErr := templates.WalkArchcoreFiles(archcoreDir, func(path string, d fs.DirEntry) error {
		name := d.Name()

		relPath, _ := filepath.Rel(baseDir, path)
		relPath = filepath.ToSlash(relPath)

		issues += checkNaming(relPath, name)

		// Check type validity (but no category placement check — directory is free-form).
		docType := templates.ExtractDocType(name)
		if docType != "" && !templates.IsValidType(docType) {
			issues++
			fmt.Println(display.FailLine(fmt.Sprintf("%s: unknown document type %q", relPath, docType)))
			fmt.Println(display.HintLine(fmt.Sprintf("valid types: %s", strings.Join(templates.ValidTypes(), ", "))))
		}

		issues += checkFrontmatter(path, relPath)
		return nil
	})
	if walkErr != nil {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("error scanning .archcore/: %v", walkErr)))
	}

	return issues
}

func checkNaming(relPath, filename string) int {
	issues := 0
	name := strings.TrimSuffix(filename, ".md")
	parts := strings.Split(name, ".")

	if len(parts) < 2 {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("%s: filename must match <slug>.<type>.md", relPath)))
		fmt.Println(display.HintLine("example: oauth-user.adr.md"))
		return issues
	}

	slug := strings.Join(parts[:len(parts)-1], ".")
	if !slugRe.MatchString(slug) {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("%s: slug must be lowercase alphanumeric with hyphens", relPath)))
		fmt.Println(display.HintLine("example: my-feature"))
	}

	return issues
}

func checkFrontmatter(absPath, relPath string) int {
	issues := 0

	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0
	}
	content := string(data)

	// Check for frontmatter delimiters.
	if !strings.HasPrefix(content, "---\n") {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("%s: missing YAML frontmatter", relPath)))
		fmt.Println(display.HintLine("file must start with --- delimiters"))
		return issues
	}

	endIdx := strings.Index(content[4:], "\n---")
	if endIdx < 0 {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("%s: missing closing --- delimiter", relPath)))
		return issues
	}

	fmContent := content[4 : 4+endIdx]

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmContent), &fm); err != nil {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("%s: invalid YAML in frontmatter", relPath)))
		fmt.Println(display.HintLine(err.Error()))
		return issues
	}

	// Check required fields.
	for _, field := range []string{"title", "status"} {
		val, ok := fm[field]
		if !ok {
			issues++
			fmt.Println(display.FailLine(fmt.Sprintf("%s: missing required field %q", relPath, field)))
		} else if str, isStr := val.(string); isStr && str == "" {
			issues++
			fmt.Println(display.FailLine(fmt.Sprintf("%s: missing required field %q", relPath, field)))
		}
	}

	// Check meta is a mapping if present.
	if meta, ok := fm["meta"]; ok {
		if _, isMap := meta.(map[string]any); !isMap {
			issues++
			fmt.Println(display.FailLine(fmt.Sprintf("%s: \"meta\" must be an object (YAML mapping)", relPath)))
		}
	}

	return issues
}

func checkManifest(baseDir string, fix bool) int {
	issues := 0
	manifestPath := filepath.Join(baseDir, ".archcore", sync.ManifestFile)

	data, err := os.ReadFile(manifestPath)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(display.CheckLine("No sync manifest (first sync pending)"))
		return 0
	}
	if err != nil {
		fmt.Println(display.FailLine(fmt.Sprintf("Cannot read sync manifest: %v", err)))
		return 1
	}

	jsonIssues := sync.ValidateManifestJSON(data)
	for _, issue := range jsonIssues {
		issues++
		fmt.Println(display.FailLine(fmt.Sprintf("Sync manifest: %s", issue)))
	}

	if len(jsonIssues) == 0 {
		var m sync.Manifest
		if err := json.Unmarshal(data, &m); err == nil {
			if m.Files == nil {
				m.Files = make(map[string]string)
			}
			semIssues := sync.ValidateManifest(&m)
			for _, issue := range semIssues {
				issues++
				fmt.Println(display.FailLine(fmt.Sprintf("Sync manifest: %s", issue)))
			}

			danglingIssues := checkDanglingRelations(baseDir, m.Relations)
			if fix && len(danglingIssues) > 0 {
				archcoreDir := filepath.Join(baseDir, ".archcore")
				removed := m.CleanupRelations(archcoreDir)
				if err := sync.SaveManifest(baseDir, &m); err != nil {
					issues++
					fmt.Println(display.FailLine(fmt.Sprintf("Failed to save manifest after cleanup: %v", err)))
				} else {
					fmt.Println(display.CheckLine(fmt.Sprintf("Removed %d orphaned relation(s)", removed)))
				}
			} else {
				for _, issue := range danglingIssues {
					issues++
					fmt.Println(display.FailLine(fmt.Sprintf("Sync manifest: %s", issue)))
				}
			}

			if len(semIssues) == 0 && (len(danglingIssues) == 0 || fix) {
				fmt.Println(display.CheckLine(fmt.Sprintf("Sync manifest valid (%d file(s) tracked, %d relation(s))", len(m.Files), len(m.Relations))))
			}
		} else {
			issues++
			fmt.Println(display.FailLine(fmt.Sprintf("Sync manifest: invalid JSON: %v", err)))
		}
	}

	if issues > 0 {
		fmt.Println(display.HintLine("Delete .archcore/.sync-state.json and re-sync"))
	}

	return issues
}

func checkDanglingRelations(baseDir string, relations []sync.Relation) []string {
	var issues []string
	for _, rel := range relations {
		srcPath := filepath.Join(baseDir, ".archcore", rel.Source)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("relation source %q does not exist on disk", rel.Source))
		}
		tgtPath := filepath.Join(baseDir, ".archcore", rel.Target)
		if _, err := os.Stat(tgtPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("relation target %q does not exist on disk", rel.Target))
		}
	}
	return issues
}
