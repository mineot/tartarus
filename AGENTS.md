# AGENTS.md

State of the Tartarus project. Read this before touching anything.

## The project

A CLI to manage and run named sets of terminal commands, with manuals, import
and export in JSON.

- Module: `tartarus` (`go.mod`)
- `go.mod` declares Go 1.25.0; the installed toolchain is 1.26.0
- Only dependency: `github.com/mattn/go-sqlite3` (CGO)
- Current branch: `break/v2`

## Commands

```bash
go build ./...
go vet ./...
go test -race ./store/
gofmt -l .
```

The project has tests **only** in `store/`. `repositories/` and `backup/` have
none.

## Current state: nearly everything is commented out on purpose

This repository is being reworked. **The commented out code is not broken — it
was switched off on purpose** so the `store` could be rewritten in isolation. Do
not uncomment, fix or "repair" those files unless asked.

| File | State |
| --- | --- |
| `store/store.go` | Live. Rewritten (see below). |
| `store/tx.go` | Live. New. |
| `store/store_test.go` | Live. New. |
| `store/migration.go` | Fully commented out, only `package store`. |
| `repositories/commands.go` | Commented out. Types `Command`, `CommandItem` included. |
| `repositories/manuals.go` | Commented out. Type `Manual` included. |
| `backup/import.go` | Commented out. |
| `backup/export.go` | Commented out. |
| `backup/backup.go` | Live. Only the JSON structs for the backup format. |
| `backup/restore-legacy.go` | Live. Empty stub. |
| `helpers/helpers.go` | Live. Dev/prod paths. |
| `main.go` | Live. Only prints a banner. |

Consequence: `helpers` has no live consumer left. `GetProductionStorePath` and
`GetDevelopmentStorePath` will be used by whoever calls `store.New`, not by the
store itself.

## The store

`store.New(ctx, path)` owns the connection. There is no `Open()` anymore, no
`Begin`/`Commit`/`Rollback`, and no `tx` field on the struct.

```go
func New(ctx context.Context, path string) (*Store, error)
func (s *Store) Close() error                       // idempotent
func (s *Store) Exec(query string, args ...any) (sql.Result, error)
func (s *Store) Query(query string, args ...any) (*sql.Rows, error)
func (s *Store) WithTx(fn func(*Tx) error) error

func (t *Tx) Exec(query string, args ...any) (sql.Result, error)
func (t *Tx) Query(query string, args ...any) (*sql.Rows, error)
func (t *Tx) QueryRow(query string, args ...any) *sql.Row
```

### Transaction contract

A transaction is **callback scoped, never instance state**. That is why there is
no `tx` field on `Store`: there is no way for one operation's transaction to
leak into the next. Any number of operations inside the same callback commit or
roll back together:

```go
err := store.WithTx(func(tx *store.Tx) error {
	if _, err := tx.Exec(insertCommand, ...); err != nil {
		return err
	}
	return nil
})
```

`fn` returning `nil` commits; returning an error rolls back and returns the
error. A panic inside `fn` also rolls back, before it propagates.

### Concurrency contract

- A `Store` may be shared across goroutines, but **concurrency goes through
  `WithTx`**.
- One transaction at a time per `Store`. A nested or concurrent `WithTx` returns
  `ErrTxActive`.
- While a transaction is in progress, `Store.Exec` and `Store.Query` return
  `ErrUseTx` instead of waiting. Consequence: **a simple read in parallel with a
  write does not work**, it fails. This is deliberate — with a single connection
  in the pool, waiting there would serialize access without changing the
  outcome.
- Inside the `WithTx` callback, use only the `*Tx`. Calling `Store.Exec` there
  returns `ErrUseTx`.

### Connection configuration

`store.go` builds the DSN with `_busy_timeout=5000`, `_journal_mode=WAL` and
`_txlock=immediate`, and caps the pool at 1 connection. Before changing any of
this:

- **WAL is persistent.** The mode is written to the file header, so the root
  `dev.db` is already in WAL and stays that way for every future open.
- `_txlock=immediate` swaps `BEGIN` for `BEGIN IMMEDIATE`, which takes the write
  lock up front. Without it, a transaction that only reads can fail when it is
  promoted to a write, and the `_busy_timeout` does not cover that case.
- The single-connection pool is what makes `SQLITE_BUSY` impossible. It is also
  what makes a leaked connection lock up the whole `Store`, which is why the
  rollback on panic is not optional.

### `Store.QueryRow` does not exist, on purpose

`*sql.Row` has no way to carry an error, so `ErrClosed` and `ErrUseTx` would be
swallowed and surface in the `Scan` with the wrong message. If you need it
outside a transaction, use the signature `QueryRow(...) (*sql.Row, error)`. On
`Tx` it is safe and it exists.

## Database schema

Defined in `store/migrations/0001_create_version_one.up.sql`: `commands`,
`command_items`, `manuals`, `migrations`.

## Known outstanding work

In order of urgency. None of it has been dealt with yet.

1. **`PRAGMA foreign_keys = ON` in migration 0001 is a no-op.** SQLite ignores
   that pragma inside a transaction, and `up()` runs everything inside one. So
   the referential integrity of `command_items.command_id REFERENCES commands(id)
   ON DELETE CASCADE` **has never actually been applied**. `_foreign_keys=on` in
   the DSN would fix it, but it turns on enforcement that is currently off and
   may expose bugs in the code that has not been rewritten yet. Handle it
   together with re-enabling `migration.go`.
2. **`store/migration.go` is commented out and incompatible with the new API.**
   The direct `s.db` accesses (lines 53, 88, 132, 140 at `HEAD`) have to become
   calls to a wrapper or to the `*Tx`. `ResetMigrations` (166-192) has to become
   a `WithTx` — it still calls `s.Open()`, `s.Begin()`, `s.Commit()` and
   `s.Rollback()`, none of which exist anymore.
3. **`backup/import.go` has a double `Open` bug.** It called `st.Open()` and then
   `st.ResetMigrations()`, which opened the store again. The old API had no
   guard, so this leaked a connection. It will disappear in the rewrite, but
   worth confirming.
4. **`repositories/` and `backup/` have to be converted to `WithTx`.** The old
   pattern was `st := store.Store{}` + `Open()` + `defer Close()` per operation,
   with a manual transaction. That becomes `store.WithTx(func(tx *store.Tx)
   error { ... })`. Since `Store` is now a shared resource, decide who owns it: one
   `Store` per operation still works, or one `Store` per process.
5. **`main.go` does nothing** beyond the banner.

## Style

- Comments and error messages are in English.
- SQL always uses `?` placeholders, never string interpolation.
- Errors are always wrapped or returned; no `panic` on a normal path.
