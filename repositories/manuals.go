package repositories

import (
	"errors"
	"fmt"
	"time"

	"tartarus/store"
)

// Manual is a document attached to a command, explaining what it does and how to use it.
type Manual struct {
	ID          uint64
	Name        string
	Body        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrManualNotFound    = errors.New("manual not found")
	ErrManualNameRequired = errors.New("manual name is required")
	ErrManualBodyRequired = errors.New("manual body is required")
	ErrManualIDRequired   = errors.New("manual id is required")
)

const (
	selectManuals = `
SELECT id, name, body, COALESCE(description, ''), created_at, updated_at
FROM manuals
ORDER BY id
`

	selectManualQuery = `
SELECT id, name, body, COALESCE(description, ''), created_at, updated_at
FROM manuals
WHERE id = ?
`

	insertManualQuery = `
INSERT INTO manuals (name, body, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
`

	updateManualQuery = `
UPDATE manuals
SET name = ?, body = ?, description = ?, updated_at = ?
WHERE id = ?
`

	deleteManualQuery = `
DELETE FROM manuals
WHERE id = ?
`
)

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
			&manual.Description,
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

// GetManual returns the manual with the given id.
// If no manual is found, it returns an empty Manual and nil error.
func (r *Repos) GetManual(id uint64) (Manual, error) {
	rows, err := r.Store.Query(selectManualQuery, id)

	if err != nil {
		return Manual{}, fmt.Errorf("repositories: selecting manual: %w", err)
	}

	defer rows.Close()

	var manual Manual

	if rows.Next() {
		if err := rows.Scan(
			&manual.ID,
			&manual.Name,
			&manual.Body,
			&manual.Description,
			&manual.CreatedAt,
			&manual.UpdatedAt,
		); err != nil {
			return Manual{}, fmt.Errorf("repositories: scanning manual: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return Manual{}, fmt.Errorf("repositories: iterating manual: %w", err)
	}

	return manual, nil
}

// InsertManual validates and persists a new manual.
// On success, m is populated with the inserted values (ID, timestamps).
func (r *Repos) InsertManual(m *Manual) error {
	if m.Name == "" {
		return ErrManualNameRequired
	}

	if m.Body == "" {
		return ErrManualBodyRequired
	}

	now := time.Now().UTC()

	var id int64

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(insertManualQuery, m.Name, m.Body, m.Description, now, now)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()

		if err != nil {
			return err
		}

		var inserted Manual

		row := tx.QueryRow(selectManualQuery, uint64(id))

		if err := row.Scan(
			&inserted.ID,
			&inserted.Name,
			&inserted.Body,
			&inserted.Description,
			&inserted.CreatedAt,
			&inserted.UpdatedAt,
		); err != nil {
			return err
		}

		m.ID = inserted.ID
		m.Name = inserted.Name
		m.Body = inserted.Body
		m.Description = inserted.Description
		m.CreatedAt = inserted.CreatedAt
		m.UpdatedAt = inserted.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: inserting manual: %w", err)
	}

	return nil
}

// UpdateManual validates and updates an existing manual.
// If no row is affected, it returns "manual not found".
func (r *Repos) UpdateManual(m *Manual) error {
	if m.ID == 0 {
		return ErrManualIDRequired
	}

	if m.Name == "" {
		return ErrManualNameRequired
	}

	if m.Body == "" {
		return ErrManualBodyRequired
	}

	now := time.Now().UTC()

	err := r.Store.WithTx(func(tx *store.Tx) error {
		res, err := tx.Exec(updateManualQuery, m.Name, m.Body, m.Description, now, m.ID)

		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()

		if err != nil {
			return err
		}

		if affected == 0 {
			return ErrManualNotFound
		}

		var updated Manual

		row := tx.QueryRow(selectManualQuery, m.ID)

		if err := row.Scan(
			&updated.ID,
			&updated.Name,
			&updated.Body,
			&updated.Description,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		); err != nil {

			return err
		}

		m.ID = updated.ID
		m.Name = updated.Name
		m.Body = updated.Body
		m.Description = updated.Description
		m.CreatedAt = updated.CreatedAt
		m.UpdatedAt = updated.UpdatedAt

		return nil
	})

	if err != nil {
		return fmt.Errorf("repositories: updating manual: %w", err)
	}

	return nil
}

// DeleteManual removes the manual with the given id.
// If the id does not exist, it returns nil (idempotent).
func (r *Repos) DeleteManual(id uint64) error {
	if id == 0 {
		return ErrManualIDRequired
	}

	err := r.Store.WithTx(func(tx *store.Tx) error {
		_, err := tx.Exec(deleteManualQuery, id)
		return err
	})

	if err != nil {
		return fmt.Errorf("repositories: deleting manual: %w", err)
	}

	return nil
}
