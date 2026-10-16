package views

import (
	"errors"
	"fmt"
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

	rootCmd.AddCommand(
		listAllCommands(r),
		viewCommand(r),
		createCommand(r),
		editCommand(r),
	)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all commands",
		Args:  cobra.NoArgs,
		Long: `List every command stored in the database as a table with its ID, name and
description.

What it does:
  Reads all commands from the store and prints them ordered by ID. When a
  filter is given, only commands whose name or description contains the filter
  as a literal substring are shown (SQLite LIKE, ASCII case-insensitive).

What it expects:
  No positional arguments. One optional flag:
    -f, --filter <text>   literal substring matched against the name or
                          description. An empty or whitespace-only value is
                          rejected.

What it returns:
  A table with the columns ID, Name and Description. If nothing matches, it
  prints "No commands found" instead. It exits non-zero when the filter is
  invalid or when the database query fails.

Examples:
  tartarus cmds ls
  tartarus cmds ls --filter deploy
  tartarus cmds ls -f "build"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("filter") && strings.TrimSpace(filterRaw) == "" {
				return fmt.Errorf("the --filter value cannot be empty")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			commands, err := r.GetCommands(filterRaw)

			if err != nil {
				return fmt.Errorf("fail to get commands: %w", err)
			}

			var rows [][]string

			for _, command := range commands {
				rows = append(rows, []string{
					command.Name,
					command.Description,
				})
			}

			table := components.Table{
				Title:    "List all commands",
				EmptyMsg: "No commands found",
				Rows:     rows,
				Header: []string{
					"Name",
					"Description",
				},
			}

			table.Render()

			return nil
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

func viewCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "v <name>",
		Short: "View a command by name",
		Long: `Show a single command and its items.

What it does:
  Reads the command with the given ID and prints its details (ID, name and
  description), followed by a second table listing every item that belongs to
  it (ID, script and description).

What it expects:
  Exactly one positional argument: the numeric command ID.
    v <id>   the ID of the command to display

What it returns:
  A "Command Details" table with the fields ID, Name and Description, then a
  "Command Items" table with the columns ID, Script and Description. If the
  command has no items, it prints "No items found" for the second table. It
  exits non-zero when the ID is missing, not numeric, or when either database
  query fails. A command that does not exist is displayed as an empty row set.

Examples:
  tartarus cmds v cmdname`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly 1 argument (name), got %d", len(args))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := r.GetCommand(args[0])

			if err != nil {
				return fmt.Errorf("fail to get command: %w", err)
			}

			cmdTable := components.Table{
				Title:    "Command",
				EmptyMsg: "No command found",
				Rows: [][]string{
					{
						"Name:",
						command.Name,
					},
					{
						"Description:",
						command.Description,
					},
				},
			}

			cmdTable.Render()

			fmt.Println()

			scripts, err := r.GetCommandScripts(command.ID)

			if err != nil {
				return fmt.Errorf("fail to get command scripts: %w", err)
			}

			scriptRows := [][]string{}

			for _, script := range scripts {
				scriptRows = append(scriptRows, []string{
					script.Script,
					script.Description,
				})
			}

			cmdScripts := components.Table{
				Title:    "Command scripts",
				EmptyMsg: "No command scripts found",
				Header:   []string{"Script", "Description"},
				Rows:     scriptRows,
			}

			cmdScripts.Render()

			return nil
		},
	}

	return rootCmd
}

func createCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "c",
		Short: "Create a new command",
		Long: `Create a new command.

What it does:
  Inserts a new command with the given name and description, then prints the
  ID assigned to it.

What it expects:
  Exactly two positional arguments:
    c <name> <description>   the command name and its description

  The name is required and must be unique; an empty or already-used name is
  rejected.

What it returns:
  On success it prints "Command created with id <id>". It exits non-zero when
  the argument count is wrong or when the insert fails (for example, a missing
  or duplicate name).

Examples:
  tartarus cmds c deploy "Deploy the app to production"
  tartarus cmds c build "Compile the project"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			components.PromptTitle("Create a new command")

			command.Name, err = components.Prompt(components.PromptOpts{
				Label:   "Name",
				Default: "",
				Validate: func(name string) error {
					return repositories.ValidateCommandName(name)
				},
			})

			if err != nil {
				return err
			}

			command.Description, err = components.Prompt(components.PromptOpts{
				Label:   "Description",
				Default: "",
				Validate: func(description string) error {
					if description == "" {
						return errors.New("description is required")
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			err = r.InsertCommand(&command)

			if err != nil {
				return fmt.Errorf("fail to insert command: %w", err)
			}

			fmt.Printf("\nCommand successfully created %s\n", command.Name)

			return nil
		},
	}

	return rootCmd
}

func editCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "e <name>",
		Short: "Edit an existing command by name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly 1 argument (name), got %d", len(args))
			}

			return nil
		},
		Long: `Edit an existing command.

What it does:
  Loads the command with the given ID and updates its name and description.
  An empty argument is ignored, keeping the current value, so you can change
  only one field at a time without retyping the other.

What it expects:
  Exactly three positional arguments:
    e <id> <name> <description>   the ID of the command to edit, plus its new
                                  name and description. Pass "" to keep the
                                  current value.

  The name, when changed, must be unique; an already-used name is rejected.

What it returns:
  On success it prints "Command edited with id <id>". It exits non-zero when
  the argument count is wrong, when the ID is not numeric or does not exist,
  or when the update fails (for example, a duplicate name).

Examples:
  tartarus cmds e 1 deploy "Deploy the app to staging"
  tartarus cmds e 1 "" "Only a new description"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return err
			}

			command.Name, err = components.Prompt(components.PromptOpts{
				Label:   "Name",
				Default: command.Name,
				Validate: func(name string) error {
					return repositories.ValidateCommandName(name)
				},
			})

			if err != nil {
				return err
			}

			command.Description, err = components.Prompt(components.PromptOpts{
				Label:   "Description",
				Default: command.Description,
				Validate: func(description string) error {
					if description == "" {
						return errors.New("description is required")
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			err = r.UpdateCommand(&command)

			if err != nil {
				return fmt.Errorf("fail to update command: %w", err)
			}

			fmt.Printf("\nCommand successfully edited %s\n", command.Name)

			return nil
		},
	}

	return rootCmd
}

// func deleteCommand(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "d <id>",
// 		Short: "Delete an existing command by id",
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) < 1 || len(args) > 1 {
// 				return fmt.Errorf("expected 1 argument (id), got %d", len(args))
// 			}

// 			return nil
// 		},
// 		Long: `Delete an existing command.

// What it does:
//   Removes the command with the given ID together with all of its items.

// What it expects:
//   Exactly one positional argument:
//     d <id>   the ID of the command to delete

// What it returns:
//   On success it prints "Command deleted with id <id>". It exits non-zero when
//   the argument count is wrong, when the ID is not numeric, or when no command
//   has that ID.

// Examples:
//   tartarus cmds d 1
//   tartarus cmds d 42`,
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			idStr := args[0]

// 			ID, err := strconv.ParseUint(idStr, 10, 64)

// 			if err != nil {
// 				return fmt.Errorf("invalid id value %q", idStr)
// 			}

// 			command, err := r.GetCommand(ID)

// 			if err != nil {
// 				return fmt.Errorf("fail to get command: %w", err)
// 			}

// 			if command.ID == 0 {
// 				return fmt.Errorf("command not found with id %d", ID)
// 			}

// 			if err := r.DeleteCommand(ID); err != nil {
// 				return fmt.Errorf("fail to delete command: %w", err)
// 			}

// 			fmt.Printf("Command deleted with id %d\n", command.ID)

// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

// func appendCommandItem(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "ap <commandId> <name> <script> <description>",
// 		Short: "Append a new item to an existing command",
// 		Args:  cobra.NoArgs,
// 		Long:  ``,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

// func changeCommandItem(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "ch <commandId> <commandItemId> <name> <script> <description>",
// 		Short: "Change an existing item of an existing command",
// 		Args:  cobra.NoArgs,
// 		Long:  ``,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

// func removeCommandItem(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "rm <commandItemId>",
// 		Short: "Remove an existing item of an existing command",
// 		Args:  cobra.NoArgs,
// 		Long:  ``,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return nil
// 		},
// 	}

// 	return rootCmd
// }
