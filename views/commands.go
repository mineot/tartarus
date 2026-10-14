package views

import (
	"fmt"
	"strconv"
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

	listAllCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all commands",
		Run: func(cmd *cobra.Command, args []string) {
			listAllCommands(r)
		},
	}

	listAllCmd.Flags().StringVarP(
		&filterRaw,
		"filter",
		"f",
		"",
		"Filter commands by name or description. Example: -f <name> or --filter <description>",
	)

	viewCmd := &cobra.Command{
		Use:   "v",
		Short: "View a command by id. Exemplo: v {id}",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			idStr := args[0]

			ID, err := strconv.ParseUint(idStr, 10, 64)

			if err != nil {
				fmt.Printf("Invalid id value: %s", idStr)
			}

			viewCommand(r, ID)
		},
	}

	rootCmd.AddCommand(listAllCmd)
	rootCmd.AddCommand(viewCmd)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) {
	cmds, err := r.GetCommands(filterRaw)

	if err != nil {
		fmt.Printf("Fail to list all commands: %v", err)
	}

	list := components.List{
		Title:    "List of Commands",
		EmptyMsg: "No commands found",
		Headers: []string{
			"ID",
			"Name",
			"Description",
		},
		Rows: [][]string{},
	}

	for _, cmd := range cmds {
		list.Rows = append(list.Rows, []string{
			fmt.Sprintf("%d", cmd.ID),
			cmd.Name,
			cmd.Description,
		})
	}

	list.Print()
}

func viewCommand(r *repositories.Repos, ID uint64) {
	cmd, err := r.GetCommand(ID)

	if err != nil {
		panic(err)
	}

	cmdList := components.List{
		Title:    "Command Details",
		EmptyMsg: "Command not found",
		Rows:     [][]string{},
		Headers:  []string{"Field", "Value"},
	}
	cmdList.Rows = append(cmdList.Rows, []string{"ID", fmt.Sprintf("%d", cmd.ID)})
	cmdList.Rows = append(cmdList.Rows, []string{"Name", cmd.Name})
	cmdList.Rows = append(cmdList.Rows, []string{"Description", cmd.Description})
	cmdList.Print()

	fmt.Println()

	items, err := r.GetCommandItems(ID)

	if err != nil {
		panic(err)
	}

	itemList := components.List{
		Title:    "Command Items",
		EmptyMsg: "No items found",
		Rows:     [][]string{},
		Headers:  []string{"ID", "Script", "Description"},
	}

	for _, item := range items {
		itemList.Rows = append(itemList.Rows, []string{
			fmt.Sprintf("%d", item.ID),
			item.Script,
			item.Description,
		})
	}

	itemList.Print()
}
