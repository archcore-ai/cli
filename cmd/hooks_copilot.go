package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"archcore-cli/internal/agents"
	"archcore-cli/internal/config"
	"archcore-cli/internal/display"

	"github.com/spf13/cobra"
)

// copilotHooksConfig represents the .github/hooks/<name>.json structure.
type copilotHooksConfig struct {
	Version int                            `json:"version"`
	Hooks   map[string][]copilotHookEntry `json:"hooks"`
}

// copilotHookEntry represents a single hook in Copilot config.
type copilotHookEntry struct {
	Type string `json:"type"`
	Bash string `json:"bash"`
}

// copilotHookEvents maps event names to archcore commands.
var copilotHookEvents = []struct {
	Event   string
	Command string
}{
	{"sessionStart", "archcore hooks copilot session-start"},
}

func newHooksCopilotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copilot",
		Short: "Handle GitHub Copilot hook events",
	}
	cmd.AddCommand(
		newSessionStartHookCmd("session-start", "Handle Copilot SessionStart hook event"),
	)
	return cmd
}

// runCopilotHooksInstall writes hooks config to .github/hooks/archcore.json.
func runCopilotHooksInstall(baseDir string) error {
	if !config.DirExists(baseDir) {
		return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
	}

	hooksDir := filepath.Join(baseDir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating .github/hooks/ directory: %w", err)
	}

	hooksPath := filepath.Join(hooksDir, "archcore.json")

	var cfg copilotHooksConfig
	data, err := os.ReadFile(hooksPath)
	if err == nil {
		if unmarshalErr := json.Unmarshal(data, &cfg); unmarshalErr != nil {
			_ = os.WriteFile(hooksPath+".bak", data, 0o644)
			fmt.Println(display.WarnLine(fmt.Sprintf("Corrupted %s backed up, starting fresh", hooksPath)))
			cfg = copilotHooksConfig{}
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", hooksPath, err)
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Hooks == nil {
		cfg.Hooks = make(map[string][]copilotHookEntry)
	}

	for _, ev := range copilotHookEvents {
		if copilotHasCommand(cfg.Hooks[ev.Event], ev.Command) {
			fmt.Println(display.WarnLine(fmt.Sprintf("Copilot: already installed: %s", ev.Event)))
			continue
		}
		cfg.Hooks[ev.Event] = append(cfg.Hooks[ev.Event], copilotHookEntry{
			Type: "command",
			Bash: ev.Command,
		})
		fmt.Println(display.CheckLine(fmt.Sprintf("Copilot: installed hook: %s", ev.Event)))
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(hooksPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", hooksPath, err)
	}

	// Also install MCP config for Copilot.
	if err := installMCPForAgent(baseDir, agents.ByID(agents.Copilot)); err != nil {
		fmt.Println(display.WarnLine(fmt.Sprintf("Copilot MCP install: %v", err)))
	}

	return nil
}

// copilotHasCommand checks whether any entry already contains the exact command.
func copilotHasCommand(entries []copilotHookEntry, command string) bool {
	for _, e := range entries {
		if e.Bash == command {
			return true
		}
	}
	return false
}
