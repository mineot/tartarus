package views

import (
	"tartarus/components"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

func NewCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cmds",
		Short: RenderLogo("Manage commands"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(listAllCommands(r))

	return rootCmd
}

func listAllCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ls",
		Short: RenderLogo("List all commands"),
		Run: func(cmd *cobra.Command, args []string) {
			components.CommandsList(r)
		},
	}

	return rootCmd
}
