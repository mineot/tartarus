package views

import (
	"context"
	"errors"
	"fmt"
	"tartarus/components"
	"tartarus/repositories"
	"tartarus/themes"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

const clearScreen = "\x1b[2J\x1b[H"

var filterRaw string

type commandAction int

const (
	actionNone commandAction = iota
	actionNew
	actionEdit
	actionScripts
)

func NewCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cmds",
		Short: "Manage commands",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for {
				fmt.Print(clearScreen)
				themes.RenderLogo()

				action, err := tableCommands(r)

				if err != nil {
					return err
				}

				if action != actionNew {
					return nil
				}

				fmt.Print(clearScreen)

				back, err := formComponent(cmd.Context(), r)

				if err != nil {
					return err
				}

				if !back {
					return nil
				}
			}
		},
	}

	return rootCmd
}

func tableCommands(r *repositories.Repos) (commandAction, error) {
	commands, err := r.GetCommands("")

	if err != nil {
		return actionNone, err
	}

	rows := [][]string{}

	for _, command := range commands {
		rows = append(rows, []string{
			command.Name,
			command.Description,
		})
	}

	var action commandAction

	err = components.Table(components.TableOpts{
		Title:    "Commands",
		EmptyMsg: "No commands found.",
		Shortcuts: []components.Shortcut{
			{
				Key:   "n",
				Label: "new command",
				Run: func(selected int) (quit bool, err error) {
					action = actionNew
					return true, nil
				},
			},
			{
				Key:   "e",
				Label: "edit command",
				Run: func(selected int) (quit bool, err error) {
					return false, nil
				},
			},
			{
				Key:   "s",
				Label: "scripts",
				Run: func(selected int) (quit bool, err error) {
					return false, nil
				},
			},
		},
		Headers: []string{"Name", "Description"},
		Rows:    rows,
	})

	if err != nil {
		return actionNone, err
	}

	return action, nil
}

func formComponent(ctx context.Context, r *repositories.Repos) (bool, error) {
	var name, description string
	back := false

	_, err := components.Form(ctx, components.FormOpts{
		Title: "New command",
		Fields: []huh.Field{
			huh.NewInput().Key("name").Title("Name").Value(&name),
			huh.NewInput().Key("description").Title("Description").Value(&description),
		},
		Shortcuts: []components.FormShortcut{
			{
				Key:   "esc",
				Label: "back",
				Run: func(focus string) (quit bool, err error) {
					back = true
					return true, nil
				},
			},
		},
	})

	if errors.Is(err, huh.ErrUserAborted) {
		return back, nil
	}

	if err != nil {
		return false, err
	}

	command := repositories.Command{
		Name:        name,
		Description: description,
	}

	if err := r.InsertCommand(&command); err != nil {
		return false, err
	}

	return true, nil
}

// func listAllCommands(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "ls",
// 		Short: "List all commands",
// 		Args:  cobra.NoArgs,
// 		Long: `List every command stored in the database as a table with its ID, name and
// description.

// What it does:
//   Reads all commands from the store and prints them ordered by ID. When a
//   filter is given, only commands whose name or description contains the filter
//   as a literal substring are shown (SQLite LIKE, ASCII case-insensitive).

// What it expects:
//   No positional arguments. One optional flag:
//     -f, --filter <text>   literal substring matched against the name or
//                           description. An empty or whitespace-only value is
//                           rejected.

// What it returns:
//   A table with the columns ID, Name and Description. If nothing matches, it
//   prints "No commands found" instead. It exits non-zero when the filter is
//   invalid or when the database query fails.

// Examples:
//   tartarus cmds ls
//   tartarus cmds ls --filter deploy
//   tartarus cmds ls -f "build"`,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			if cmd.Flags().Changed("filter") && strings.TrimSpace(filterRaw) == "" {
// 				return fmt.Errorf("the --filter value cannot be empty")
// 			}

// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cmds, err := r.GetCommands(filterRaw)

// 			if err != nil {
// 				return fmt.Errorf("fail to list all commands: %w", err)
// 			}

// 			list := components.List{
// 				Title:    "List of Commands",
// 				EmptyMsg: "No commands found",
// 				Headers: []string{
// 					"ID",
// 					"Name",
// 					"Description",
// 				},
// 				Rows: [][]string{},
// 			}

// 			for _, cmd := range cmds {
// 				list.Rows = append(list.Rows, []string{
// 					fmt.Sprintf("%d", cmd.ID),
// 					cmd.Name,
// 					cmd.Description,
// 				})
// 			}

// 			list.Print()

// 			return nil
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

// func viewCommand(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "v <id>",
// 		Short: "View a command by id",
// 		Long: `Show a single command and its items.

// What it does:
//   Reads the command with the given ID and prints its details (ID, name and
//   description), followed by a second table listing every item that belongs to
//   it (ID, script and description).

// What it expects:
//   Exactly one positional argument: the numeric command ID.
//     v <id>   the ID of the command to display

// What it returns:
//   A "Command Details" table with the fields ID, Name and Description, then a
//   "Command Items" table with the columns ID, Script and Description. If the
//   command has no items, it prints "No items found" for the second table. It
//   exits non-zero when the ID is missing, not numeric, or when either database
//   query fails. A command that does not exist is displayed as an empty row set.

// Examples:
//   tartarus cmds v 1
//   tartarus cmds v 42`,
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) == 0 {
// 				return fmt.Errorf("missing required argument: id (usage: tartarus cmds v <id>)")
// 			}

// 			if len(args) > 1 {
// 				return fmt.Errorf("expected exactly 1 argument (id), got %d", len(args))
// 			}

// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			idStr := args[0]

// 			ID, err := strconv.ParseUint(idStr, 10, 64)

// 			if err != nil {
// 				return fmt.Errorf("invalid id value %q", idStr)
// 			}

// 			command, err := r.GetCommand(ID)

// 			if err != nil {
// 				return fmt.Errorf("fail to view command: %w", err)
// 			}

// 			cmdList := components.List{
// 				Title:    "Command Details",
// 				EmptyMsg: "Command not found",
// 				Rows:     [][]string{},
// 				Headers:  []string{"Field", "Value"},
// 			}

// 			cmdList.Rows = append(cmdList.Rows, []string{"ID", fmt.Sprintf("%d", command.ID)})
// 			cmdList.Rows = append(cmdList.Rows, []string{"Name", command.Name})
// 			cmdList.Rows = append(cmdList.Rows, []string{"Description", command.Description})
// 			cmdList.Print()

// 			fmt.Println()

// 			items, err := r.GetCommandScripts(ID)

// 			if err != nil {
// 				return fmt.Errorf("fail to list command items: %w", err)
// 			}

// 			itemList := components.List{
// 				Title:    "Command Items",
// 				EmptyMsg: "No items found",
// 				Rows:     [][]string{},
// 				Headers:  []string{"ID", "Script", "Description"},
// 			}

// 			for _, item := range items {
// 				itemList.Rows = append(itemList.Rows, []string{
// 					fmt.Sprintf("%d", item.ID),
// 					item.Script,
// 					item.Description,
// 				})
// 			}

// 			itemList.Print()

// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

// func createCommand(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "c <name> <description>",
// 		Short: "Create a new command",
// 		Long: `Create a new command.

// What it does:
//   Inserts a new command with the given name and description, then prints the
//   ID assigned to it.

// What it expects:
//   Exactly two positional arguments:
//     c <name> <description>   the command name and its description

//   The name is required and must be unique; an empty or already-used name is
//   rejected.

// What it returns:
//   On success it prints "Command created with id <id>". It exits non-zero when
//   the argument count is wrong or when the insert fails (for example, a missing
//   or duplicate name).

// Examples:
//   tartarus cmds c deploy "Deploy the app to production"
//   tartarus cmds c build "Compile the project"`,
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) < 2 || len(args) > 2 {
// 				return fmt.Errorf("expected 2 arguments (name and description), got %d", len(args))
// 			}

// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			command := repositories.Command{
// 				Name:        args[0],
// 				Description: args[1],
// 			}

// 			if err := r.InsertCommand(&command); err != nil {
// 				return fmt.Errorf("fail to create command: %w", err)
// 			}

// 			fmt.Printf("Command created with id %d\n", command.ID)

// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

// func editCommand(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "e <id> <name> <description>",
// 		Short: "Edit an existing command by id",
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) < 3 || len(args) > 3 {
// 				return fmt.Errorf("expected 3 arguments (id, name and description), got %d", len(args))
// 			}

// 			return nil
// 		},
// 		Long: `Edit an existing command.

// What it does:
//   Loads the command with the given ID and updates its name and description.
//   An empty argument is ignored, keeping the current value, so you can change
//   only one field at a time without retyping the other.

// What it expects:
//   Exactly three positional arguments:
//     e <id> <name> <description>   the ID of the command to edit, plus its new
//                                   name and description. Pass "" to keep the
//                                   current value.

//   The name, when changed, must be unique; an already-used name is rejected.

// What it returns:
//   On success it prints "Command edited with id <id>". It exits non-zero when
//   the argument count is wrong, when the ID is not numeric or does not exist,
//   or when the update fails (for example, a duplicate name).

// Examples:
//   tartarus cmds e 1 deploy "Deploy the app to staging"
//   tartarus cmds e 1 "" "Only a new description"`,
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

// 			if args[1] != "" && args[1] != command.Name {
// 				command.Name = args[1]
// 			}

// 			if args[2] != "" && args[2] != command.Description {
// 				command.Description = args[2]
// 			}

// 			if err := r.UpdateCommand(&command); err != nil {
// 				return fmt.Errorf("fail to edit command: %w", err)
// 			}

// 			fmt.Printf("Command edited with id %d\n", command.ID)

// 			return nil
// 		},
// 	}

// 	return rootCmd
// }

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
