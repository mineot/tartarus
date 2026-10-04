package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"
)

// Migration is one row of the migrations bookkeeping table: which version was
// applied, and when.
//
// Nothing outside this file reads it, which is why every field is private. Only
// versionNumber carries meaning for the logic; id and versionDate are scanned
// because the table has those columns, and the order of the Scan has to match
// the order of the columns in the SELECT.
type Migration struct {
	id            uint64
	versionDate   time.Time
	versionNumber uint64
}

// query is the pair of SQL statements that moves the schema one version forward
// or one version back. Both halves are embedded from store/migrations.
type query struct {
	up   string
	down string
}

var (
	// currentVersion is the schema version this binary expects the database to be
	// at. Moving it means adding the next numbered pair of files under
	// store/migrations and an entry to queries below.
	//
	// It is an int because that is what the version number is stored as; up and
	// down take a uint64 and convert at the call site.
	currentVersion int = 1

	// queries maps a version number to its up and down SQL.
	//
	// A version with an empty statement counts as missing, so a half-added entry
	// fails loudly instead of silently doing nothing and leaving the database
	// claiming a version it never actually reached.
	queries map[uint64]query

	//go:embed migrations/0000_drop_all_tables.sql
	dropAllTables string

	//go:embed migrations/0001_create_version_one.up.sql
	upVersionOne string

	//go:embed migrations/0001_create_version_one.down.sql
	downVersionOne string
)

// init wires the embedded SQL into queries, so adding a migration means adding
// its 000N files and one entry here.
func init() {
	queries = map[uint64]query{
		1: {
			up:   upVersionOne,
			down: downVersionOne,
		},
	}
}

// up applies every migration after stored, up to and including cur, inside tx.
//
// It runs in the caller's transaction instead of opening its own, so a failure
// halfway through leaves the database as it was.
func up(stored uint64, cur uint64, tx *Tx) error {
	if stored >= cur {
		return nil
	}

	// One timestamp for the whole batch, so applying several versions in a row
	// records them as applied at the same moment, which is when they were
	// actually applied.
	now := time.Now().UTC()

	for v := stored + 1; v <= cur; v++ {
		q, ok := queries[v]

		if !ok || q.up == "" {
			return fmt.Errorf("store: missing up migration for version %d", v)
		}

		if _, err := tx.Exec(q.up); err != nil {
			return fmt.Errorf("store: applying up migration %d: %w", v, err)
		}

		const insertVersion = `INSERT INTO migrations (version_date, version_number) VALUES (?, ?)`

		if _, err := tx.Exec(insertVersion, now, v); err != nil {
			return fmt.Errorf("store: recording version %d: %w", v, err)
		}
	}

	return nil
}

// down rolls the schema back from stored down to cur, inside tx.
//
// It only runs when the database is ahead of the binary, which means an older
// build was opened against a newer schema.
func down(stored uint64, cur uint64, tx *Tx) error {
	if stored <= cur {
		return nil
	}

	for v := stored; v > cur; v-- {
		q, ok := queries[v]

		if !ok || q.down == "" {
			return fmt.Errorf("store: missing down migration for version %d", v)
		}

		if _, err := tx.Exec(q.down); err != nil {
			return fmt.Errorf("store: applying down migration %d: %w", v, err)
		}

		const deleteVersion = `DELETE FROM migrations WHERE version_number = ?`

		if _, err := tx.Exec(deleteVersion, v); err != nil {
			return fmt.Errorf("store: removing version %d: %w", v, err)
		}
	}

	return nil
}

// runMigrations brings the schema up to currentVersion.
//
// It assumes it is already inside a transaction and does not open one. Both
// callers need to be atomic, and a shared body that began its own transaction
// would hit ErrTxActive coming from the one they already started.
func runMigrations(tx *Tx) error {
	// The table comes first because storedVersion reads it, and on a fresh
	// database creating it is also what makes storedVersion return zero instead
	// of failing.
	if err := ensureMigrationsTable(tx); err != nil {
		return err
	}

	stored, err := storedVersion(tx)

	if err != nil {
		return err
	}

	cur := uint64(currentVersion)

	// up and down are mutually exclusive, they each return early when there is
	// nothing to do. up comes first because being behind is the ordinary case.
	if err = up(stored, cur, tx); err != nil {
		return err
	}

	return down(stored, cur, tx)
}

// RunMigrations applies whatever migrations the database is still missing.
//
// It is idempotent: a database already at currentVersion is left alone, so it is
// safe to call on every start.
func (s *Store) RunMigrations() error {
	return s.WithTx(func(tx *Tx) error {
		return runMigrations(tx)
	})
}

// ResetMigrations drops every table and rebuilds the schema from scratch.
//
// The drop and the rebuild are a single transaction. Committing the drop before
// migrating would leave the database without a schema in between, and a failure
// after that point would leave it that way. There is no test for the rollback
// path; the guarantee is the single transaction.
func (s *Store) ResetMigrations() error {
	return s.WithTx(func(tx *Tx) error {
		// dropAllTables has to list command_items before commands. With foreign
		// keys enforced, dropping a table that other tables point at performs an
		// implicit delete of its rows, which fails on an immediate constraint
		// while the children are still there. The order is a contract between
		// that file and the DSN in store.go.
		if _, err := tx.Exec(dropAllTables); err != nil {
			return fmt.Errorf("store: dropping tables: %w", err)
		}

		// runMigrations, not RunMigrations: we are already inside a transaction,
		// and the exported one would open a second and hit ErrTxActive.
		return runMigrations(tx)
	})
}

// ensureMigrationsTable creates the bookkeeping table if it is missing.
//
// This is also the step that puts the table back after a reset, since
// dropAllTables drops migrations along with everything else.
func ensureMigrationsTable(tx *Tx) error {
	const createMigrations = `
	CREATE TABLE IF NOT EXISTS migrations (
		id				INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		version_date	TIMESTAMP NOT NULL,
		version_number	INTEGER NOT NULL
	)
	`

	if _, err := tx.Exec(createMigrations); err != nil {
		return fmt.Errorf("store: creating migrations table: %w", err)
	}

	return nil
}

// storedVersion returns the highest version number recorded in the database, or
// zero when no migration has been applied yet.
//
// The second sort key is there so a database with a duplicated version number,
// which nothing stops someone from writing by hand, still resolves to a single
// row instead of an arbitrary one.
func storedVersion(tx *Tx) (uint64, error) {
	const selectVersion = `SELECT id, version_date, version_number FROM migrations ORDER BY version_number DESC, id DESC LIMIT 1`

	var last Migration

	err := tx.QueryRow(selectVersion).Scan(
		&last.id,
		&last.versionDate,
		&last.versionNumber,
	)

	// No rows means the table was just created, which is the normal state of a
	// fresh database. Zero is what makes up start at version 1.
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("store: reading the stored version: %w", err)
	}

	return last.versionNumber, nil
}
