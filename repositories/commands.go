package repositories

import (
	"errors"
	"fmt"
	"tartarus/store"
	"time"
)

type CommandItem struct {
	ID        uint64
	CommandID uint64
	Script    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Command struct {
	ID        uint64
	Name      string
	Items     []CommandItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrCommandNotFound           = errors.New("command not found")
	ErrCommandNameRequired       = errors.New("command name is required")
	ErrCommandIDRequired         = errors.New("command id is required")
	ErrCommandItemIDRequired     = errors.New("command item id is required")
	ErrCommandItemScriptRequired = errors.New("command item script is required")
	ErrCommandItemNotFound       = errors.New("command item not found")
)

const (
	selectCommands = `
SELECT id, name, created_at, updated_at
FROM commands
ORDER BY id
`

	selectCommand = `
SELECT id, name, created_at, updated_at
FROM commands
WHERE id = ?
`

	insertCommand = `
INSERT INTO commands (name, created_at, updated_at)
VALUES (?, ?, ?)
`

	updateCommand = `
UPDATE commands
SET name = ?, updated_at = ?
WHERE id = ?
`

	deleteCommandItemsByCommand = `
DELETE FROM command_items
WHERE command_id = ?
`

	deleteCommand = `
DELETE FROM commands
WHERE id = ?
`

	selectCommandItems = `
SELECT id, command_id, script, created_at, updated_at
FROM command_items
WHERE command_id = ?
ORDER BY id
`

	selectCommandItem = `
SELECT id, command_id, script, created_at, updated_at
FROM command_items
WHERE command_id = ? AND id = ?
`

	insertCommandItem = `
INSERT INTO command_items (command_id, script, created_at, updated_at)
VALUES (?, ?, ?, ?)
`

	updateCommandItem = `
UPDATE command_items
SET script = ?, updated_at = ?
WHERE id = ? AND command_id = ?
`

	deleteCommandItem = `
DELETE FROM command_items
WHERE id = ? AND command_id = ?
`
)

// GetCommands returns every command, in creation order.
func (r *Repos) GetCommands() ([]Command, error) {
	rows, err := r.Store.Query(selectCommands)

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

// GetCommand returns the command with the given id.
// If no command is found, it returns an empty Command and nil error.
func (r *Repos) GetCommand(id uint64) (Command, error) {
	rows, err := r.Store.Query(selectCommand, id)

	if err != nil {
		return Command{}, fmt.Errorf("repositories: selecting command: %w", err)
	}

	defer rows.Close()

	var command Command

	if rows.Next() {
		if err := rows.Scan(
			&command.ID,
			&command.Name,
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

// GetCommandItems returns all items for the given command.
func (r *Repos) GetCommandItems(commandID uint64) ([]CommandItem, error) {
	rows, err := r.Store.Query(selectCommandItems, commandID)

	if err != nil {
		return nil, fmt.Errorf("repositories: selecting command items: %w", err)
	}

	defer rows.Close()

	var items []CommandItem

	for rows.Next() {
		var item CommandItem

		if err := rows.Scan(
			&item.ID,
			&item.CommandID,
			&item.Script,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scanning command item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterating command items: %w", err)
	}

	return items, nil
}

// GetCommandItem returns a specific item for a command.
// If not found, returns empty CommandItem and nil error.
func (r *Repos) GetCommandItem(commandID, itemID uint64) (CommandItem, error) {
	rows, err := r.Store.Query(selectCommandItem, commandID, itemID)

	if err != nil {
		return CommandItem{}, fmt.Errorf("repositories: selecting command item: %w", err)
	}

	defer rows.Close()

	var item CommandItem

	if rows.Next() {
		if err := rows.Scan(
			&item.ID,
			&item.CommandID,
			&item.Script,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return CommandItem{}, fmt.Errorf("repositories: scanning command item: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return CommandItem{}, fmt.Errorf("repositories: iterating command item: %w", err)
	}

	return item, nil
}

// InsertCommand validates and persists a new command.
// On success, c is populated with inserted values.
func (r *Repos) InsertCommand(c *Command) error {
	if c.Name == "" {
		return ErrCommandNameRequired
	}

	now := time.Now().UTC()

	var id int64

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(insertCommand, c.Name, now, now)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()

		if err != nil {
			return err
		}

		var inserted Command

		row := tx.QueryRow(selectCommand, uint64(id))

		if err := row.Scan(
			&inserted.ID,
			&inserted.Name,
			&inserted.CreatedAt,
			&inserted.UpdatedAt,
		); err != nil {
			return err
		}

		c.ID = inserted.ID
		c.Name = inserted.Name
		c.CreatedAt = inserted.CreatedAt
		c.UpdatedAt = inserted.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: inserting command: %w", err)
	}

	return nil
}

// AppendCommandItem validates and appends an item to a command.
func (r *Repos) AppendCommandItem(commandID uint64, item *CommandItem) error {
	if commandID == 0 {
		return ErrCommandIDRequired
	}

	if item.Script == "" {
		return ErrCommandItemScriptRequired
	}

	now := time.Now().UTC()

	var id int64

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(insertCommandItem, commandID, item.Script, now, now)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()

		if err != nil {
			return err
		}

		var inserted CommandItem

		row := tx.QueryRow(selectCommandItem, commandID, uint64(id))

		if err := row.Scan(
			&inserted.ID,
			&inserted.CommandID,
			&inserted.Script,
			&inserted.CreatedAt,
			&inserted.UpdatedAt,
		); err != nil {
			return err
		}

		item.ID = inserted.ID
		item.CommandID = inserted.CommandID
		item.Script = inserted.Script
		item.CreatedAt = inserted.CreatedAt
		item.UpdatedAt = inserted.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: appending command item: %w", err)
	}

	return nil
}

// UpdateCommand validates and updates an existing command.
func (r *Repos) UpdateCommand(c *Command) error {
	if c.ID == 0 {
		return ErrCommandIDRequired
	}

	if c.Name == "" {
		return ErrCommandNameRequired
	}

	now := time.Now().UTC()

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(updateCommand, c.Name, now, c.ID)

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

		row := tx.QueryRow(selectCommand, c.ID)

		if err := row.Scan(
			&updated.ID,
			&updated.Name,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		); err != nil {
			return err
		}

		c.ID = updated.ID
		c.Name = updated.Name
		c.CreatedAt = updated.CreatedAt
		c.UpdatedAt = updated.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: updating command: %w", err)
	}

	return nil
}

// UpdateCommandItem validates and updates a command item.
func (r *Repos) UpdateCommandItem(commandID uint64, item *CommandItem) error {
	if commandID == 0 {
		return ErrCommandIDRequired
	}

	if item.ID == 0 {
		return ErrCommandItemIDRequired
	}

	if item.Script == "" {
		return ErrCommandItemScriptRequired
	}

	now := time.Now().UTC()

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(updateCommandItem, item.Script, now, item.ID, commandID)

		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()

		if err != nil {
			return err
		}

		if affected == 0 {
			return ErrCommandItemNotFound
		}

		var updated CommandItem

		row := tx.QueryRow(selectCommandItem, commandID, item.ID)

		if err := row.Scan(
			&updated.ID,
			&updated.CommandID,
			&updated.Script,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		); err != nil {
			return err
		}

		item.ID = updated.ID
		item.CommandID = updated.CommandID
		item.Script = updated.Script
		item.CreatedAt = updated.CreatedAt
		item.UpdatedAt = updated.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: updating command item: %w", err)
	}

	return nil
}

// DeleteCommand removes a command and its items.
// If the command does not exist, it returns nil (idempotent).
func (r *Repos) DeleteCommand(id uint64) error {
	if id == 0 {
		return ErrCommandIDRequired
	}

	err := r.Store.WithTx(func(tx *store.Tx) error {
		if _, err := tx.Exec(deleteCommandItemsByCommand, id); err != nil {
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

// RemoveCommandItem removes an item from a command.
// If the item does not exist, it returns nil (idempotent).
func (r *Repos) RemoveCommandItem(commandID, itemID uint64) error {
	if commandID == 0 {
		return ErrCommandIDRequired
	}

	if itemID == 0 {
		return ErrCommandItemIDRequired
	}

	err := r.Store.WithTx(func(tx *store.Tx) error {
		_, err := tx.Exec(deleteCommandItem, itemID, commandID)
		return err
	})
	
	if err != nil {
		return fmt.Errorf("repositories: removing command item: %w", err)
	}
	
	return nil
}
