package repositories

import (
	"fmt"
	"time"
)

// Manual is a document attached to a command, explaining what it does and how to use it.
type Manual struct {
	ID        uint64
	Name      string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const selectManuals = `
	SELECT id, name, body, created_at, updated_at
	FROM manuals
	ORDER BY id
`

// GetManuals returns every manual, in creation order.
//
// It reads through Store.Query rather than inside a WithTx: a read does not need
// a transaction, and _txlock=immediate would take the write lock to begin with,
// serializing readers against writers for nothing. The cost is that it cannot
// run while a WithTx is in progress on the same Store, which returns
// store.ErrUseTx. Read after the write has committed.
//
// The returned slice is nil when there are no manuals, so callers should test it
// with len.
func (r *Repos) GetManuals() ([]Manual, error) {
	rows, err := r.Store.Query(selectManuals)

	if err != nil {
		return nil, fmt.Errorf("repositories: selecting manuals: %w", err)
	}

	defer rows.Close()

	var manuals []Manual

	for rows.Next() {
		var manual Manual

		if err = rows.Scan(
			&manual.ID,
			&manual.Name,
			&manual.Body,
			&manual.CreatedAt,
			&manual.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scanning manual: %w", err)
		}

		manuals = append(manuals, manual)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterating manuals: %w", err)
	}

	return manuals, nil
}

// func GetManual(ID uint64) (Manual, error) {
// 	st := store.Store{}

// 	if err := st.Open(); err != nil {
// 		return Manual{}, err
// 	}

// 	defer st.Close()

// 	query := `select id, name, body, created_at, updated_at from manuals where id = ?`

// 	rows, err := st.Query(query, ID)

// 	if err != nil {
// 		return Manual{}, err
// 	}

// 	defer rows.Close()

// 	var manual Manual

// 	if rows.Next() {
// 		if err := rows.Scan(
// 			&manual.ID,
// 			&manual.Name,
// 			&manual.Body,
// 			&manual.CreatedAt,
// 			&manual.UpdatedAt,
// 		); err != nil {
// 			return Manual{}, err
// 		}
// 	}

// 	if err := rows.Err(); err != nil {
// 		return Manual{}, err
// 	}

// 	return manual, nil
// }

// func (m *Manual) Insert() error {
// 	if m.Name == "" {
// 		return errors.New("manual name is required")
// 	}

// 	if m.Body == "" {
// 		return errors.New("manual body is required")
// 	}

// 	st := store.Store{}

// 	if err := st.Open(); err != nil {
// 		return err
// 	}

// 	defer st.Close()

// 	if err := st.Begin(); err != nil {
// 		return err
// 	}

// 	query := `insert into manuals (name, body, created_at, updated_at) values (?, ?, ?, ?)`

// 	result, err := st.Exec(query, m.Name, m.Body, time.Now(), time.Now())

// 	if err != nil {
// 		st.Rollback()
// 		return err
// 	}

// 	if err := st.Commit(); err != nil {
// 		return err
// 	}

// 	ID, err := result.LastInsertId()

// 	if err != nil {
// 		return err
// 	}

// 	insertedCommand, err := GetManual(uint64(ID))

// 	if err != nil {
// 		return err
// 	}

// 	m.ID = insertedCommand.ID
// 	m.Name = insertedCommand.Name
// 	m.Body = insertedCommand.Body
// 	m.CreatedAt = insertedCommand.CreatedAt
// 	m.UpdatedAt = insertedCommand.UpdatedAt

// 	return nil
// }

// func (m *Manual) Update() error {
// 	if m.ID == 0 {
// 		return errors.New("manual id is required")
// 	}

// 	if m.Name == "" {
// 		return errors.New("manual name is required")
// 	}

// 	if m.Body == "" {
// 		return errors.New("manual body is required")
// 	}

// 	st := store.Store{}

// 	if err := st.Open(); err != nil {
// 		return err
// 	}

// 	defer st.Close()

// 	if err := st.Begin(); err != nil {
// 		return err
// 	}

// 	query := `update manuals set name = ?, body = ?, updated_at = ? where id = ?`

// 	result, err := st.Exec(query, m.Name, m.Body, time.Now(), m.ID)

// 	if err != nil {
// 		st.Rollback()
// 		return err
// 	}

// 	if err := st.Commit(); err != nil {
// 		return err
// 	}

// 	affected, err := result.RowsAffected()

// 	if err != nil {
// 		return err
// 	}

// 	if affected == 0 {
// 		return errors.New("command not found")
// 	}

// 	updatedCommand, err := GetManual(m.ID)

// 	if err != nil {
// 		return err
// 	}

// 	m.ID = updatedCommand.ID
// 	m.Name = updatedCommand.Name
// 	m.Body = updatedCommand.Body
// 	m.CreatedAt = updatedCommand.CreatedAt
// 	m.UpdatedAt = updatedCommand.UpdatedAt

// 	return nil
// }

// func (m *Manual) Delete() error {
// 	if m.ID == 0 {
// 		return errors.New("manual id is required")
// 	}

// 	st := store.Store{}

// 	if err := st.Open(); err != nil {
// 		return err
// 	}

// 	defer st.Close()

// 	if err := st.Begin(); err != nil {
// 		return err
// 	}

// 	query := `delete from manuals where id = ?`

// 	_, err := st.Exec(query, m.ID)

// 	if err != nil {
// 		st.Rollback()
// 		return err
// 	}

// 	if err := st.Commit(); err != nil {
// 		return err
// 	}

// 	return nil
// }
