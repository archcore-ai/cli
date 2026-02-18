package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"archcore-cli/internal/display"
	"archcore-cli/templates"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

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

// runCreate performs the create logic without any cobra dependency.
func runCreate(baseDir, docType, filename, title, status string) error {
	if !templates.IsValidType(docType) {
		return fmt.Errorf("invalid document type %q (valid: %s)", docType, strings.Join(templates.ValidTypes(), ", "))
	}

	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("filename is required")
	}
	if strings.ContainsAny(filename, "/\\") {
		return fmt.Errorf("invalid filename %q: must not contain path separators", filename)
	}

	if title == "" {
		title = filenameToTitle(filename)
	}
	if status == "" {
		status = "draft"
	}

	category := templates.CategoryForType(templates.DocumentType(docType))
	dir := filepath.Join(baseDir, ".archcore", category)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	outputFile := filepath.Join(dir, filename+"."+docType+".md")
	if _, err := os.Stat(outputFile); err == nil {
		return fmt.Errorf("file already exists: %s", outputFile)
	}

	body := templates.GenerateTemplate(templates.DocumentType(docType))

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString("title: " + title + "\n")
	buf.WriteString("status: " + status + "\n")
	buf.WriteString("---\n\n")
	buf.WriteString(body)

	if err := os.WriteFile(outputFile, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	rel, _ := filepath.Rel(baseDir, outputFile)
	fmt.Println(display.CheckLine("Created " + rel))
	return nil
}

func newCreateCmd() *cobra.Command {
	var flagTitle string
	var flagStatus string

	cmd := &cobra.Command{
		Use:   "create [type] [filename]",
		Short: "Create a new document from a template",
		Long: `Create a new archcore document with YAML frontmatter and a type-specific template.

Examples:
  archcore create rfc oauth-tokens
  archcore create adr use-postgres --title "Use PostgreSQL" --status accepted
  archcore create                    # interactive mode`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			var docType, filename string

			switch len(args) {
			case 2:
				docType = args[0]
				filename = args[1]
			case 1:
				docType = args[0]
				if !templates.IsValidType(docType) {
					return fmt.Errorf("invalid document type %q (valid: %s)", docType, strings.Join(templates.ValidTypes(), ", "))
				}
				err := huh.NewInput().
					Title("Filename (slug)").
					Placeholder("e.g. oauth-tokens").
					Value(&filename).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("filename is required")
						}
						return nil
					}).
					Run()
				if err != nil {
					return err
				}
			case 0:
				typeOptions := make([]huh.Option[string], 0, len(templates.ValidTypes()))
				for _, t := range templates.ValidTypes() {
					cat := templates.CategoryForType(templates.DocumentType(t))
					typeOptions = append(typeOptions, huh.NewOption(t+" ("+cat+")", t))
				}
				err := huh.NewSelect[string]().
					Title("Document type").
					Options(typeOptions...).
					Value(&docType).
					Run()
				if err != nil {
					return err
				}
				err = huh.NewInput().
					Title("Filename (slug)").
					Placeholder("e.g. oauth-tokens").
					Value(&filename).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("filename is required")
						}
						return nil
					}).
					Run()
				if err != nil {
					return err
				}
			}

			return runCreate(cwd, docType, filename, flagTitle, flagStatus)
		},
	}

	cmd.Flags().StringVar(&flagTitle, "title", "", "document title (default: derived from filename)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "document status (default: draft)")

	return cmd
}
