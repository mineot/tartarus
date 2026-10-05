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

func newEmptyStore(t *testing.T) *Store {
	t.Helper()

	s, err := newAt(context.Background(), filepath.Join(t.TempDir(), "test.db"))

	if err != nil {
		t.Fatalf("newAt: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	return s
}

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
		t.Fatalf("want a database with no tables, got %v", got)
	}

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("want 1 recorded version, got %d", got)
	}

	const selectVersion = `SELECT version_number FROM migrations`

	rows, err := s.Query(selectVersion)

	if err != nil {
		t.Fatalf("reading version_number: %v", err)
	}

	defer rows.Close()

	if !rows.Next() {
		t.Fatal("want a row in migrations, got none")
	}

	var version int

	if err = rows.Scan(&version); err != nil {
		t.Fatalf("scanning version_number: %v", err)
	}

	if want := int(currentVersion); version != want {
		t.Fatalf("want version %d, got %d", want, version)
	}
}

func TestRunMigrationsIsIdempotent(t *testing.T) {
	s := newEmptyStore(t)

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("first RunMigrations: %v", err)
	}

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("second RunMigrations: %v", err)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("the second call should be a no-op, got %d versions", got)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("want %v, got %v", want, got)
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
		t.Fatalf("want 1 command before the reset, got %d", got)
	}

	if err := s.ResetMigrations(); err != nil {
		t.Fatalf("ResetMigrations: %v", err)
	}

	want := []string{"command_items", "commands", "manuals", "migrations"}

	if got := tableNames(t, s); !slices.Equal(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}

	if got := rowCount(t, s, "commands"); got != 0 {
		t.Fatalf("the reset should have emptied commands, got %d", got)
	}

	if got := rowCount(t, s, "migrations"); got != 1 {
		t.Fatalf("want version 1 registered again, got %d", got)
	}
}

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
		t.Fatal("PRAGMA foreign_keys returned nothing")
	}

	var enabled int

	if err = rows.Scan(&enabled); err != nil {
		t.Fatalf("scanning PRAGMA foreign_keys: %v", err)
	}

	if enabled != 1 {
		t.Fatalf("want foreign_keys on, got %d", enabled)
	}
}

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
		t.Fatal("a command_items row with a nonexistent command_id should have failed")
	}

	if !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatalf("want a FOREIGN KEY error, got %v", err)
	}

	if got := rowCount(t, s, "command_items"); got != 0 {
		t.Fatalf("the refused insert should not have survived, got %d", got)
	}
}

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
		t.Fatalf("want 1 command_item, got %d", got)
	}

	const deleteCommand = `DELETE FROM commands WHERE id = 1`

	if err := s.WithTx(func(tx *Tx) error {
		_, err := tx.Exec(deleteCommand)

		return err
	}); err != nil {
		t.Fatalf("deleting command: %v", err)
	}

	if got := rowCount(t, s, "command_items"); got != 0 {
		t.Fatalf("the cascade should have deleted the command_item, got %d", got)
	}

	if got := rowCount(t, s, "commands"); got != 0 {
		t.Fatalf("want 0 commands, got %d", got)
	}
}
