package cmd

import (
	"fmt"
	"os"

	"archcore-cli/internal/api"
	"archcore-cli/internal/config"
	"archcore-cli/internal/display"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check your archcore setup for issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			fmt.Println(display.Banner())
			fmt.Println()

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			// Run validation checks (structure + documents).
			issues := runValidateChecks(cwd, false)

			// Early return if .archcore/ doesn't exist — settings and
			// server checks depend on it.
			if !config.DirExists(cwd) {
				return nil
			}

			// Settings file valid.
			settings, err := config.Load(cwd)
			if err != nil {
				issues++
				fmt.Println(display.FailLine("Settings file missing or invalid"))
				fmt.Println(display.HintLine(err.Error()))
				fmt.Println(display.HintLine("Run 'archcore init' to reconfigure"))
			} else {
				fmt.Println(display.CheckLine(fmt.Sprintf("Settings valid (sync: %s)", settings.Sync)))
			}

			// Server reachable (only when a server URL is configured).
			if settings != nil {
				if serverURL := settings.ServerURL(); serverURL != "" {
					client := api.NewClient(serverURL)
					if err := client.CheckHealth(ctx); err != nil {
						issues++
						fmt.Println(display.FailLine("Server unreachable at " + serverURL))
						fmt.Println(display.HintLine(err.Error()))
					} else {
						fmt.Println(display.CheckLine("Server is reachable"))
					}
				}
			}

			fmt.Println()
			if issues == 0 {
				fmt.Println(display.Success.Render("  All checks passed!"))
			} else {
				fmt.Println(display.Warn.Render(fmt.Sprintf("  %d issue(s) found", issues)))
			}

			return nil
		},
	}
}
