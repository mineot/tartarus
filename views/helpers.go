package views

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	msgNoCmdFound         = "No command found!"
	msgNoCmdScrsFound     = "No command scripts found!"
	msgNoCmdsFound        = "No commands found!"
	msgNoCmdSrcFound      = "No command script found!"
	msgNoRemovedCmd       = "The command script was not removed!"
	msgNoDeletedCmd       = "The command was not deleted!"
	msgSuccessAppendedCmd = "Command script successfully appended"
	msgSuccessChangedCmd  = "Command script successfully changed"
	msgSuccessCreatedCmd  = "Command successfully created"
	msgSuccessDeletedCmd  = "Command successfully deleted"
	msgSuccessEditedCmd   = "Command successfully edited"
	msgSuccessRemoveCmd   = "Command script successfully removed"

	textAppSrcCmd    = "Append script to a command"
	textChangeSrcCmd = "Change script of a command"
	textCmdSrcs      = "Command scripts"
	textCreateCmd    = "Create new command"
	textDeleteCmd    = "Delete command"
	textDescription  = "Description"
	textEditCmd      = "Edit command"
	textListAllCmds  = "List all commands"
	textName         = "Name"
	textScript       = "Script"
	textViewCmd      = "View command"
	textRemSrcCmd    = "Remove script from a command"

	reqDescription = "Description is required"
	reqScript      = "Script is required"
)

var (
	tbHrCmds    = []string{"Name", "Description"}
	tbHrCmdScrs = []string{"Name", "Script", "Description"}
)

func failTo(description string, err error) error {
	return fmt.Errorf("Fail to %s: %w", description, err)
}

func validateArgs(total int, argName string, args []string) error {
	if len(args) != total {
		return fmt.Errorf("expected exactly %d argument (%s), got %d", total, argName, len(args))
	}

	return nil
}

func validateFlags(cmd *cobra.Command, filter string, variable string, msg string) error {
	if cmd.Flags().Changed(filter) && strings.TrimSpace(variable) == "" {
		return errors.New(msg)
	}

	return nil
}

func successMsg(msg string) {
	fmt.Printf("\n%s\n", msg)
}

func deleteMsg(msg string) string {
	s := strings.Builder{}
	s.WriteString("Are you sure?")
	s.WriteString("\n")
	s.WriteString(msg)
	return s.String()
}

const (
	longDescListAllCmds = `List every command stored in the database as a table with its name and
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
  tartarus cmds ls -f "build"`

	longDescShowCmd = `Show a single command and its scripts.

What it does:
  Reads the command with the given name and prints its details (name and
  description), followed by a second table listing every script that belongs
  to it (name, script and description).

What it expects:
  Exactly one positional argument:
    v <name>   the name of the command to display

What it returns:
  A "View command" table with the columns Name and Description, then a "Command
  scripts" table with the columns Name, Script and Description. If the
  command has no scripts, it prints "No command scripts found" for the
  second table. It exits non-zero when the argument count is wrong or when
  either database query fails. If no command matches, it prints
  "No command found!" and exits zero.

Examples:
  tartarus cmds v deploy`

	longDescCreateCmd = `Create a new command.

What it does:
  Asks for the name and the description of the new command, then inserts it
  into the database.

What it expects:
  No positional arguments. Two interactive prompts, in order:
    Name          required; 1-12 ASCII lowercase letters or digits, unique.
    Description   required; a non-empty text.

What it returns:
  On success it prints "Command successfully created: <name>". It exits
  non-zero when a prompt fails (interrupt or read/write error) or when the
  insert fails (for example, a duplicate name).

Examples:
  tartarus cmds c`

	longDescEditCmd = `Edit an existing command.

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
  On success it prints "Command successfully edited: <name>". It exits non-zero
  when the argument count is wrong, when a prompt fails, or when a database
  operation fails (for example, a duplicate name). If no command matches,
  it prints "No command found!" and exits zero.

Examples:
  tartarus cmds e deploy`

	longDescDeleteCmd = `Delete an existing command.

What it does:
  Loads the command with the given name and asks for confirmation. When
  confirmed, it removes the command together with all of its scripts.

What it expects:
  Exactly one positional argument:
    d <name>   the name of the command to delete

  The confirmation is answered with y or n and defaults to n; an invalid
  answer is asked again.

What it returns:
  When confirmed it prints "Command successfully deleted: <name>". When
  declined it prints "The command was not deleted!" and exits zero. It exits
  non-zero when the argument count is wrong, when a prompt fails, or when a
  database operation fails. If no command matches, it prints
  "No command found!" and exits zero.

Examples:
  tartarus cmds d deploy`

	longDescAppendScriptCmd = `Append a script to an existing command.

What it does:
  Loads the command and asks for the script name, script text and description.

What it expects:
  ap <commandName>
  Name is required: 1-12 ASCII lowercase letters or digits.
  Script and Description are required. Name and Script must each be unique
  across all commands.

What it returns:
  Prints "Command script successfully appended: <commandName>/<scriptName>".
  A missing command prints "No command found!" and exits zero.
  Invalid argument counts, prompt failures and database errors exit non-zero.

Examples:
  tartarus cmds ap deploy`

	longDescChangeScriptCmd = `Change a script of an existing command.

What it does:
  Loads the command and script, then asks for Name, Script and Description.
  Empty answers keep the current values. All three fields are required.
  Name must contain 1-12 ASCII lowercase letters or digits.
  Name and Script must each be unique across all commands.

What it expects:
  ch <commandName> <scriptName>

What it returns:
  Prints "Command script successfully changed: <commandName>/<scriptName>"
  using the updated script name. A missing command or script prints
  "No command found!" or "No command script found!" and exits zero.
  Invalid argument counts, prompt failures and database errors exit non-zero.

Examples:
  tartarus cmds ch deploy build`

	longDescRemoveScriptCmd = `Remove a script from an existing command.

What it does:
  Loads the command and script and asks for confirmation before removal.
  Accepts y/yes or n/no, case-insensitively; an empty answer cancels.
  Invalid answers are asked again.

What it expects:
  rm <commandName> <scriptName>

What it returns:
  Prints "Command script successfully removed: <commandName>/<scriptName>"
  when confirmed, or "The command script was not removed!" when cancelled.
  A missing command or script prints "No command found!" or
  "No command script found!" and exits zero, as does cancellation.
  Invalid argument counts, prompt failures and database errors exit non-zero.

Examples:
  tartarus cmds rm deploy build`
)
