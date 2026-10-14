package views

import (
	"fmt"
	"tartarus/components"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

var filterRaw string
var commandId uint64

func NewCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cmds",
		Short: "Manage commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(listAllCmd(r))

	return rootCmd
}

func listAllCmd(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all commands",
		Run: func(cmd *cobra.Command, args []string) {
			listAllCommands(r)
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

func listAllCommands(r *repositories.Repos) {
	cmds, err := r.GetCommands(filterRaw)

	if err != nil {
		fmt.Printf("Fail to list all commands: %v", err)
	}

	list := components.List{}
	list.Title("List of Commands")
	list.EmptyMsg("No commands found!")

	for _, cmd := range cmds {
		list.AppendItem(
			cmd.ID,
			cmd.Name,
			cmd.Description,
		)
	}

	list.Print()
}

// func viewCommand(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "vw",
// 		Short: "View a command by id",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			if commandId == 0 {
// 				fmt.Print("Please provide a command id")
// 			}

// 			cmd, err := r.GetCommand(commandId)

// 			if err != nil {
// 				panic(err)
// 			}
// 		},
// 	}

// 	rootCmd.Flags().StringVarP(
// 		&filterRaw,
// 		"filter",
// 		"f",
// 		"",
// 		"Filter commands by name or description. Example: -f <name> or --filter <description>",
// 	)

// 	return rootCmd
// }
