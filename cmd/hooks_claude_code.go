package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"archcore-cli/internal/display"

	"github.com/spf13/cobra"
)

// hookInput represents the JSON payload Claude Code sends via stdin.
type hookInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	HookEventName string `json:"hook_event_name"`
	Source        string `json:"source"`
}

// hookOutput is the JSON response written to stdout.
type hookOutput struct {
	HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"`
	SystemMessage      string         `json:"systemMessage,omitempty"`
}

// resolveBaseDir returns the base directory from hook input, falling back to cwd.
func resolveBaseDir(input *hookInput) (string, error) {
	if input.CWD != "" {
		return input.CWD, nil
	}
	return os.Getwd()
}

// readHookInput parses the hook input JSON from a reader.
func readHookInput(r io.Reader) (*hookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing hook input: %w", err)
	}
	return &input, nil
}

func newHooksClaudeCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude-code",
		Short: "Handle Claude Code hook events",
	}
	cmd.AddCommand(
		newSessionStartCmd(),
	)
	return cmd
}

func newSessionStartCmd() *cobra.Command {
	return newSessionStartHookCmd("session-start", "Handle SessionStart hook event")
}

// --- Hook command factories ---

// newSessionStartHookCmd creates a session-start hook command (shared across agents).
func newSessionStartHookCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:    use,
		Short:  short,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := readHookInput(os.Stdin)
			if err != nil {
				return err
			}
			baseDir, err := resolveBaseDir(input)
			if err != nil {
				return err
			}
			out, err := handleSessionStart(baseDir)
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(out)
			return err
		},
	}
}

// --- Session Start Handler (Claude Code adapter) ---

func handleSessionStart(baseDir string) ([]byte, error) {
	ctx, docCount := buildSessionContext(baseDir)
	output := hookOutput{
		HookSpecificOutput: map[string]any{
			"hookEventName":     "SessionStart",
			"additionalContext": ctx,
		},
		SystemMessage: display.HookConnectedLine(docCount),
	}
	return json.Marshal(output)
}
