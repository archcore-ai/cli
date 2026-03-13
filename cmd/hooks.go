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
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// hookEntry represents a single hook command configuration.
type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// hookMatcher represents a matcher with its hooks array.
type hookMatcher struct {
	Matcher string      `json:"matcher"`
	Hooks   []hookEntry `json:"hooks"`
}

// archcoreHooks defines the hooks we install for Claude Code, in deterministic order.
var archcoreHooks = []struct {
	Event   string
	Matcher hookMatcher
}{
	{"SessionStart", hookMatcher{Matcher: "", Hooks: []hookEntry{{Type: "command", Command: "archcore hooks claude-code session-start"}}}},
}

func newHooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage agent hooks integration",
	}
	cmd.AddCommand(
		newHooksInstallCmd(),
		newHooksClaudeCodeCmd(),
		newHooksCursorCmd(),
		newHooksGeminiCLICmd(),
		newHooksCopilotCmd(),
	)
	return cmd
}

func newHooksInstallCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install archcore hooks for coding agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if !config.DirExists(cwd) {
				return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
			}

			if agentFlag != "" {
				return runHooksInstallForAgent(cwd, agents.AgentID(agentFlag))
			}
			return runHooksInstallAutoDetect(cwd)
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "install hooks for a specific agent (e.g. cursor, gemini-cli)")
	return cmd
}

// hooksInstallers maps agent IDs to their hooks install functions.
// This is the single source of truth for which agents support hooks.
var hooksInstallers = map[agents.AgentID]func(string) error{
	agents.ClaudeCode: runHooksInstall,
	agents.Cursor:     runCursorHooksInstall,
	agents.GeminiCLI:  runGeminiCLIHooksInstall,
	agents.Copilot:    runCopilotHooksInstall,
}

// installHooksForAgent installs hooks for a single agent that supports them.
// Returns nil if the agent doesn't support hooks.
func installHooksForAgent(baseDir string, agent *agents.Agent) error {
	installer, ok := hooksInstallers[agent.ID]
	if !ok {
		return nil
	}
	return installer(baseDir)
}

// runHooksInstallForAgent installs hooks for a specific agent by ID.
func runHooksInstallForAgent(baseDir string, id agents.AgentID) error {
	agent := agents.ByID(id)
	if agent == nil {
		return fmt.Errorf("unknown agent %q — valid agents: %v", id, agents.AllIDs())
	}
	if err := installHooksForAgent(baseDir, agent); err != nil {
		return err
	}
	if _, ok := hooksInstallers[id]; !ok {
		fmt.Println(display.WarnLine(fmt.Sprintf("%s does not support hooks", agent.DisplayName)))
	}
	return nil
}

// runHooksInstallAutoDetect detects agents and installs hooks for all that support them.
// Falls back to Claude Code if no agents detected.
func runHooksInstallAutoDetect(baseDir string) error {
	detected := agents.Detect(baseDir)
	if len(detected) == 0 {
		detected = []*agents.Agent{agents.ByID(agents.ClaudeCode)}
	}

	for _, agent := range detected {
		if err := installHooksForAgent(baseDir, agent); err != nil {
			fmt.Println(display.WarnLine(fmt.Sprintf("%s hooks: %v", agent.DisplayName, err)))
		}
	}

	// Also install MCP for all detected agents.
	if err := runMCPInstallAutoDetect(baseDir); err != nil {
		fmt.Println(display.WarnLine(fmt.Sprintf("MCP auto-detect install: %v", err)))
	}

	return nil
}

// runHooksInstall installs Claude Code hooks into .claude/settings.json.
func runHooksInstall(baseDir string) error {
	if !config.DirExists(baseDir) {
		return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
	}

	claudeDir := filepath.Join(baseDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("creating .claude/ directory: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Read existing settings or start empty (ordered to preserve key order).
	raw := orderedmap.New[string, json.RawMessage]()
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if unmarshalErr := json.Unmarshal(data, raw); unmarshalErr != nil {
			_ = os.WriteFile(settingsPath+".bak", data, 0o644)
			fmt.Println(display.WarnLine(fmt.Sprintf("Corrupted %s backed up, starting fresh", settingsPath)))
			raw = orderedmap.New[string, json.RawMessage]()
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", settingsPath, err)
	}

	// Parse existing hooks section (ordered to preserve event order).
	hooks := orderedmap.New[string, []hookMatcher]()
	if hooksRaw, ok := raw.Get("hooks"); ok {
		if err := json.Unmarshal(hooksRaw, hooks); err != nil {
			return fmt.Errorf("parsing hooks section: %w", err)
		}
	}

	// Merge each archcore hook in deterministic order.
	for _, h := range archcoreHooks {
		existing, _ := hooks.Get(h.Event)
		if hasCommand(existing, h.Matcher.Hooks[0].Command) {
			fmt.Println(display.WarnLine(fmt.Sprintf("Already installed: %s", h.Event)))
			continue
		}
		hooks.Set(h.Event, append(existing, h.Matcher))
		fmt.Println(display.CheckLine(fmt.Sprintf("Installed hook: %s", h.Event)))
	}

	// Write back.
	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	raw.Set("hooks", json.RawMessage(hooksJSON))

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", settingsPath, err)
	}

	// Also install MCP config (soft error — warn but don't fail).
	if err := runMCPInstall(baseDir); err != nil {
		fmt.Println(display.WarnLine(fmt.Sprintf("MCP install: %v", err)))
	}

	return nil
}

// hasCommand checks whether any matcher already contains the exact command.
func hasCommand(matchers []hookMatcher, command string) bool {
	for _, m := range matchers {
		for _, h := range m.Hooks {
			if h.Command == command {
				return true
			}
		}
	}
	return false
}
