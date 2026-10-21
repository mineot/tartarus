package views

import (
	"errors"
	"fmt"
	"strings"
	"tartarus/components"
	"tartarus/repositories"

	"github.com/spf13/cobra"
)

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
		changeCommandScript(r),
		removeCommandScript(r),
	)

	return rootCmd
}

func listAllCommands(r *repositories.Repos) *cobra.Command {
	var filterRaw string
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
		Short: textViewCmd,
		Long:  longDescShowCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "name", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := r.GetCommand(args[0])

			if err != nil {
				components.PromptTitle(textViewCmd)
				return failTo("get command", err)
			}

			if !command.Exists() {
				components.PromptTitle(textViewCmd)
				fmt.Println(msgNoCmdFound)
				return nil
			}

			cmdTable := components.Table{
				Title:  textViewCmd,
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
			components.PromptTitle(textCreateCmd)

			var command = repositories.Command{}
			var err error

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
			components.PromptTitle(textEditCmd)

			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return failTo("get command", err)
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
			components.PromptTitle(textDeleteCmd)

			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return failTo("get command", err)
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			question, err := components.Confirm(components.ConfirmOpts{
				Question:   deleteMsg(fmt.Sprintf("Do you want to delete command '%s' and all its scripts?", command.Name)),
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
		Long:  longDescAppendScriptCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(1, "commandName", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			components.PromptTitle(textAppSrcCmd)

			var command repositories.Command
			var err error

			if command, err = r.GetCommand(args[0]); err != nil {
				return failTo("get command", err)
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

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

			successMsg(strings.Join([]string{msgSuccessAppendedCmd, command.Name + "/" + script.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

func changeCommandScript(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ch <commandName> <scriptName>",
		Short: textChangeSrcCmd,
		Long:  longDescChangeScriptCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(2, "commandName | scriptName", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			components.PromptTitle(textChangeSrcCmd)

			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return failTo("get command", err)
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			var script = repositories.CommandScript{}

			script, err = r.GetCommandScript(command.ID, args[1])

			if err != nil {
				return failTo("get command script", err)
			}

			if !script.Exists() {
				fmt.Println(msgNoCmdSrcFound)
				return nil
			}

			script.Name, err = components.Prompt(components.PromptOpts{
				Label:   textName,
				Default: script.Name,
				Validate: func(name string) error {
					return repositories.ValidateCommandScriptName(name)
				},
			})

			if err != nil {
				return err
			}

			script.Script, err = components.Prompt(components.PromptOpts{
				Label:   textScript,
				Default: script.Script,
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
				Label:   textDescription,
				Default: script.Description,
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

			err = r.UpdateCommandScript(&script)

			if err != nil {
				return failTo("update command script", err)
			}

			successMsg(strings.Join([]string{msgSuccessChangedCmd, command.Name + "/" + script.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}

func removeCommandScript(r *repositories.Repos) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rm <commandName> <scriptName>",
		Short: textRemSrcCmd,
		Long:  longDescRemoveScriptCmd,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(2, "commandName | scriptName", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			components.PromptTitle(textRemSrcCmd)

			var command = repositories.Command{}
			var err error

			command, err = r.GetCommand(args[0])

			if err != nil {
				return failTo("get command", err)
			}

			if !command.Exists() {
				fmt.Println(msgNoCmdFound)
				return nil
			}

			var script = repositories.CommandScript{}

			script, err = r.GetCommandScript(command.ID, args[1])

			if err != nil {
				return failTo("get command script", err)
			}

			if !script.Exists() {
				fmt.Println(msgNoCmdSrcFound)
				return nil
			}

			confirmed, err := components.Confirm(components.ConfirmOpts{
				Question:   deleteMsg(fmt.Sprintf("Do you want to remove script '%s' from command '%s'?", script.Name, command.Name)),
				YesDefault: false,
			})
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Printf("\n%s\n", msgNoRemovedCmd)
				return nil
			}

			err = r.RemoveCommandScript(script.CommandID, script.ID)

			if err != nil {
				return failTo("remove command script", err)
			}

			successMsg(strings.Join([]string{msgSuccessRemoveCmd, command.Name + "/" + script.Name}, ": "))

			return nil
		},
	}

	return rootCmd
}
