package views

import (
	"fmt"
	"tartarus/components"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

var filterRaw string

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
			cmds, err := r.GetCommands(filterRaw)

			if err != nil {
				panic(err)
			}

			list := components.List{}
			list.Title("List of Commands")
			list.EmptyMsg("No commands found!")

			for _, cmd := range cmds {
				list.AppendItem(
					fmt.Sprintf("%s - ID: %d", cmd.Name, cmd.ID),
					cmd.Description,
				)
			}

			list.Print()
		},
	}

	rootCmd.Flags().StringVarP(
		&filterRaw,
		"filter",
		"f",
		"",
		"Filter commands by name or description. Example: -f <name> or --filter <description>",
	)

	return rootCmd
}
