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
	Items     []CommandItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetCommands() ([]Command, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return nil, err
	}

	defer st.Close()

	query := `select id, "name", created_at, updated_at from commands`

	rows, err := st.Query(query)

	if err != nil {
		return nil, err
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
			return nil, err
		}

		commands = append(commands, command)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func GetCommand(ID uint64) (Command, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return Command{}, err
	}

	defer st.Close()

	query := `select id, "name", created_at, updated_at from commands where id = ?`

	rows, err := st.Query(query, ID)

	if err != nil {
		return Command{}, err
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
			return Command{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return Command{}, err
	}

	return command, nil
}

func (c *Command) GetItems() ([]CommandItem, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return nil, err
	}

	defer st.Close()

	query := `select id, command_id, script, created_at, updated_at from command_items where command_id = ?`

	rows, err := st.Query(query, c.ID)

	if err != nil {
		return nil, err
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
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Command) GetItem(ID uint64) (CommandItem, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return CommandItem{}, err
	}

	defer st.Close()

	query := `select id, command_id, script, created_at, updated_at from command_items where command_id = ? and id = ?`

	rows, err := st.Query(query, c.ID, ID)

	if err != nil {
		return CommandItem{}, err
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
			return CommandItem{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return CommandItem{}, err
	}

	return item, nil
}

func (c *Command) Insert() error {
	if c.Name == "" {
		return errors.New("command name is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `insert into commands (name, created_at, updated_at) values (?, ?, ?)`

	result, err := st.Exec(query, c.Name, time.Now(), time.Now())

	if err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	ID, err := result.LastInsertId()

	if err != nil {
		return err
	}

	insertedCommand, err := GetCommand(uint64(ID))

	if err != nil {
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

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `insert into command_items (command_id, script, created_at, updated_at) values (?, ?, ?, ?)`

	if _, err := st.Exec(query, c.ID, item.Script, time.Now(), time.Now()); err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	var err error

	if c.Items, err = c.GetItems(); err != nil {
		return err
	}

	return nil
}

func (c *Command) Update() error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	if c.Name == "" {
		return errors.New("command name is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `update commands set name = ?, updated_at = ? where id = ?`

	result, err := st.Exec(query, c.Name, time.Now(), c.ID)

	if err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("command not found")
	}

	updatedCommand, err := GetCommand(c.ID)

	if err != nil {
		return err
	}

	c.Name = updatedCommand.Name
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

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `update command_items set script = ?, updated_at = ? where id = ? and command_id = ?`

	if _, err := st.Exec(query, item.Script, time.Now(), item.ID, c.ID); err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	var err error

	if c.Items, err = c.GetItems(); err != nil {
		return err
	}

	return nil
}

func (c *Command) Delete() error {
	if c.ID == 0 {
		return errors.New("command id is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `delete from command_items where command_id = ?`

	if _, err := st.Exec(query, c.ID); err != nil {
		st.Rollback()
		return err
	}

	query = `delete from commands where id = ?`

	if _, err := st.Exec(query, c.ID); err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *Command) RemoveItem(ID uint64) error {
	if ID == 0 {
		return errors.New("command item id is required")
	}

	if c.ID == 0 {
		return errors.New("command id is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `delete from command_items where command_id = ? and id = ?`

	if _, err := st.Exec(query, c.ID, ID); err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	var err error

	if c.Items, err = c.GetItems(); err != nil {
		return err
	}

	return nil
}
