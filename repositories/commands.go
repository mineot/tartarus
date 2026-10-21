package repositories

import (
	"errors"
	"fmt"
	"strings"
	"tartarus/store"
	"time"
)

// CommandScript is a single script in a command
type CommandScript struct {
	ID          uint64
	CommandID   uint64
	Name        string
	Script      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Command is a command to be executed
type Command struct {
	ID          uint64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrCommandNotFound             = errors.New("command not found")
	ErrCommandNameRequired         = errors.New("command name is required")
	ErrCommandNameInvalid          = errors.New("command name must contain only lowercase letters and digits")
	ErrCommandNameTooLong          = errors.New("command name must be at most 12 characters")
	ErrCommandIDRequired           = errors.New("command id is required")
	ErrCommandScriptIDRequired     = errors.New("command script id is required")
	ErrCommandScriptNameRequired   = errors.New("command script name is required")
	ErrCommandScriptNameInvalid    = errors.New("command script name must contain only lowercase letters and digits")
	ErrCommandScriptNameTooLong    = errors.New("command script name must be at most 12 characters")
	ErrCommandScriptScriptRequired = errors.New("command script script is required")
	ErrCommandScriptNotFound       = errors.New("command script not found")
)

const maxCommandNameLength = 12

const (
	selectCommands = `
SELECT id, name, COALESCE(description, ''), created_at, updated_at
FROM commands
ORDER BY id
`

	selectCommandsWithFilters = `
SELECT id, name, COALESCE(description, ''), created_at, updated_at
FROM commands
WHERE name LIKE ? ESCAPE '\' OR COALESCE(description, '') LIKE ? ESCAPE '\'
ORDER BY id
`

	selectCommand = `
SELECT id, name, COALESCE(description, ''), created_at, updated_at
FROM commands
WHERE name = ?
`

	selectCommandByID = `
SELECT id, name, COALESCE(description, ''), created_at, updated_at
FROM commands
WHERE id = ?
`

	insertCommand = `
INSERT INTO commands (name, description, created_at, updated_at)
VALUES (?, ?, ?, ?)
`

	updateCommand = `
UPDATE commands
SET name = ?, description = ?, updated_at = ?
WHERE id = ?
`

	deleteCommandScriptsByCommand = `
DELETE FROM command_scripts
WHERE command_id = ?
`

	deleteCommand = `
DELETE FROM commands
WHERE id = ?
`

	selectCommandScripts = `
SELECT id, command_id, name, script, COALESCE(description, ''), created_at, updated_at
FROM command_scripts
WHERE command_id = ?
ORDER BY id
`

	selectCommandScriptsByName = `
SELECT id, command_id, name, script, COALESCE(description, ''), created_at, updated_at
FROM command_scripts
WHERE command_id = ? AND name LIKE ?
ORDER BY id
`

	selectCommandScript = `
SELECT id, command_id, name, script, COALESCE(description, ''), created_at, updated_at
FROM command_scripts
WHERE command_id = ? AND name = ?
`

	selectCommandScriptByID = `
SELECT id, command_id, name, script, COALESCE(description, ''), created_at, updated_at
FROM command_scripts
WHERE command_id = ? AND id = ?
`

	insertCommandScript = `
INSERT INTO command_scripts (command_id, name, script, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
`

	updateCommandScript = `
UPDATE command_scripts
SET name = ?, script = ?, description = ?, updated_at = ?
WHERE id = ? AND command_id = ?
`

	deleteCommandScript = `
DELETE FROM command_scripts
WHERE id = ? AND command_id = ?
`
)

func (c *Command) Exists() bool {
	return c.ID > 0 && c.Name != ""
}

func (c *CommandScript) Exists() bool {
	return c.ID > 0 && c.Name != "" && c.Script != ""
}

// ValidateCommandName rejects empty names, names longer than
// maxCommandNameLength, and names containing anything other than lowercase
// letters and digits.
func ValidateCommandName(name string) error {
	if name == "" {
		return ErrCommandNameRequired
	}

	if len(name) > maxCommandNameLength {
		return ErrCommandNameTooLong
	}

	for i := 0; i < len(name); i++ {
		c := name[i]

		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return ErrCommandNameInvalid
		}
	}

	return nil
}

// ValidateCommandScriptName rejects empty names, names longer than
// maxCommandNameLength, and names containing anything other than lowercase
// letters and digits. It applies the same rule as ValidateCommandName.
func ValidateCommandScriptName(name string) error {
	if name == "" {
		return ErrCommandScriptNameRequired
	}

	if len(name) > maxCommandNameLength {
		return ErrCommandScriptNameTooLong
	}

	for i := 0; i < len(name); i++ {
		c := name[i]

		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return ErrCommandScriptNameInvalid
		}
	}

	return nil
}

// GetCommands returns commands in creation order. Search matches literal
// substrings in name or description using SQLite LIKE (ASCII case-insensitive).
// Empty search matches all commands.
func (r *Repos) GetCommands(searchTerm string) ([]Command, error) {
	var query = selectCommands

	if searchTerm != "" {
		searchTerm = "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(searchTerm) + "%"
		query = selectCommandsWithFilters
	}

	rows, err := r.Store.Query(query, searchTerm, searchTerm)

	if err != nil {
		return nil, fmt.Errorf("repositories: selecting commands: %w", err)
	}

	defer rows.Close()

	var commands []Command

	for rows.Next() {
		var command Command

		if err := rows.Scan(
			&command.ID,
			&command.Name,
			&command.Description,
			&command.CreatedAt,
			&command.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scanning command: %w", err)
		}

		commands = append(commands, command)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterating commands: %w", err)
	}

	return commands, nil
}

// GetCommand returns the command with the given name.
// If no command is found, it returns an empty Command and nil error.
func (r *Repos) GetCommand(name string) (Command, error) {
	rows, err := r.Store.Query(selectCommand, name)

	if err != nil {
		return Command{}, fmt.Errorf("repositories: selecting command: %w", err)
	}

	defer rows.Close()

	var command Command

	if rows.Next() {
		if err := rows.Scan(
			&command.ID,
			&command.Name,
			&command.Description,
			&command.CreatedAt,
			&command.UpdatedAt,
		); err != nil {
			return Command{}, fmt.Errorf("repositories: scanning command: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return Command{}, fmt.Errorf("repositories: iterating command: %w", err)
	}

	return command, nil
}

// GetCommandScripts returns all scripts for the given command.
func (r *Repos) GetCommandScripts(commandID uint64) ([]CommandScript, error) {
	rows, err := r.Store.Query(selectCommandScripts, commandID)

	if err != nil {
		return nil, fmt.Errorf("repositories: selecting command scripts: %w", err)
	}

	defer rows.Close()

	var scripts []CommandScript

	for rows.Next() {
		var script CommandScript

		if err := rows.Scan(
			&script.ID,
			&script.CommandID,
			&script.Name,
			&script.Script,
			&script.Description,
			&script.CreatedAt,
			&script.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scanning command script: %w", err)
		}

		scripts = append(scripts, script)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterating command scripts: %w", err)
	}

	return scripts, nil
}

// GetCommandScriptsByName returns the scripts of a command whose name matches
// searchTerm as a literal substring (SQLite LIKE, ASCII case-insensitive),
// ordered by id. An empty search matches every script of the command.
func (r *Repos) GetCommandScriptsByName(commandID uint64, searchTerm string) ([]CommandScript, error) {
	if commandID == 0 {
		return nil, ErrCommandIDRequired
	}

	var query = selectCommandScripts

	if searchTerm != "" {
		searchTerm = "%" + searchTerm + "%"
		query = selectCommandScriptsByName
	}

	rows, err := r.Store.Query(query, commandID, searchTerm)

	if err != nil {
		return nil, fmt.Errorf("repositories: selecting command scripts: %w", err)
	}

	defer rows.Close()

	var scripts []CommandScript

	for rows.Next() {
		var script CommandScript

		if err := rows.Scan(
			&script.ID,
			&script.CommandID,
			&script.Name,
			&script.Script,
			&script.Description,
			&script.CreatedAt,
			&script.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scanning command script: %w", err)
		}

		scripts = append(scripts, script)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterating command scripts: %w", err)
	}

	return scripts, nil
}

// GetCommandScript returns a specific script for a command.
// If not found, returns empty CommandScript and nil error.
func (r *Repos) GetCommandScript(commandID uint64, scriptName string) (CommandScript, error) {
	rows, err := r.Store.Query(selectCommandScript, commandID, scriptName)

	if err != nil {
		return CommandScript{}, fmt.Errorf("repositories: selecting command script: %w", err)
	}

	defer rows.Close()

	var script CommandScript

	if rows.Next() {
		if err := rows.Scan(
			&script.ID,
			&script.CommandID,
			&script.Name,
			&script.Script,
			&script.Description,
			&script.CreatedAt,
			&script.UpdatedAt,
		); err != nil {
			return CommandScript{}, fmt.Errorf("repositories: scanning command script: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return CommandScript{}, fmt.Errorf("repositories: iterating command script: %w", err)
	}

	return script, nil
}

// InsertCommand validates and persists a new command.
// On success, c is populated with inserted values.
func (r *Repos) InsertCommand(c *Command) error {
	if err := ValidateCommandName(c.Name); err != nil {
		return err
	}

	now := time.Now().UTC()

	var id int64

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(insertCommand, c.Name, c.Description, now, now)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()

		if err != nil {
			return err
		}

		var inserted Command

		row := tx.QueryRow(selectCommandByID, uint64(id))

		if err := row.Scan(
			&inserted.ID,
			&inserted.Name,
			&inserted.Description,
			&inserted.CreatedAt,
			&inserted.UpdatedAt,
		); err != nil {
			return err
		}

		c.ID = inserted.ID
		c.Name = inserted.Name
		c.Description = inserted.Description
		c.CreatedAt = inserted.CreatedAt
		c.UpdatedAt = inserted.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: inserting command: %w", err)
	}

	return nil
}

// AppendCommandScript validates and appends an script to a command.
func (r *Repos) AppendCommandScript(script *CommandScript) error {
	if script.CommandID == 0 {
		return ErrCommandIDRequired
	}

	if err := ValidateCommandScriptName(script.Name); err != nil {
		return err
	}

	if script.Script == "" {
		return ErrCommandScriptScriptRequired
	}

	now := time.Now().UTC()

	var id int64

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(
			insertCommandScript,
			script.CommandID,
			script.Name,
			script.Script,
			script.Description,
			now,
			now,
		)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()

		if err != nil {
			return err
		}

		var inserted CommandScript

		row := tx.QueryRow(selectCommandScriptByID, script.CommandID, uint64(id))

		if err := row.Scan(
			&inserted.ID,
			&inserted.CommandID,
			&inserted.Name,
			&inserted.Script,
			&inserted.Description,
			&inserted.CreatedAt,
			&inserted.UpdatedAt,
		); err != nil {
			return err
		}

		script.ID = inserted.ID
		script.CommandID = inserted.CommandID
		script.Name = inserted.Name
		script.Script = inserted.Script
		script.Description = inserted.Description
		script.CreatedAt = inserted.CreatedAt
		script.UpdatedAt = inserted.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: appending command script: %w", err)
	}

	return nil
}

// UpdateCommand validates and updates an existing command.
func (r *Repos) UpdateCommand(c *Command) error {
	if c.ID == 0 {
		return ErrCommandIDRequired
	}

	if err := ValidateCommandName(c.Name); err != nil {
		return err
	}

	now := time.Now().UTC()

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(updateCommand, c.Name, c.Description, now, c.ID)

		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()

		if err != nil {
			return err
		}

		if affected == 0 {
			return ErrCommandNotFound
		}

		var updated Command

		row := tx.QueryRow(selectCommandByID, c.ID)

		if err := row.Scan(
			&updated.ID,
			&updated.Name,
			&updated.Description,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		); err != nil {
			return err
		}

		c.ID = updated.ID
		c.Name = updated.Name
		c.Description = updated.Description
		c.CreatedAt = updated.CreatedAt
		c.UpdatedAt = updated.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: updating command: %w", err)
	}

	return nil
}

// UpdateCommandScript validates and updates a command script.
func (r *Repos) UpdateCommandScript(script *CommandScript) error {
	if script.CommandID == 0 {
		return ErrCommandIDRequired
	}

	if script.ID == 0 {
		return ErrCommandScriptIDRequired
	}

	if err := ValidateCommandScriptName(script.Name); err != nil {
		return err
	}

	if script.Script == "" {
		return ErrCommandScriptScriptRequired
	}

	now := time.Now().UTC()

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(updateCommandScript, script.Name, script.Script, script.Description, now, script.ID, script.CommandID)

		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()

		if err != nil {
			return err
		}

		if affected == 0 {
			return ErrCommandScriptNotFound
		}

		var updated CommandScript

		row := tx.QueryRow(selectCommandScriptByID, script.CommandID, script.ID)

		if err := row.Scan(
			&updated.ID,
			&updated.CommandID,
			&updated.Name,
			&updated.Script,
			&updated.Description,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		); err != nil {
			return err
		}

		script.ID = updated.ID
		script.CommandID = updated.CommandID
		script.Name = updated.Name
		script.Script = updated.Script
		script.Description = updated.Description
		script.CreatedAt = updated.CreatedAt
		script.UpdatedAt = updated.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: updating command script: %w", err)
	}

	return nil
}

// DeleteCommand removes a command and its scripts.
// If the command does not exist, it returns nil (idempotent).
func (r *Repos) DeleteCommand(id uint64) error {
	if id == 0 {
		return ErrCommandIDRequired
	}

	err := r.Store.WithTx(func(tx *store.Tx) error {
		if _, err := tx.Exec(deleteCommandScriptsByCommand, id); err != nil {
			return err
		}

		if _, err := tx.Exec(deleteCommand, id); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: deleting command: %w", err)
	}

	return nil
}

// RemoveCommandScript removes an script from a command.
// If the script does not exist, it returns nil (idempotent).
func (r *Repos) RemoveCommandScript(commandID, scriptID uint64) error {
	if commandID == 0 {
		return ErrCommandIDRequired
	}

	if scriptID == 0 {
		return ErrCommandScriptIDRequired
	}

	err := r.Store.WithTx(func(tx *store.Tx) error {
		_, err := tx.Exec(deleteCommandScript, scriptID, commandID)
		return err
	})

	if err != nil {
		return fmt.Errorf("repositories: removing command script: %w", err)
	}

	return nil
}
