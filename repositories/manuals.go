package repositories

import (
	"errors"
	"tartarus/store"
	"time"
)

type Manual struct {
	ID        uint64
	Name      string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetManuals() ([]Manual, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return nil, err
	}

	defer st.Close()

	query := `select id, name, body, created_at, updated_at from manuals`

	rows, err := st.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var manuals []Manual

	for rows.Next() {
		var manual Manual

		if err := rows.Scan(
			&manual.ID,
			&manual.Name,
			&manual.Body,
			&manual.CreatedAt,
			&manual.UpdatedAt,
		); err != nil {
			return nil, err
		}

		manuals = append(manuals, manual)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return manuals, nil
}

func GetManual(ID uint64) (Manual, error) {
	st := store.Store{}

	if err := st.Open(); err != nil {
		return Manual{}, err
	}

	defer st.Close()

	query := `select id, name, body, created_at, updated_at from manuals where id = ?`

	rows, err := st.Query(query, ID)

	if err != nil {
		return Manual{}, err
	}

	defer rows.Close()

	var manual Manual

	if rows.Next() {
		if err := rows.Scan(
			&manual.ID,
			&manual.Name,
			&manual.Body,
			&manual.CreatedAt,
			&manual.UpdatedAt,
		); err != nil {
			return Manual{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return Manual{}, err
	}

	return manual, nil
}

func (m *Manual) Insert() error {
	if m.Name == "" {
		return errors.New("manual name is required")
	}

	if m.Body == "" {
		return errors.New("manual body is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `insert into manuals (name, body, created_at, updated_at) values (?, ?, ?, ?)`

	result, err := st.Exec(query, m.Name, m.Body, time.Now(), time.Now())

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

	insertedCommand, err := GetManual(uint64(ID))

	if err != nil {
		return err
	}

	m.ID = insertedCommand.ID
	m.Name = insertedCommand.Name
	m.Body = insertedCommand.Body
	m.CreatedAt = insertedCommand.CreatedAt
	m.UpdatedAt = insertedCommand.UpdatedAt

	return nil
}

func (m *Manual) Update() error {
	if m.ID == 0 {
		return errors.New("manual id is required")
	}

	if m.Name == "" {
		return errors.New("manual name is required")
	}

	if m.Body == "" {
		return errors.New("manual body is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `update manuals set name = ?, body = ?, updated_at = ? where id = ?`

	result, err := st.Exec(query, m.Name, m.Body, time.Now(), m.ID)

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

	updatedCommand, err := GetManual(m.ID)

	if err != nil {
		return err
	}

	m.ID = updatedCommand.ID
	m.Name = updatedCommand.Name
	m.Body = updatedCommand.Body
	m.CreatedAt = updatedCommand.CreatedAt
	m.UpdatedAt = updatedCommand.UpdatedAt

	return nil
}

func (m *Manual) Delete() error {
	if m.ID == 0 {
		return errors.New("manual id is required")
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.Begin(); err != nil {
		return err
	}

	query := `delete from manuals where id = ?`

	_, err := st.Exec(query, m.ID)

	if err != nil {
		st.Rollback()
		return err
	}

	if err := st.Commit(); err != nil {
		return err
	}

	return nil
}
