package cmd

import (
	"fmt"
	"os"

	"archcore-cli/internal/agents"
	"archcore-cli/internal/config"
	"archcore-cli/internal/display"
	mcpserver "archcore-cli/internal/mcp"

	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP stdio server for archcore documents",
		Long:  "Starts an MCP (Model Context Protocol) stdio server that exposes archcore document tools.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, display.WelcomeBanner())
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, display.Dim.Render("  MCP server running on stdio..."))

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if !config.DirExists(cwd) {
				return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
			}
			return mcpserver.RunStdio(cwd)
		},
	}

	cmd.AddCommand(newMCPInstallCmd())
	return cmd
}

func newMCPInstallCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install MCP server config for coding agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if !config.DirExists(cwd) {
				return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
			}

			if agentFlag != "" {
				return runMCPInstallForAgent(cwd, agents.AgentID(agentFlag))
			}
			return runMCPInstallAutoDetect(cwd)
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "install for a specific agent (e.g. cursor, gemini-cli)")
	return cmd
}

// runMCPInstallForAgent installs MCP config for a specific agent.
func runMCPInstallForAgent(baseDir string, id agents.AgentID) error {
	agent := agents.ByID(id)
	if agent == nil {
		return fmt.Errorf("unknown agent %q — valid agents: %v", id, agents.AllIDs())
	}
	return installMCPForAgent(baseDir, agent)
}

// runMCPInstallAutoDetect detects agents and installs MCP config for all found.
// Falls back to Claude Code if no agents detected.
func runMCPInstallAutoDetect(baseDir string) error {
	detected := agents.Detect(baseDir)
	if len(detected) == 0 {
		detected = []*agents.Agent{agents.ByID(agents.ClaudeCode)}
	}

	for _, agent := range detected {
		if err := installMCPForAgent(baseDir, agent); err != nil {
			fmt.Println(display.WarnLine(fmt.Sprintf("%s MCP install: %v", agent.DisplayName, err)))
		}
	}
	return nil
}

// installMCPForAgent installs MCP config for a single agent.
func installMCPForAgent(baseDir string, agent *agents.Agent) error {
	if agent.ManualMCPInstallHint != "" {
		fmt.Println(display.WarnLine(fmt.Sprintf("%s: %s", agent.DisplayName, agent.ManualMCPInstallHint)))
		return nil
	}

	if err := agent.WriteMCPConfig(baseDir); err != nil {
		return err
	}
	configPath := ""
	if agent.MCPConfigPath != nil {
		configPath = agent.MCPConfigPath(baseDir)
	}
	if configPath != "" {
		fmt.Println(display.CheckLine(fmt.Sprintf("Installed MCP config for %s", agent.DisplayName)))
	}
	return nil
}

// runMCPInstall is the legacy function for backward compatibility.
// Used by runHooksInstall to also install MCP (Claude Code only).
func runMCPInstall(baseDir string) error {
	if !config.DirExists(baseDir) {
		return fmt.Errorf(".archcore/ not found — run 'archcore init' first")
	}
	agent := agents.ByID(agents.ClaudeCode)
	if agent == nil {
		return fmt.Errorf("claude-code agent not found in registry")
	}
	return agent.WriteMCPConfig(baseDir)
}
