package views

import (
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

func New(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "tartarus",
		Short:         RenderLogo("Organize and automate your terminal workflows"),
		Version:       "2.0.0",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(NewCommands(r))

	return rootCmd
}
