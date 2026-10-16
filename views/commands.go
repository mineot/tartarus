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
		deleteCommand(r),
	)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all commands",
		Args:  cobra.NoArgs,
		Long: `List every command stored in the database as a table with its name and
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
  A table with the columns Name and Description. If nothing matches, it
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
		Long: `Show a single command and its scripts.

What it does:
  Reads the command with the given name and prints its details (name and
  description), followed by a second table listing every script that belongs
  to it (name, script and description).

What it expects:
  Exactly one positional argument:
    v <name>   the name of the command to display

What it returns:
  A "Command" table with the rows Name and Description, then a "Command
  scripts" table with the columns Name, Script and Description. If the
  command has no scripts, it prints "No command scripts found" for the
  second table. It exits non-zero when the argument count is wrong or when
  either database query fails. A name that matches no command is shown with
  empty details and an empty script table.

Examples:
  tartarus cmds v deploy`,
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

			if command.Name == "" {
				fmt.Println("No command found")
				return nil
			}

			cmdTable := components.Table{
				Title:    "Command",
				EmptyMsg: "No command found",
				Header:   []string{"Name", "Description"},
				Rows: [][]string{
					{
						command.Name,
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
					script.Name,
					script.Script,
					script.Description,
				})
			}

			cmdScripts := components.Table{
				Title:    "Command scripts",
				EmptyMsg: "No command scripts found",
				Header:   []string{"Name", "Script", "Description"},
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
  Asks for the name and the description of the new command, then inserts it
  into the database.

What it expects:
  No positional arguments. Two interactive prompts, in order:
    Name          required; 1-12 ASCII lowercase letters or digits, unique.
    Description   required; a non-empty text.

What it returns:
  On success it prints "Command successfully created <name>". It exits
  non-zero when a prompt fails (interrupt or read/write error) or when the
  insert fails (for example, a duplicate name).

Examples:
  tartarus cmds c`,
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
		Short: "Edit a command by name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly 1 argument (name), got %d", len(args))
			}

			return nil
		},
		Long: `Edit an existing command.

What it does:
  Loads the command with the given name and asks for its new name and
  description, offering the current values as defaults. An empty answer keeps
  the current value, so you can change one field at a time without retyping
  the other.

What it expects:
  Exactly one positional argument:
    e <name>   the name of the command to edit

  A new name follows the same rule as creation (1-12 ASCII lowercase letters
  or digits) and must be unique; an already-used name is rejected.

What it returns:
  On success it prints "Command successfully edited <name>". It exits non-zero
  when the argument count is wrong, when no command has the given name (the
  loaded command has no ID, so the update fails), when a prompt fails, or
  when the update fails (for example, a duplicate name).

Examples:
  tartarus cmds e deploy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return err
			}

			if command.ID == 0 && command.Name == "" {
				fmt.Println("Command not found")
				return nil
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

func deleteCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "d <name>",
		Short: "Delete a command by name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 1 {
				return fmt.Errorf("expected 1 argument (name), got %d", len(args))
			}

			return nil
		},
		Long: `Delete an existing command.

What it does:
  Loads the command with the given name and asks for confirmation. When
  confirmed, it removes the command together with all of its scripts.

What it expects:
  Exactly one positional argument:
    d <name>   the name of the command to delete

  The confirmation is answered with y or n and defaults to n; an invalid
  answer is asked again.

What it returns:
  When confirmed it prints "Command successfully deleted <name>". When
  declined it prints "The command was not deleted!" and exits zero. It exits
  non-zero when the argument count is wrong, when no command has the given
  name (the delete fails because the loaded command has no ID), or when the
  delete fails.

Examples:
  tartarus cmds d deploy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return err
			}

			if command.ID == 0 && command.Name == "" {
				fmt.Println("Command not found")
				return nil
			}

			question, err := components.Confirm(components.ConfirmOpts{
				Question:   fmt.Sprintf("Are you sure you want to delete this command '%s'?", command.Name),
				YesDefault: false,
			})

			if err != nil {
				return err
			}

			if !question {
				fmt.Print("\nThe command was not deleted!")
				return nil
			}

			err = r.DeleteCommand(command.ID)

			if err != nil {
				return fmt.Errorf("fail to delete command: %w", err)
			}

			fmt.Printf("\nCommand successfully deleted %s\n", command.Name)

			return nil
		},
	}

	return rootCmd
}

// func appendCommandItem(r *repositories.Repos) *cobra.Command {
// 	rootCmd := &cobra.Command{
// 		Use:   "ap <commandName>",
// 		Short: "Append a new script to a command",
// 		Args:  cobra.NoArgs,
// 		Long:  ``, // TODO: criar esse conteúdo, precisa representar o que esse comando faz
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			if len(args) < 1 || len(args) > 1 {
// 				return fmt.Errorf("expected 1 argument (commandName), got %d", len(args))
// 			}

// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			var command repositories.Command
// 			var err error

// 			if command, err = r.GetCommand(args[0]); err != nil {
// 				return err
// 			}

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
