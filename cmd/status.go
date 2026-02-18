package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"archcore-cli/internal/api"
	"archcore-cli/internal/config"
	"archcore-cli/internal/display"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current archcore configuration and connection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if !config.DirExists(cwd) {
				if jsonOutput {
					return jsonError("not initialized — run 'archcore init'")
				}
				fmt.Println(display.FailLine("Not initialized"))
				fmt.Println(display.HintLine("Run 'archcore init' to set up"))
				return nil
			}

			settings, err := config.Load(cwd)
			if err != nil {
				if jsonOutput {
					return jsonError("invalid settings — run 'archcore init'")
				}
				fmt.Println(display.FailLine("Invalid settings"))
				fmt.Println(display.HintLine(err.Error()))
				fmt.Println(display.HintLine("Run 'archcore init' to reconfigure"))
				return nil
			}

			// Check server connectivity when a server URL is configured.
			connected := false
			serverURL := settings.ServerURL()
			if serverURL != "" {
				client := api.NewClient(serverURL)
				connected = client.CheckHealth(ctx) == nil
			}

			if jsonOutput {
				out := map[string]any{
					"sync": settings.Sync,
				}
				if settings.Sync != config.SyncTypeNone {
					out["project_id"] = settings.ProjectID
				}
				if settings.Sync == config.SyncTypeOnPrem {
					out["archcore_url"] = settings.ArchcoreURL
				}
				if serverURL != "" {
					out["connected"] = connected
				}
				return json.NewEncoder(os.Stdout).Encode(out)
			}

			fmt.Println(display.Banner())
			fmt.Println()
			fmt.Println(display.KeyValue("Sync", settings.Sync))
			if settings.Sync != config.SyncTypeNone {
				pidStr := "not connected"
				if settings.ProjectID != nil {
					pidStr = fmt.Sprintf("%d", *settings.ProjectID)
				}
				fmt.Println(display.KeyValue("Project", pidStr))
			}
			if settings.Sync == config.SyncTypeOnPrem {
				fmt.Println(display.KeyValue("Archcore URL", settings.ArchcoreURL))
			}
			fmt.Println()

			if serverURL != "" {
				if connected {
					fmt.Println(display.CheckLine("Server is reachable"))
				} else {
					fmt.Println(display.FailLine("Server is unreachable"))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

func jsonError(msg string) error {
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"error": msg,
	})
}
