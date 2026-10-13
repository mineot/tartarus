package views

import (
	"fmt"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

func NewCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cmds",
		Short: "Manage commands",
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
		Short: "List all commands",
		Run: func(cmd *cobra.Command, args []string) {
			cmds, _ := r.GetCommands("10")
			fmt.Println(cmds)
		},
	}

	return rootCmd
}
