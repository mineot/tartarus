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
		appendCommandScript(r),
	)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ls",
		Short: textListAllCmds,
		Args:  cobra.NoArgs,
		Long:  longDescListAllCmds,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validateFlags(cmd, "filter", filterRaw, "the --filter value cannot be empty")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			commands, err := r.GetCommands(filterRaw)

			if err != nil {
				return failTo("get commands", err)
			}

			var rows [][]string

			for _, command := range commands {
				rows = append(rows, []string{
					command.Name,
					command.Description,
				})
			}

			table := components.Table{
				Title:    textListAllCmds,
				EmptyMsg: msgNoCmdsFound,
				Header:   tbHrCmds,
				Rows:     rows,
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
		Short: textCmd,
		Long:  longDescShowCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "name", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := r.GetCommand(args[0])

			if err != nil {
				return failTo("get command", err)
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			cmdTable := components.Table{
				Title:  textCmd,
				Header: tbHrCmds,
				Rows:   [][]string{{command.Name, command.Description}},
			}

			cmdTable.Render()

			fmt.Println()

			scripts, err := r.GetCommandScripts(command.ID)

			if err != nil {
				return failTo("get command scripts", err)
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
				Title:    textCmdSrcs,
				EmptyMsg: msgNoCmdScrsFound,
				Header:   tbHrCmdScrs,
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
		Short: textCreateCmd,
		Long:  longDescCreateCmd,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			components.PromptTitle(textCreateCmd)

			command.Name, err = components.Prompt(components.PromptOpts{
				Label:   textName,
				Default: "",
				Validate: func(name string) error {
					return repositories.ValidateCommandName(name)
				},
			})

			if err != nil {
				return err
			}

			command.Description, err = components.Prompt(components.PromptOpts{
				Label:   textDescription,
				Default: "",
				Validate: func(description string) error {
					if description == "" {
						return errors.New(reqDescription)
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			err = r.InsertCommand(&command)

			if err != nil {
				return failTo("insert command", err)
			}

			successMsg(strings.Join([]string{msgSuccessCreatedCmd, command.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

func editCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "e <name>",
		Short: textEditCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "name", args)
		},
		Long: longDescEditCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			components.PromptTitle(textEditCmd)

			command, err = r.GetCommand(args[0])

			if err != nil {
				return err
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			command.Name, err = components.Prompt(components.PromptOpts{
				Label:   textName,
				Default: command.Name,
				Validate: func(name string) error {
					return repositories.ValidateCommandName(name)
				},
			})

			if err != nil {
				return err
			}

			command.Description, err = components.Prompt(components.PromptOpts{
				Label:   textDescription,
				Default: command.Description,
				Validate: func(description string) error {
					if description == "" {
						return errors.New(reqDescription)
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			err = r.UpdateCommand(&command)

			if err != nil {
				return failTo("update command", err)
			}

			successMsg(strings.Join([]string{msgSuccessEditedCmd, command.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

func deleteCommand(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "d <name>",
		Short: textDeleteCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "name", args)
		},
		Long: longDescDeleteCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			var command = repositories.Command{}
			var err error

			components.PromptTitle(textDeleteCmd)

			command, err = r.GetCommand(args[0])

			if err != nil {
				return err
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			question, err := components.Confirm(components.ConfirmOpts{
				Question:   deleteMsg(fmt.Sprintf("Do you want to delete this command '%s'?", command.Name)),
				YesDefault: false,
			})

			if err != nil {
				return err
			}

			if !question {
				fmt.Println()
				fmt.Println(msgNoDeletedCmd)
				return nil
			}

			err = r.DeleteCommand(command.ID)

			if err != nil {
				return failTo("delete command", err)
			}

			successMsg(strings.Join([]string{msgSuccessDeletedCmd, command.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

func appendCommandScript(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ap <commandName>",
		Short: textAppSrcCmd,
		Long:  londDescAppendScriptCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "commandName", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var command repositories.Command
			var err error

			if command, err = r.GetCommand(args[0]); err != nil {
				return err
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			components.PromptTitle(textAppSrcCmd)

			var script = repositories.CommandScript{CommandID: command.ID}

			script.Name, err = components.Prompt(components.PromptOpts{
				Label: textName,
				Validate: func(name string) error {
					return repositories.ValidateCommandScriptName(name)
				},
			})

			if err != nil {
				return err
			}

			script.Script, err = components.Prompt(components.PromptOpts{
				Label: textScript,
				Validate: func(script string) error {
					if script == "" {
						// TODO: validar o comando, se ele é um comando valido e funcional
						return errors.New(reqScript)
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			script.Description, err = components.Prompt(components.PromptOpts{
				Label: textDescription,
				Validate: func(description string) error {
					if description == "" {
						return errors.New(reqDescription)
					}

					return nil
				},
			})

			if err != nil {
				return err
			}

			err = r.AppendCommandScript(&script)

			if err != nil {
				return failTo("append command script", err)
			}

			successMsg(strings.Join([]string{"Script appended to command", script.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

// // func changeCommandItem(r *repositories.Repos) *cobra.Command {
// // 	rootCmd := &cobra.Command{
// // 		Use:   "ch <commandId> <commandItemId> <name> <script> <description>",
// // 		Short: "Change an existing item of an existing command",
// // 		Args:  cobra.NoArgs,
// // 		Long:  ``,
// // 		PreRunE: func(cmd *cobra.Command, args []string) error {
// // 			return nil
// // 		},
// // 		RunE: func(cmd *cobra.Command, args []string) error {
// // 			return nil
// // 		},
// // 	}

// // 	return rootCmd
// // }

// // func removeCommandItem(r *repositories.Repos) *cobra.Command {
// // 	rootCmd := &cobra.Command{
// // 		Use:   "rm <commandItemId>",
// // 		Short: "Remove an existing item of an existing command",
// // 		Args:  cobra.NoArgs,
// // 		Long:  ``,
// // 		PreRunE: func(cmd *cobra.Command, args []string) error {
// // 			return nil
// // 		},
// // 		RunE: func(cmd *cobra.Command, args []string) error {
// // 			return nil
// // 		},
// // 	}

// // 	return rootCmd
// // }
