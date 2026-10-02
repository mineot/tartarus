package repositories

import (
	"errors"
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetCommands() ([]Command, error) {
	store := store.Store{}

	if err := store.Open(); err != nil {
		return nil, err
	}

	query := `select id, "name", created_at, updated_at from commands`

	rows, err := store.Query(query)

	if err != nil {
		return nil, err
	}

	defer store.Close()

	var commands []Command

	for rows.Next() {
		var command Command

		if err := rows.Scan(
			&command.ID,
			&command.Name,
			&command.CreatedAt,
			&command.UpdatedAt,
		); err != nil {
			return nil, err
		}

		commands = append(commands, command)
	}

	return []Command{}, nil
}

func GetCommand(ID uint64) (Command, error) {
	store := store.Store{}

	if err := store.Open(); err != nil {
		return Command{}, err
	}

	query := `select id, "name", created_at, updated_at from commands where id = ?`

	rows, err := store.Query(query, ID)

	if err != nil {
		return Command{}, err
	}

	defer store.Close()

	var command Command

	if rows.Next() {
		if err := rows.Scan(
			&command.ID,
			&command.Name,
			&command.CreatedAt,
			&command.UpdatedAt,
		); err != nil {
			return Command{}, err
		}
	}

	return command, nil
}

func (c *Command) GetItems() ([]CommandItem, error) {
	store := store.Store{}

	if err := store.Open(); err != nil {
		return nil, err
	}

	query := `select id, command_id, script, created_at, updated_at from command_items where command_id = ?`

	rows, err := store.Query(query, c.ID)

	if err != nil {
		return nil, err
	}

	defer store.Close()

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
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (c *Command) GetItem(ID uint64) (CommandItem, error) {
	store := store.Store{}

	if err := store.Open(); err != nil {
		return CommandItem{}, err
	}

	query := `select id, command_id, script, created_at, updated_at from command_items where command_id = ? and id = ?`

	rows, err := store.Query(query, c.ID, ID)

	if err != nil {
		return CommandItem{}, err
	}

	defer store.Close()

	var item CommandItem

	if rows.Next() {
		if err := rows.Scan(
			&item.ID,
			&item.CommandID,
			&item.Script,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return CommandItem{}, err
		}
	}

	return item, nil
}

func (c *Command) Insert() error {
	if c.Name == "" {
		return errors.New("command name is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `insert into commands (name, created_at, updated_at) values (?, ?, ?)`

	result, err := store.Exec(query, c.Name, time.Now(), time.Now())

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	ID, err := result.LastInsertId()

	if err != nil {
		return err
	}

	insertedCommand, err := GetCommand(uint64(ID))

	if err != nil {
		store.Rollback()
		return err
	}

	c.ID = insertedCommand.ID
	c.Name = insertedCommand.Name
	c.CreatedAt = insertedCommand.CreatedAt
	c.UpdatedAt = insertedCommand.UpdatedAt

	return nil
}

func (c *Command) AppendItem(item CommandItem) error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	if item.Script == "" {
		return errors.New("command script is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `insert into command_items (command_id, script, created_at, updated_at) values (?, ?, ?, ?)`

	_, err := store.Exec(query, c.ID, item.Script, time.Now(), time.Now())

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	return nil
}

func (c *Command) Update() error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	if c.Name == "" {
		return errors.New("command name is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `update commands set name = ?, updated_at = ? where id = ?`

	result, err := store.Exec(query, c.Name, time.Now(), c.ID)

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	ID, err := result.RowsAffected()

	if err != nil {
		return err
	}

	updatedCommand, err := GetCommand(uint64(ID))

	c.ID = updatedCommand.ID
	c.Name = updatedCommand.Name
	c.CreatedAt = updatedCommand.CreatedAt
	c.UpdatedAt = updatedCommand.UpdatedAt

	return nil
}

func (c *Command) UpdateItem(item CommandItem) error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	if item.ID == 0 {
		return errors.New("command item id is required")
	}

	if item.Script == "" {
		return errors.New("command item script is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `update command_items set script = ?, updated_at = ? where id = ? and command_id = ?`

	_, err := store.Exec(query, item.Script, time.Now(), c.ID, item.ID)

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	return nil
}

func (c *Command) Delete() error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `delete from command_items where command_id = ?`

	_, err := store.Exec(query, c.ID)

	if err != nil {
		store.Rollback()
		return err
	}

	query = `delete from commands where id = ?`

	_, err = store.Exec(query, c.ID)

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	return nil
}

func (c *Command) RemoveItem(ID uint64) error {
	if ID == 0 {
		return errors.New("command item id is required")
	}

	if c.ID == 0 {
		return errors.New("command id is required")
	}

	store := store.Store{}

	store.Begin()

	if err := store.Open(); err != nil {
		store.Rollback()
		return err
	}

	query := `delete from command_items where command_id = ? and id = ?`

	_, err := store.Exec(query, c.ID, ID)

	if err != nil {
		store.Rollback()
		return err
	}

	store.Commit()

	defer store.Close()

	return nil
}
