package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

			issues := 0

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			// 1. .archcore/ directory exists
			if !config.DirExists(cwd) {
				fmt.Println(display.FailLine(".archcore/ directory not found"))
				fmt.Println(display.HintLine("Run 'archcore init' to set up"))
				return nil
			}
			fmt.Println(display.CheckLine(".archcore/ directory exists"))

			// 2. Subdirectories exist
			for _, sub := range []string{"vision", "knowledge", "experience"} {
				p := filepath.Join(cwd, ".archcore", sub)
				if info, err := os.Stat(p); err != nil || !info.IsDir() {
					issues++
					fmt.Println(display.FailLine(fmt.Sprintf(".archcore/%s/ missing", sub)))
					fmt.Println(display.HintLine("Run 'archcore init' to recreate"))
				} else {
					fmt.Println(display.CheckLine(fmt.Sprintf(".archcore/%s/ exists", sub)))
				}
			}

			// 3. Settings file valid
			settings, err := config.Load(cwd)
			if err != nil {
				issues++
				fmt.Println(display.FailLine("Settings file missing or invalid"))
				fmt.Println(display.HintLine(err.Error()))
				fmt.Println(display.HintLine("Run 'archcore init' to reconfigure"))
			} else {
				fmt.Println(display.CheckLine(fmt.Sprintf("Settings valid (sync: %s)", settings.Sync)))
			}

			// 4. Server reachable (only when a server URL is configured)
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
