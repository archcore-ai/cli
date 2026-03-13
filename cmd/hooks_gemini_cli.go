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

// geminiHookEvents maps Gemini CLI event names to archcore commands.
var geminiHookEvents = []struct {
	Event   string
	Command string
}{
	{"SessionStart", "archcore hooks gemini-cli session-start"},
}

func newHooksGeminiCLICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gemini-cli",
		Short: "Handle Gemini CLI hook events",
	}
	cmd.AddCommand(
		newSessionStartHookCmd("session-start", "Handle Gemini CLI SessionStart hook event"),
	)
	return cmd
}

// runGeminiCLIHooksInstall writes hooks config into .gemini/settings.json.
func runGeminiCLIHooksInstall(baseDir string) error {
	if !config.DirExists(baseDir) {
		return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
	}

	geminiDir := filepath.Join(baseDir, ".gemini")
	if err := os.MkdirAll(geminiDir, 0o755); err != nil {
		return fmt.Errorf("creating .gemini/ directory: %w", err)
	}

	settingsPath := filepath.Join(geminiDir, "settings.json")

	var raw map[string]json.RawMessage
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &raw); err != nil {
			_ = os.WriteFile(settingsPath+".bak", data, 0o644)
			fmt.Println(display.WarnLine(fmt.Sprintf("Corrupted %s backed up, starting fresh", settingsPath)))
			raw = make(map[string]json.RawMessage)
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		raw = make(map[string]json.RawMessage)
	} else {
		return fmt.Errorf("reading %s: %w", settingsPath, err)
	}

	// Parse existing hooks section.
	var hooks map[string][]hookMatcher
	if hooksRaw, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
			return fmt.Errorf("parsing hooks section: %w", err)
		}
	} else {
		hooks = make(map[string][]hookMatcher)
	}

	for _, ev := range geminiHookEvents {
		entry := hookMatcher{
			Matcher: "",
			Hooks:   []hookEntry{{Type: "command", Command: ev.Command}},
		}
		if hasCommand(hooks[ev.Event], ev.Command) {
			fmt.Println(display.WarnLine(fmt.Sprintf("Gemini CLI: already installed: %s", ev.Event)))
			continue
		}
		hooks[ev.Event] = append(hooks[ev.Event], entry)
		fmt.Println(display.CheckLine(fmt.Sprintf("Gemini CLI: installed hook: %s", ev.Event)))
	}

	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	raw["hooks"] = json.RawMessage(hooksJSON)

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", settingsPath, err)
	}

	// Also install MCP config for Gemini CLI.
	if err := installMCPForAgent(baseDir, agents.ByID(agents.GeminiCLI)); err != nil {
		fmt.Println(display.WarnLine(fmt.Sprintf("Gemini CLI MCP install: %v", err)))
	}

	return nil
}
