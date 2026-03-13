package cmd

import (
	"context"
	"fmt"
	"os"
	// "strings" // re-enable with sync type selector

	"archcore-cli/internal/agents"
	"archcore-cli/internal/api"
	"archcore-cli/internal/config"
	"archcore-cli/internal/display"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type initResult struct {
	serverReachable bool // only meaningful when ServerURL != ""
}

// runInit performs the init logic after prompts have been resolved.
// settings is a fully constructed Settings value (from NewNoneSettings, etc.)
func runInit(ctx context.Context, baseDir string, settings *config.Settings) (*initResult, error) {
	if err := config.InitDir(baseDir); err != nil {
		return nil, fmt.Errorf("creating .archcore/ directory: %w", err)
	}

	result := &initResult{}

	if err := config.Save(baseDir, settings); err != nil {
		return nil, fmt.Errorf("saving settings: %w", err)
	}

	if serverURL := settings.ServerURL(); serverURL != "" {
		client := api.NewClient(serverURL)
		if err := client.CheckHealth(ctx); err != nil {
			return result, fmt.Errorf("cannot reach server at %s: %w", serverURL, err)
		}
		result.serverReachable = true
	}

	return result, nil
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize archcore in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			fmt.Println(display.WelcomeBanner())
			fmt.Println()

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if config.DirExists(cwd) {
				var reinit bool
				err := huh.NewConfirm().
					Title(".archcore/ already exists. Reinitialize?").
					Value(&reinit).
					Run()
				if err != nil {
					return err
				}
				if !reinit {
					fmt.Println(display.Dim.Render("  Cancelled."))
					return nil
				}
			}

			// Sync is temporarily disabled — always use "none".
			// To re-enable, uncomment the block below and remove the hardcoded line.
			settings := config.NewNoneSettings()

			// var syncType string
			// err = huh.NewSelect[string]().
			// 	Title("Select sync option").
			// 	Options(
			// 		huh.NewOption("No sync - store artifacts locally without remote synchronization", config.SyncTypeNone),
			// 		huh.NewOption("Archcore On-Prem - sync with central context platform. Boost your MCP with smart Archcore GraphRAG", config.SyncTypeOnPrem),
			// 	).
			// 	Value(&syncType).
			// 	Run()
			// if err != nil {
			// 	return err
			// }
			//
			// var settings *config.Settings
			// switch syncType {
			// case config.SyncTypeNone:
			// 	settings = config.NewNoneSettings()
			// case config.SyncTypeCloud:
			// 	settings = config.NewCloudSettings()
			// case config.SyncTypeOnPrem:
			// 	serverURL := "http://localhost:8080"
			// 	err = huh.NewInput().
			// 		Title("Archcore URL").
			// 		Value(&serverURL).
			// 		Run()
			// 	if err != nil {
			// 		return err
			// 	}
			// 	serverURL = strings.TrimRight(serverURL, "/")
			// 	settings = config.NewOnPremSettings(serverURL)
			// default:
			// 	return fmt.Errorf("unsupported sync type: %q", syncType)
			// }

			result, err := runInit(ctx, cwd, settings)
			if err != nil && result != nil {
				// Server unreachable — soft fail
				fmt.Println(display.FailLine(err.Error()))
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Println(display.CheckLine("Created .archcore/ directory"))
			if result.serverReachable {
				fmt.Println(display.CheckLine("Server is reachable"))
			}
			fmt.Println(display.CheckLine("Settings saved to .archcore/settings.json"))

			// Auto-detect agents and install hooks + MCP config for all found.
			detected := agents.Detect(cwd)
			if len(detected) == 0 {
				detected = []*agents.Agent{agents.ByID(agents.ClaudeCode)}
			}
			for _, agent := range detected {
				if err := installHooksForAgent(cwd, agent); err != nil {
					fmt.Println(display.WarnLine(fmt.Sprintf("%s hooks: %v", agent.DisplayName, err)))
				}
				if err := installMCPForAgent(cwd, agent); err != nil {
					fmt.Println(display.WarnLine(fmt.Sprintf("%s MCP: %v", agent.DisplayName, err)))
				}
			}

			fmt.Println()
			fmt.Println(display.Success.Render("  Ready! Run 'archcore status' to verify."))
			return nil
		},
	}
}
