package backup

import (
	"encoding/json"
	"fmt"
	"os"

	"tartarus/repositories"
)

// Import replaces the contents of the database with the contents of the backup
// file at path.
//
// It destroys whatever was there: the schema is dropped and rebuilt by
// ResetMigrations before the first insert, so an import is always a replacement
// and never a merge. Importing on top of existing rows would hit the UNIQUE
// constraints on commands.name, command_scripts.name, command_scripts.script and
// manuals.name anyway.
//
// The reset is also what makes a failed import recoverable. The inserts are not
// atomic — every repositories write opens its own transaction — so a bad row
// halfway through leaves a partial database behind. Re-running is still safe,
// because the reset at the top clears that partial state before the UNIQUE
// constraints can collide with it. A transaction spanning the whole import would
// be the better answer, and it is not available: there is no InsertCommandTx.
//
// Ids and timestamps in the file are not preserved. The repositories stamp
// time.Now() on every write and let SQLite assign ids from AUTOINCREMENT, so the
// id in the file is ignored and the imported rows carry import time.
//
// It must not be called from inside a store.WithTx callback: ResetMigrations
// opens its own transaction, which returns store.ErrTxActive. That is a
// different sentinel from the store.ErrUseTx Export runs into, for the same
// constraint.
func Import(r *repositories.Repos, path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("backup: reading %s: %w", path, err)
	}

	var file jsonFile

	if err = json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("backup: unmarshalling %s: %w", path, err)
	}

	// Validate before the reset, not after. Every check here is one the database
	// would reject on insert, and running them first is the difference between a
	// rejected file and a rejected file that also destroyed the previous data.
	if err = validate(file); err != nil {
		return fmt.Errorf("backup: validating: %w", err)
	}

	if err = r.Store.ResetMigrations(); err != nil {
		return fmt.Errorf("backup: resetting the database: %w", err)
	}

	for _, cmd := range file.Commands {
		c := &repositories.Command{
			Name:        cmd.Name,
			Description: cmd.Description,
		}

		if err = r.InsertCommand(c); err != nil {
			return fmt.Errorf("backup: inserting command %q: %w", cmd.Name, err)
		}

		for _, script := range cmd.Scripts {
			// c.ID is the id the row just got, not cmd.ID from the file: after a
			// reset the two can point at entirely different commands, and
			// AppendCommandScript overwrites the script's CommandID with the row it
			// reads back regardless.
			err = r.AppendCommandScript(c.ID, &repositories.CommandScript{
				Name:        script.Name,
				Script:      script.Script,
				Description: script.Description,
			})

			if err != nil {
				return fmt.Errorf("backup: appending script %q of command %q: %w", script.Script, cmd.Name, err)
			}
		}
	}

	for _, man := range file.Manuals {
		m := &repositories.Manual{
			Name:        man.Name,
			Body:        man.Body,
			Description: man.Description,
		}

		if err = r.InsertManual(m); err != nil {
			return fmt.Errorf("backup: inserting manual %q: %w", man.Name, err)
		}
	}

	return nil
}

// validate rejects a decoded backup file that the repositories or the schema
// would refuse anyway, so those refusals happen before the reset instead of
// after it.
//
// It reports the first problem it finds rather than collecting them all: the
// point is to avoid destroying the database, and one clear error is enough to
// achieve that.
//
// The duplicate checks exist because the constraints are global rather than
// scoped to a command. commands.name and manuals.name are unique on their own,
// and command_scripts.name and command_scripts.script are unique across the
// whole table, so two different commands cannot hold the same script name or the
// same script — which looks like a schema mistake rather than a rule, and is
// worth reporting as one.
func validate(file jsonFile) error {
	names := make(map[string]struct{}, len(file.Commands))
	scriptNames := make(map[string]struct{})
	scripts := make(map[string]struct{})

	for i, cmd := range file.Commands {
		if cmd.Name == "" {
			return fmt.Errorf("command at index %d has no name", i)
		}

		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("command name %q appears more than once", cmd.Name)
		}

		names[cmd.Name] = struct{}{}

		for _, script := range cmd.Scripts {
			if err := repositories.ValidateCommandScriptName(script.Name); err != nil {
				return fmt.Errorf("command %q has an invalid script name %q: %w", cmd.Name, script.Name, err)
			}

			if _, ok := scriptNames[script.Name]; ok {
				return fmt.Errorf("script name %q appears more than once, and it is unique across all commands", script.Name)
			}

			scriptNames[script.Name] = struct{}{}

			if script.Script == "" {
				return fmt.Errorf("command %q has a script with no script", cmd.Name)
			}

			if _, ok := scripts[script.Script]; ok {
				return fmt.Errorf("script %q appears more than once, and it is unique across all commands", script.Script)
			}

			scripts[script.Script] = struct{}{}
		}
	}

	manualNames := make(map[string]struct{}, len(file.Manuals))

	for i, man := range file.Manuals {
		if man.Name == "" {
			return fmt.Errorf("manual at index %d has no name", i)
		}

		if man.Body == "" {
			return fmt.Errorf("manual %q has no body", man.Name)
		}

		if _, ok := manualNames[man.Name]; ok {
			return fmt.Errorf("manual name %q appears more than once", man.Name)
		}

		manualNames[man.Name] = struct{}{}
	}

	return nil
}
