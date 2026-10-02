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
	currentVersion int = 0

	//go:embed migrations/0001_create_migrations_table.up.sql
	createMigrationTableUp string

	//go:embed queries/0001_check_last_version.sql
	checkLastVersionQuery string

	//go:embed queries/0002_insert_migration.sql
	insertMigrationQuery string

	//go:embed queries/0003_delete_migration.sql
	deleteMigrationQuery string

	ups   map[uint64]string
	downs map[uint64]string
)

func init() {
	ups = map[uint64]string{}
	downs = map[uint64]string{}
}

func migrationDate() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func (s *Store) RunMigrations() error {
	if _, err := s.db.Exec(createMigrationTableUp); err != nil {
		return err
	}

	var lastVersion Migration

	err := s.db.QueryRow(checkLastVersionQuery).Scan(
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

			if _, err := tx.Exec(insertMigrationQuery, migrationDate(), v); err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit()
	}

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

			if _, err := tx.Exec(deleteMigrationQuery, v); err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit()
	}

	return nil
}
