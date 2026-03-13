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

// cursorHooksConfig represents the .cursor/hooks.json structure.
type cursorHooksConfig struct {
	Version int                          `json:"version"`
	Hooks   map[string][]cursorHookEntry `json:"hooks"`
}

// cursorHookEntry represents a single hook in Cursor config.
type cursorHookEntry struct {
	Command string `json:"command"`
	Type    string `json:"type"`
}

// cursorHookEvents maps event names to archcore commands.
var cursorHookEvents = []struct {
	Event   string
	Command string
}{
	{"sessionStart", "archcore hooks cursor session-start"},
}

func newHooksCursorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Handle Cursor hook events",
	}
	cmd.AddCommand(
		newSessionStartHookCmd("session-start", "Handle Cursor SessionStart hook event"),
	)
	return cmd
}

// runCursorHooksInstall writes hooks config to .cursor/hooks.json.
func runCursorHooksInstall(baseDir string) error {
	if !config.DirExists(baseDir) {
		return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
	}

	cursorDir := filepath.Join(baseDir, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		return fmt.Errorf("creating .cursor/ directory: %w", err)
	}

	hooksPath := filepath.Join(cursorDir, "hooks.json")

	var cfg cursorHooksConfig
	data, err := os.ReadFile(hooksPath)
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			_ = os.WriteFile(hooksPath+".bak", data, 0o644)
			fmt.Println(display.WarnLine(fmt.Sprintf("Corrupted %s backed up, starting fresh", hooksPath)))
			cfg = cursorHooksConfig{}
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", hooksPath, err)
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Hooks == nil {
		cfg.Hooks = make(map[string][]cursorHookEntry)
	}

	for _, ev := range cursorHookEvents {
		if cursorHasCommand(cfg.Hooks[ev.Event], ev.Command) {
			fmt.Println(display.WarnLine(fmt.Sprintf("Cursor: already installed: %s", ev.Event)))
			continue
		}
		cfg.Hooks[ev.Event] = append(cfg.Hooks[ev.Event], cursorHookEntry{
			Command: ev.Command,
			Type:    "command",
		})
		fmt.Println(display.CheckLine(fmt.Sprintf("Cursor: installed hook: %s", ev.Event)))
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(hooksPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", hooksPath, err)
	}

	// Also install MCP config for Cursor.
	if err := installMCPForAgent(baseDir, agents.ByID(agents.Cursor)); err != nil {
		fmt.Println(display.WarnLine(fmt.Sprintf("Cursor MCP install: %v", err)))
	}

	return nil
}

// cursorHasCommand checks whether any entry already contains the exact command.
func cursorHasCommand(entries []cursorHookEntry, command string) bool {
	for _, e := range entries {
		if e.Command == command {
			return true
		}
	}
	return false
}
