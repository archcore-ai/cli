package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "archcore",
		Short:         "Archcore CLI — context engineering platform",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		newInitCmd(),
		newCreateCmd(),
		newStatusCmd(),
		newConfigCmd(),
		newDoctorCmd(),
	)

	return root
}
