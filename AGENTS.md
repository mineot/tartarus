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
make build    # go build with -ldflags, injects the git version
make run      # go run . with no ldflags, so it is always development
make test     # go test -race ./...
make vet
make fmt      # gofmt -l .
```

Or the raw commands, which is what the Makefile runs:

```bash
go build ./...
go vet ./...
go test -race ./...
gofmt -l .
```

Tests exist in `store/`, `helpers/`, `repositories/` and `backup/`.

## Current state: nearly everything is commented out on purpose

This repository is being reworked. **The commented out code is not broken — it
was switched off on purpose** so the `store` could be rewritten in isolation. Do
not uncomment, fix or "repair" those files unless asked.

| File | State |
| --- | --- |
| `store/store.go` | Live. Rewritten (see below). |
| `store/tx.go` | Live. New. |
| `store/store_test.go` | Live. New. |
| `store/migration.go` | Live. Rewritten for `WithTx` (see below). |
| `store/migration_test.go` | Live. New. |
| `repositories/commands.go` | Live. Full CRUD for commands and command items converted. |
| `repositories/manuals.go` | Live. Full CRUD for manuals converted. |
| `repositories/manuals_test.go` | Live. New. |
| `repositories/commands_test.go` | Live. New. |
| `backup/import.go` | Commented out. |
| `backup/export.go` | Live. Rewritten (see below). |
| `backup/export_test.go` | Live. New. |
| `backup/backup.go` | Live. Only the JSON structs for the backup format. |
| `backup/restore-legacy.go` | Live. Empty stub. |
| `helpers/helpers.go` | Live. `GetStorePath` and `SetDevStorePath`. |
| `helpers/helpers_test.go` | Live. New. |
| `main.go` | Live. Opens the store, migrates, prints the path, lists manuals. |

`store.New` is the only consumer of `helpers.GetStorePath`. The prod/dev policy
stays in `helpers`; `store` only asks for the path, and `Store.Path` reports
which file was opened.

## Build mode and paths

`helpers` owns the only policy in the project that answers "which database?".
`store.New` calls it.

```go
func helpers.GetStorePath() (string, error) // ~/.tartarus/tartarus.db, or <project root>/dev.db
```

The mode comes from `helpers.version`, injected at link time by the Makefile:

```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
go build -ldflags "-X tartarus/helpers.version=$(VERSION)"
```

- `make build` (or any `go build` with `-ldflags`) injects a version → **production**.
- `go run .`, and a bare `go build .` with no `-ldflags`, leave `version` at its
  declared default `"dev"` → **development**.

There is **no automatic** way for a binary to tell how it was built: with no
flags, `go build` and `go run` produce identical build info. The Makefile is the
pre-build step. This replaced an older `IsProduction()` that sniffed the
executable path for the substring `"go-build"` — that only worked for `go run`
and sent a locally built binary to `~/.tartarus`, since `.gitignore` has always
expected the binary at the project root.

Two things to know before touching this:

- `-X` only rewrites a `string` variable whose initializer is a constant, and
  the import path in the flag has to be the full one (`tartarus/helpers`).
- There is no build-step-free fallback. `runtime/debug.ReadBuildInfo()` with
  `Main.Version == "(devel)"` can tell a local build from
  `go install tartarus@v1.0`, but it cannot separate `go build` from `go run`, so
  it is not a substitute. It would only work as a safety net so a mis-built
  binary does not mistake itself for a release.

## The store

`store.New(ctx)` owns the connection. There is no `Open()` anymore, no
`Begin`/`Commit`/`Rollback`, and no `tx` field on the struct. The path is not a
constructor argument: `New` asks `helpers.GetStorePath` for it, and `newAt` (the
unexported path-taking constructor) is what the tests use.

```go
func New(ctx context.Context) (*Store, error)
func (s *Store) Path() string                           // which file was opened
func (s *Store) Close() error                           // idempotent
func (s *Store) Exec(query string, args ...any) (sql.Result, error)
func (s *Store) Query(query string, args ...any) (*sql.Rows, error)
func (s *Store) WithTx(fn func(*Tx) error) error

func (t *Tx) Exec(query string, args ...any) (sql.Result, error)
func (t *Tx) Query(query string, args ...any) (*sql.Rows, error)
func (t *Tx) QueryRow(query string, args ...any) *sql.Row
```

`New(ctx)` itself is not covered by tests: calling it touches the real database,
and `helpers.version` has no exported override, so the suite exercises `newAt`
instead.

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

Every error that leaves package `store` carries the `store: ` prefix, either from
a sentinel or from a `fmt.Errorf("store: <what>: %w", err)` wrapper. `Store.Exec`,
`Store.Query`, `WithTx`'s `BeginTx` and `Commit` all wrap; the `Tx` methods do
**not**, deliberately, since a caller inside a callback already has the
transaction scope in hand and would otherwise get `repositories: inserting
command: store: executing statement: ...` on every database error. The
asymmetry is deliberate, not an oversight — if you change one side, change the
other and update this paragraph.

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

`store.go` builds the DSN with `_busy_timeout=5000`, `_journal_mode=WAL`,
`_txlock=immediate` and `_foreign_keys=on`, and caps the pool at 1 connection.
Before changing any of this:

- **WAL is persistent.** The mode is written to the file header, so the first
  time the root `dev.db` is opened it turns WAL and stays that way for every
  future open. It has already happened: header bytes 18/19 are `02 02`. Expect
  `-wal` and `-shm` sidecar files next to it while a connection is open; they
  are in `.gitignore` and disappear on clean close.
- `_txlock=immediate` swaps `BEGIN` for `BEGIN IMMEDIATE`, which takes the write
  lock up front. Without it, a transaction that only reads can fail when it is
  promoted to a write, and the `_busy_timeout` does not cover that case.
- `_foreign_keys=on` is what makes
  `command_items.command_id REFERENCES commands(id) ON DELETE CASCADE` real. It
  has to come from the DSN: SQLite ignores `PRAGMA foreign_keys` inside a
  transaction, and migrations run inside one, so the pragma that used to sit on
  line 1 of `0001_create_version_one.up.sql` was always a no-op. Enforcement is
  on now, which is a behaviour change for any code that was relying on writes
  being unchecked.
- The single-connection pool is what makes `SQLITE_BUSY` impossible. It is also
  what makes a leaked connection lock up the whole `Store`, which is why the
  rollback on panic is not optional. It is also why the single connection means
  the DSN pragmas cannot drift between operations.

### `Store.QueryRow` does not exist, on purpose

`*sql.Row` has no way to carry an error, so `ErrClosed` and `ErrUseTx` would be
swallowed and surface in the `Scan` with the wrong message. If you need it
outside a transaction, use the signature `QueryRow(...) (*sql.Row, error)`. On
`Tx` it is safe and it exists.

## Migrations

`store/migration.go` is live. `currentVersion` is a compiled-in `int` (still 1),
and `queries` maps a version to its `up`/`down` SQL, embedded from
`store/migrations/`.

```go
func (s *Store) RunMigrations() error    // one WithTx
func (s *Store) ResetMigrations() error  // one WithTx: drop + rebuild together
```

Both wrap an unexported `runMigrations(tx *Tx) error`, which assumes it is
already inside a transaction. That indirection is not optional: a shared body
that opened its own transaction would hit `ErrTxActive`, coming from the
transaction its caller had already started.

- `RunMigrations` is idempotent and safe to call on every start. `main.go` is the
  only caller; `store.New` deliberately does **not** migrate, so opening a
  connection never writes schema as a side effect.
- `ResetMigrations` drops and rebuilds in **one** transaction. The old version
  committed the drop first and migrated afterwards, which left a window where the
  database had no schema, and left it that way for good if the migration failed.
- The rollback path of `ResetMigrations` has **no test**. The guarantee rests on
  it being a single transaction, not on an assertion. If that ever stops being
  true, the test is missing.
- Column names in the `migrations` table are `version_date` and `version_number`.
  The `Migration` struct fields stay camelCase because they are private Go
  fields, not columns.
- `down` only runs when the database is ahead of the binary, which means an older
  build was opened against a newer schema.

### Editing version 1 while `currentVersion` is still 1

`description` was added by editing `0001_create_version_one.up.sql` directly,
because there is no version 2 yet and the schema is still moving. That is fine
for a fresh database — every test creates one in `t.TempDir()` — and it is **not**
fine for an existing one: `RunMigrations` is idempotent, so a database that
already recorded version 1 will never see the new column. The root `dev.db` is in
exactly that state.

The repair is `ResetMigrations`, which drops and rebuilds in one transaction. It
destroys data, so it is a deliberate act, not something to wire into startup.

## Repositories

`repositories` sits on top of `store`. The old pattern was `store.Store{}` +
`Open()` + `defer Close()` **per operation**, which is gone: the `Store` is opened
once per process in `main.go` and passed in.

```go
// manuals
func (r *Repos) GetManuals() ([]Manual, error)
func (r *Repos) GetManual(id uint64) (Manual, error)
func (r *Repos) InsertManual(m *Manual) error
func (r *Repos) UpdateManual(m *Manual) error
func (r *Repos) DeleteManual(id uint64) error

// commands
func (r *Repos) GetCommands() ([]Command, error)
func (r *Repos) GetCommand(id uint64) (Command, error)
func (r *Repos) InsertCommand(c *Command) error
func (r *Repos) UpdateCommand(c *Command) error
func (r *Repos) DeleteCommand(id uint64) error

// command items
func (r *Repos) GetCommandItems(commandID uint64) ([]CommandItem, error)
func (r *Repos) GetCommandItem(commandID, itemID uint64) (CommandItem, error)
func (r *Repos) AppendCommandItem(commandID uint64, item *CommandItem) error
func (r *Repos) UpdateCommandItem(commandID uint64, item *CommandItem) error
func (r *Repos) RemoveCommandItem(commandID, itemID uint64) error
```

- **Reads go through `Store.Query`, not `WithTx`.** A read needs no transaction,
  and `_txlock=immediate` would make `WithTx` take the write lock regardless,
  serializing readers against writers for nothing.
- Consequence, and it is a real one: a read cannot run while a `WithTx` is in
  progress on the same `Store`. `Store.Query` returns `ErrUseTx`. Read after the
  write commits.
- **Writes go through `Store.WithTx`.** All of them are converted.
- A read that has to happen **inside** a write must use `tx.QueryRow`, not
  `s.Query`, for the reason above. Every `Insert*`, `Update*` and `Append*` reads
  its row back that way, inside the same transaction that wrote it, and copies
  the result onto the caller's struct.
- Errors are wrapped as `repositories: <what>: %w`, which keeps the `store`
  sentinels matchable through `errors.Is`.

### Nullable columns: `description`

`commands`, `command_items` and `manuals` each carry a `description TEXT`
column with no `NOT NULL`. The Go field is a plain `string`, matching `Name`,
`Script` and `Body`, and every SELECT wraps the column in
`COALESCE(description, '')`. That wrapper is not cosmetic: `Scan` into a `string`
fails on NULL with `converting NULL to string is unsupported`.

The consequence is that **the repository never writes NULL**. An empty
description reaches the database as `''`, and a NULL row reads back as `""`. If
NULL and empty ever need to be told apart, drop the `COALESCE` and use
`sql.NullString` on the struct — do not add a second code path.

Tests reach NULL only by writing the column directly, through the `nullable`
helper in `commands_test.go`, which `manuals_test.go` also uses. That helper is
the reason NULL is covered at all; the repository's own write path cannot produce
it.

## Export

`backup/export.go` is converted. The old version called package-level
`repositories.GetCommands()`, which opened a connection per call; there is no
package-level anything in `repositories` anymore, so the store is threaded in:

```go
func Export(r *repositories.Repos, path string) error
```

Three things about it are deliberate:

- **It takes `*repositories.Repos`, not `*store.Store`.** `backup` reads through
  the repository layer and never touches `store` directly, which is what keeps the
  layering one-way. The cost is that a nil `Repos`, or a `Repos` with a nil
  `Store`, panics on the first call — item 8 in the outstanding work below, still
  open.
- **It never starts a transaction.** Reads go through `Store.Query`, so it must
  not be called from inside a `store.WithTx` callback: those reads return
  `store.ErrUseTx`. `export_test.go` asserts exactly that, and asserts that a
  closed store propagates `store.ErrClosed`. Both sentinels survive the `backup:`
  wrap through `errors.Is`, which is why `Export` wraps rather than returning
  bare.
- **It runs one query per command for that command's items**, an N+1 over a pool of
  one connection. Sequential reads are fine — `GetCommands` drains and closes its
  rows before returning, so the connection is free for the next read — and a single
  `LEFT JOIN` would cut the query count if it ever matters. Ordering is free too:
  both queries are `ORDER BY id`.

`backup/backup.go` holds the format, and all three structs now carry
`Description` with a plain `json:"description"`, no `omitempty`. An empty
description is written as `""` rather than dropped. That matches what the format
does with every other empty field, and `json.Unmarshal` tolerates the field being
absent, so files written before the field existed still load.

`backup/import.go` is not converted yet. When it is, `Description` has to be
carried through on all three structs. `Export` has no caller until `main.go`
grows a CLI surface (item 5).

## Testing outside `store`

`store`'s tests use the unexported `newAt`, so they can point a `Store` at
`t.TempDir()`. From another package that door does not exist, which is why
`helpers` has a test-only override:

```go
func helpers.SetDevStorePath(path string) error // ErrProduction if a version is injected
```

It short-circuits the project-root search in `developmentPath`. It is refused in
any build with a version injected, so a released binary cannot be redirected to a
different file, and there is a test asserting exactly that. There is no locking on
the variable, so tests using it must not run in parallel.

### Test files carry no comments

Every `_test.go` file is comment-free, and that is deliberate. The rationale for
the test helpers lives in this file instead, in the sections above:

| Helper | Why it exists |
| --- | --- |
| `repositories.newTestStore` | the only door into a database from outside `store`, and why it must not run in parallel |
| `backup.newTestStore`, `exportPath`, `readExport` | a third copy of the same door, for the same reason, plus a scratch path and a reparse of what `Export` wrote |
| `repositories.nullable` | the only way to write SQL NULL, which the repository's own write path cannot produce |
| `repositories.insertTestCommand`, `insertTestCommandItem`, `insertManual` | set up rows the repository API would not allow |
| `store.newEmptyStore` | `store.newTestStore` creates an `items` table, which gets in the way of asserting on what `RunMigrations` builds |
| `helpers.setVersion`, `markRoot` | swap the link-time version and plant a `go.mod` marker |

If a helper needs a "why" that is not in this file, add it here. Do not put it
back in the test.

Failure messages in tests are English, matching the style rule below. The
**fixture data** inserted by tests is still Portuguese (`"compila o projeto"`,
`"descricao nova"`), which is inconsistent and can be translated whenever it
bothers someone.

## Database schema

Defined in `store/migrations/0001_create_version_one.up.sql`: `commands`,
`command_items`, `manuals`, `migrations`. A freshly created database has none of
them until `RunMigrations` has run — `store.New` only creates the file.

## Known outstanding work

In order of urgency. None of it has been dealt with yet.

1. **`backup/import.go` has a double `Open` bug.** It called `st.Open()` and then
   `st.ResetMigrations()`, which opened the store again. The old API had no
   guard, so this leaked a connection. The second `Open` is gone from
   `ResetMigrations`; confirm it is gone from the import rewrite too.
2. **`backup/import.go` is not converted.** `backup/export.go` is (see below).
   `import.go` still calls `Open()`, `Begin()`, `Commit()` and `Rollback()`.
3. **`Command.Delete` deletes `command_items` explicitly even though the FK now
   cascades.** It is not wrong, just redundant. Worth deciding whether to keep the
   belt and braces or trust the constraint.
4. **The root `dev.db` predates the `description` column.** It recorded version 1
   before that column was added, so `RunMigrations` will not give it the column.
   `ResetMigrations` fixes it and destroys the data. Not urgent: the file is
   gitignored and `dev.db` is a scratch database.
5. **`main.go` still has no CLI surface.** It opens, migrates, prints the path and
   lists manuals as a usage example.
6. **`s.Close()` from inside a `WithTx` callback deadlocks.** `WithTx` holds
   `s.mu` for the whole callback and `Close` takes `s.mu`, and `sync.Mutex` is not
   reentrant. There is no guard and the doc comment does not warn about it. This
   is the most dangerous of the list, because the same rollback-on-panic rule that
   keeps the store alive after an aborted callback is what makes `mu` span the
   callback in the first place. Either document it or return `ErrTxActive` from
   `Close` when `inTx` is set.
7. **`command_items.script` is `UNIQUE` globally, not per command.** Two different
   commands cannot hold the same script. Almost certainly it should be
   `UNIQUE(command_id, script)`. This is an edit to migration 1, so it inherits
   the caveat in the next item: it will not reach an existing database.
8. **`Repositories` has no interface and no nil guard.** `Repos.Store` is an
   exported `*store.Store`, so `repositories.New(nil)` compiles and panics on the
   first call. Rewriting `backup/` against a small interface is the natural moment
   to fix both halves at once.
9. **Coverage gaps, in rough order of value.** `down()` has never been executed,
   since it needs a database ahead of the binary, and neither have its
   `missing down migration` branch or `up()`'s `missing up migration` branch. Also
   untested: `newAt` with an empty path, `Path()` after `Close` (documented as
   still correct, but nothing holds it there), and a `WithTx` from a second
   goroutine, which currently only exercises the `inTx` fast path and not the
   `mu` window behind it.
10. **Naming and doc gaps in `repositories`.** `manuals.go` uses
    `selectManualQuery` where `commands.go` uses `selectCommand`; `Command`,
    `CommandItem` and all ten error sentinels have no doc comment, while every
    exported name in `store` has one. `GetCommand`/`GetManual` do not reject
    `id == 0` although every `Update*`/`Delete*` does, and `GetCommands` does not
    document the nil slice that `GetManuals` documents.

## Style

- Comments and error messages are in English, in tests too.
- SQL always uses `?` placeholders, never string interpolation.
- Errors are always wrapped or returned; no `panic` on a normal path.
