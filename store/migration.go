package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "embed"
)

type Migration struct {
	id            uint64
	versionDate   string
	versionNumber uint64
}

type query struct {
	up   string
	down string
}

var (
	currentVersion int = 1
	ups            map[uint64]string
	downs          map[uint64]string

	//go:embed migrations/0000_drop_all_tables.sql
	dropAllTables string

	//go:embed migrations/0001_create_version_one.up.sql
	upVersionOne string

	//go:embed migrations/0001_create_version_one.down.sql
	downVersionOne string
)

func init() {
	ups = map[uint64]string{
		1: upVersionOne,
	}
	downs = map[uint64]string{
		1: downVersionOne,
	}
}

func migrationDate() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func up(stored uint64, cur uint64, s *Store) error {
	if stored < cur {
		tx, err := s.db.Begin()

		if err != nil {
			return err
		}

		for v := stored + 1; v <= cur; v++ {
			q, ok := ups[v]

			if !ok || q == "" {
				tx.Rollback()
				return fmt.Errorf("missing up migration for version %d", v)
			}

			if _, err := tx.Exec(q); err != nil {
				tx.Rollback()
				return err
			}

			query := `INSERT INTO migrations (versionDate, versionNumber) VALUES (?, ?)`

			if _, err := tx.Exec(query, migrationDate(), v); err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit()
	}

	return nil
}

func down(stored uint64, cur uint64, s *Store) error {
	if stored > cur {
		tx, err := s.db.Begin()

		if err != nil {
			return err
		}

		for v := stored; v > cur; v-- {
			q, ok := downs[v]

			if !ok || q == "" {
				tx.Rollback()
				return fmt.Errorf("missing down migration for version %d", v)
			}

			if _, err := tx.Exec(q); err != nil {
				tx.Rollback()
				return err
			}

			query := `DELETE FROM migrations WHERE versionNumber = ?`

			if _, err := tx.Exec(query, v); err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit()
	}

	return nil
}

func (s *Store) RunMigrations() error {
	var err error

	var query = `
	CREATE TABLE IF NOT EXISTS migrations (
		id				INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		versionDate		TEXT NOT NULL,
		versionNumber	INTEGER NOT NULL
	)
	`

	if _, err = s.db.Exec(query); err != nil {
		return err
	}

	var lastVersion Migration

	query = `SELECT id, versionDate, versionNumber FROM migrations ORDER BY versionNumber DESC, id DESC LIMIT 1`

	err = s.db.QueryRow(query).Scan(
		&lastVersion.id,
		&lastVersion.versionDate,
		&lastVersion.versionNumber,
	)

	if errors.Is(err, sql.ErrNoRows) {
		lastVersion.versionNumber = 0
	} else if err != nil {
		return err
	}

	stored := lastVersion.versionNumber
	cur := uint64(currentVersion)

	if err = up(stored, cur, s); err != nil {
		return err
	}

	if err = down(stored, cur, s); err != nil {
		return err
	}

	return nil
}

func (s *Store) ResetMigrations() error {
	if err := s.Open(); err != nil {
		return err
	}

	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	if _, err := s.Exec(dropAllTables); err != nil {
		s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	if err := s.RunMigrations(); err != nil {
		s.Rollback()
		return err
	}

	return nil
}
