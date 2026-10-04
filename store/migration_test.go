package store

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func nowForTest() time.Time {
	return time.Now().UTC()
}

// newEmptyStore returns a Store pointing at a database with no tables at all.
//
// newTestStore creates an items table, which would get in the way of asserting
// on what RunMigrations builds from scratch.
func newEmptyStore(t *testing.T) *Store {
	t.Helper()

	s, err := newAt(context.Background(), filepath.Join(t.TempDir(), "test.db"))

	if err != nil {
		t.Fatalf("newAt: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	return s
}

// tableNames lists the user tables, sorted, with SQLite's internal ones left out.
func tableNames(t *testing.T, s *Store) []string {
	t.Helper()

	rows, err := s.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)

	if err != nil {
		t.Fatalf("listing tables: %v", err)
	}

	defer rows.Close()

	var names []string

	for rows.Next() {
		var name string

		if err = rows.Scan(&name); err != nil {
			t.Fatalf("scanning table name: %v", err)
		}

		names = append(names, name)
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("iterating tables: %v", err)
	}

	return names
}

// rowCount counts the rows of a table chosen by the caller, so the query is never
// built by string interpolation.
func rowCount(t *testing.T, s *Store, table string) int {
	t.Helper()

	var query string

	switch table {
	case "commands":
		query = `SELECT COUNT(*) FROM commands`
	case "command_items":
		query = `SELECT COUNT(*) FROM command_items`
	case "migrations":
		query = `SELECT COUNT(*) FROM migrations`
	default:
		t.Fatalf("rowCount: unknown table %q", table)
	}

	rows, err := s.Query(query)

	if err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}

	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("counting %s: no row returned", table)
	}

	var n int

	if err = rows.Scan(&n); err != nil {
		t.Fatalf("scanning count of %s: %v", table, err)
	}

	return n
}

func TestRunMigrationsCreatesSchema(t *testing.T) {
	s := newEmptyStore(t)

	if got := tableNames(t, s); len(got) != 0 {
		t.Fatalf("esperava um banco sem tabelas, obteve %v", got)
	}

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("esperava %v, obteve %v", want, got)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("esperava 1 versão registrada, obteve %d", got)
	}

	const selectVersion = `SELECT version_number FROM migrations`

	rows, err := s.Query(selectVersion)

	if err != nil {
		t.Fatalf("reading version_number: %v", err)
	}

	defer rows.Close()

	if !rows.Next() {
		t.Fatal("esperava uma linha de migrations, não veio nenhuma")
	}

	var version int

	if err = rows.Scan(&version); err != nil {
		t.Fatalf("scanning version_number: %v", err)
	}

	if want := int(currentVersion); version != want {
		t.Fatalf("esperava a versão %d, obteve %d", want, version)
	}
}

func TestRunMigrationsIsIdempotent(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("primeira RunMigrations: %v", err)
	}

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("segunda RunMigrations: %v", err)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("a segunda chamada deveria ser um no-op, obteve %d versões", got)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("esperava %v, obteve %v", want, got)
	}
}

func TestResetMigrationsRecreatesSchema(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	const insertCommand = `INSERT INTO commands (name, created_at, updated_at) VALUES (?, ?, ?)`

	if err := s.WithTx(func(tx *Tx) error {
		_, err := tx.Exec(insertCommand, "leftover", nowForTest(), nowForTest())

		return err
	}); err != nil {
		t.Fatalf("inserting command: %v", err)
	}

	if got := rowCount(t, s, "commands"); got != 1 {
		t.Fatalf("esperava 1 command antes do reset, obteve %d", got)
	}

	if err := s.ResetMigrations(); err != nil {
		t.Fatalf("ResetMigrations: %v", err)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("esperava %v, obteve %v", want, got)
	}

	if got := rowCount(t, s, "commands"); got != 0 {
		t.Fatalf("o reset deveria ter esvaziado commands, obteve %d", got)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("esperava a versão 1 re-registrada, obteve %d", got)
	}
}

// TestForeignKeysAreEnabled is the regression test for the pragma that never
// worked: foreign keys used to be turned on inside the migration, where SQLite
// ignores it. They now come from the DSN.
func TestForeignKeysAreEnabled(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	rows, err := s.Query(`PRAGMA foreign_keys`)

	if err != nil {
		t.Fatalf("reading PRAGMA foreign_keys: %v", err)
	}

	defer rows.Close()

	if !rows.Next() {
		t.Fatal("PRAGMA foreign_keys não devolveu nada")
	}

	var enabled int

	if err = rows.Scan(&enabled); err != nil {
		t.Fatalf("scanning PRAGMA foreign_keys: %v", err)
	}

	if enabled != 1 {
		t.Fatalf("esperava foreign_keys ligado, obteve %d", enabled)
	}
}

// TestForeignKeyRejectsOrphanItem is the other half: an item pointing at a
// command that does not exist has to be refused.
func TestForeignKeyRejectsOrphanItem(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	const insertItem = `INSERT INTO command_items (command_id, script, created_at, updated_at) VALUES (?, ?, ?, ?)`

	err := s.WithTx(func(tx *Tx) error {
		_, err := tx.Exec(insertItem, 999, "ls -la", nowForTest(), nowForTest())

		return err
	})

	if err == nil {
		t.Fatal("um command_items com command_id inexistente deveria ter falhado")
	}

	if !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatalf("esperava um erro de FOREIGN KEY, obteve %v", err)
	}

	if got := rowCount(t, s, "command_items"); got != 0 {
		t.Fatalf("o insert recusado não deveria sobrar, obteve %d", got)
	}
}

// TestForeignKeyCascadeDeletesItems checks the ON DELETE CASCADE that only takes
// effect with enforcement on.
func TestForeignKeyCascadeDeletesItems(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	const insertCommand = `INSERT INTO commands (name, created_at, updated_at) VALUES (?, ?, ?)`
	const insertItem = `INSERT INTO command_items (command_id, script, created_at, updated_at) VALUES (?, ?, ?, ?)`

	if err := s.WithTx(func(tx *Tx) error {
		if _, err := tx.Exec(insertCommand, "build", nowForTest(), nowForTest()); err != nil {
			return err
		}

		_, err := tx.Exec(insertItem, 1, "go build ./...", nowForTest(), nowForTest())

		return err
	}); err != nil {
		t.Fatalf("inserting: %v", err)
	}

	if got := rowCount(t, s, "command_items"); got != 1 {
		t.Fatalf("esperava 1 command_item, obteve %d", got)
	}

	const deleteCommand = `DELETE FROM commands WHERE id = 1`

	if err := s.WithTx(func(tx *Tx) error {
		_, err := tx.Exec(deleteCommand)

		return err
	}); err != nil {
		t.Fatalf("deleting command: %v", err)
	}

	if got := rowCount(t, s, "command_items"); got != 0 {
		t.Fatalf("o cascade deveria ter apagado o command_item, obteve %d", got)
	}

	if got := rowCount(t, s, "commands"); got != 0 {
		t.Fatalf("esperava 0 commands, obteve %d", got)
	}
}
