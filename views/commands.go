package views

import (
	"fmt"
	"strconv"
	"strings"
	"tartarus/components"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

var filterRaw string

func NewCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cmds",
		Short: "Manage commands",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	listAllCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all commands",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("filter") && strings.TrimSpace(filterRaw) == "" {
				return fmt.Errorf("the --filter value cannot be empty")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAllCommands(r)
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
		Use:   "v id",
		Short: "View a command by id",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument: id (usage: tartarus cmds v <id>)")
			}

			if len(args) > 1 {
				return fmt.Errorf("expected exactly 1 argument (id), got %d", len(args))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			idStr := args[0]

			ID, err := strconv.ParseUint(idStr, 10, 64)

			if err != nil {
				return fmt.Errorf("invalid id value %q: %w", idStr, err)
			}

			return viewCommand(r, ID)
		},
	}

	createCmd := &cobra.Command{
		Use:   "c <name> <description>",
		Short: "Create a new command",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("expected at least 2 arguments (name and description), got %d", len(args))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			command := repositories.Command{
				Name:        args[0],
				Description: strings.Join(args[1:], " "),
			}

			if err := r.InsertCommand(&command); err != nil {
				return fmt.Errorf("fail to create command: %w", err)
			}

			fmt.Printf("Command created with id %d\n", command.ID)

			return nil
		},
	}

	rootCmd.AddCommand(listAllCmd, viewCmd, createCmd)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) error {
	cmds, err := r.GetCommands(filterRaw)

	if err != nil {
		return fmt.Errorf("fail to list all commands: %w", err)
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

	return nil
}

func viewCommand(r *repositories.Repos, ID uint64) error {
	cmd, err := r.GetCommand(ID)

	if err != nil {
		return fmt.Errorf("fail to view command: %w", err)
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
		return fmt.Errorf("fail to list command items: %w", err)
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

	return nil
}
