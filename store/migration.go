package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
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
	currentVersion uint64 = 1

	//go:embed migrations/0001_create_migrations_table.up.sql
	createMigrationTableUp string

	//go:embed queries/0001_check_last_version.sql
	checkLastVersionQuery string

	ups   map[uint64]string
	downs map[uint64]string
)

func init() {
	ups = map[uint64]string{}
	downs = map[uint64]string{}
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
	cur := currentVersion

	if stored < cur {
		for v := stored + 1; v <= cur; v++ {
			q, ok := ups[v]

			if !ok || q == "" {
				return fmt.Errorf("missing up migration for version %d", v)
			}

			if _, err := s.db.Exec(q); err != nil {
				return err
			}
		}

		return nil
	}

	if stored > cur {
		for v := stored; v > cur; v-- {
			q, ok := downs[v]

			if !ok || q == "" {
				return fmt.Errorf("missing down migration for version %d", v)
			}

			if _, err := s.db.Exec(q); err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
