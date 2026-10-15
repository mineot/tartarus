# AGENTS.md

State of the Tartarus project. Read this before touching anything.

## Terminal table component

`components/table.go` is the self-contained terminal table component. Keep its
options, validation, private model, sizing, keyboard handling, rendering and
Bubble Tea program lifecycle in that file.

```go
func Table(opts TableOpts) error // package components
```

`TableOpts` contains `Title`, `Headers`, `Rows`, `EmptyMsg`, `Widths` and
`Selectable`. `Table` validates the options before opening the terminal, creates
its private `tableModel` through `newTable`, and runs Bubble Tea inline in the normal
terminal screen, without alternate-screen/fullscreen mode. It returns validation or program errors to the caller without printing
them. A Cobra command can return `components.Table(opts)` from `RunE`; the caller
loads repository data and converts it to `[][]string` before invoking the table.
There is no public `NewTable`, `TableModel`, underlying-model accessor or
`SetRows` API.

### Data and validation

- At least one header is required. Every row must have exactly one cell per
  header; malformed rows are rejected before Bubbles can render them.
- `Widths` may be omitted. Otherwise it must contain one positive integer per
  header. Widths describe content space, excluding cell padding.
- Headers are copied into column structs; each row and the widths slice are
  copied, so later changes to the caller's slices do not change the model.
- An empty data set shows `EmptyMsg`, defaulting to `No rows found.`. The title
  is optional and uses the primary theme style; empty text uses the tertiary
  style. Header borders use `themes.PrimaryColor`; the selected row uses
  `themes.AssetColor`.
- The table receives an in-memory data set. It performs no database queries,
  database pagination, filtering, retries, editing or asynchronous loading.

### Terminal size and interaction

- `Init` starts no command. Rendering waits for `tea.WindowSizeMsg`.
- Resizing deducts the title, header, footer and padding from the available space.
  Natural column widths are measured from all headers and rows using terminal
  cell widths. Explicit widths cap those measurements. Columns shrink when
  needed to fit the terminal and recover their measured widths when it grows;
  spare terminal width is not used to stretch the table.
- At most ten data rows are visible, fewer for smaller data sets or terminals.
  Additional rows are reached by keyboard scrolling. The current cursor
  is retained across resizing. Rendering is clipped to the terminal dimensions.
- If a nonempty table cannot fit all columns, its header and at least one row,
  it displays `Terminal too small. Resize or press q to exit.` instead.
- Bubbles always handles row navigation and scrolling (including up/down and
  Page Up/Page Down). The selected row is always highlighted with reverse
  video, bold text and `themes.AssetColor`, including when options are
  omitted. `Selectable` is deprecated and retained only for source compatibility;
  its value no longer disables the highlight.
- Mouse reporting is enabled with `tea.WithMouseCellMotion`. Wheel up/down
  moves the current selection by one row and scrolls it into view, sharing the
  keyboard cursor. Click selection is not implemented: inline mouse coordinates
  are relative to the terminal, not to the table origin. Selection is visual only: Enter does not return or execute a row.
- The footer shows `↑/k/PgUp up • ↓/j/PgDn down • q quit`, wrapping on narrow
  terminals. Arrows and `j`/`k` move one row; PgUp/PgDn move one visible page.
  Its height and separator are reserved before calculating visible rows.
- `q`, Escape and Ctrl+C quit regardless of focus, empty state or terminal size.
  No root Bubble Tea model or page-shortcut dispatcher is required.

### Cobra integration and historical UI

`main.go` opens and migrates the store, creates the repository, renders the logo
and calls `views.New(r).ExecuteContext(ctx)`. The store is closed before `main`
reports an execution error. The current entry point has no legacy restore flag.

The command views are being reworked: `tartarus cmds ls` currently calls
`Table` with sample rows. Repository loading and the other command handlers
remain unfinished.

The former `CommandsList`, `MenuCmds`, five-row database pagination, `/` search,
retry handling and page-shortcut editing rules do not describe this component.
`components/table_test.go` currently contains commented-out tests for the old
command list; it provides no active regression tests for `Table`. Do not restore
that historical implementation as part of table maintenance.

## Terminal prompt component

`components/prompt.go` is the self-contained prompt component. Keep its options,
line reading and rendering in that file. It uses only the standard library: no
Huh, no Bubble Tea, no Lip Gloss and no raw mode.

```go
func Prompt(opts PromptOpts) (string, error) // package components
func Confirm(opts ConfirmOpts) (bool, error) // package components
```

`PromptOpts` carries `Label`, `Default` and `Validate`. `ConfirmOpts` carries
`Question` and `Default`. Both read from `os.Stdin` and write to `os.Stderr`,
keeping stdout free for the command's output. There is no multi-field model: the
caller sequences calls to build a multi-step flow.

### Prompt

- `Prompt` writes `Label: `, reads one line and returns the value trimmed of
  surrounding whitespace. An empty line returns `Default`.
- A nonempty `Default` is shown between brackets: `Label [default]: `. There is
  no editable inline buffer, which would need raw mode; the user types the new
  value, or submits the empty line to keep the current one.
- A line without a trailing newline (the last line of a pipe) is returned as-is.
- `Validate` runs on the value about to be returned, after `Default` has been
  substituted. On error the prompt does not advance: the error is printed in red
  (`\x1b[31m`) and the rejected value becomes the new default, so it is shown in
  brackets and can be adjusted.
- Ctrl+C is not intercepted; it behaves like any other CLI and the process
  receives SIGINT.

### Confirm

- `Confirm` writes `Question (s/N): ` or `Question (S/n): `, depending on
  `Default`, and reads one line.
- It accepts `s`, `sim`, `y` and `yes` for true and `n`, `nao` and `no` for
  false, case-insensitively. An empty line returns `Default`.
- Any other answer prints a red hint (`responda s ou n`) and asks again.

### Errors and tests

- Every error is wrapped as `components: prompt: %w`.
- A closed input with no data returns `io.ErrUnexpectedEOF`, matchable through
  `errors.Is`. There is no context parameter and no `ErrAborted` sentinel.
- `components/prompt_test.go` is active. It drives the unexported `prompt` and
  `confirm`, which take an `io.Reader` and an `io.Writer`, so no TTY is needed.
  It covers valid, trimmed, newline-less, defaulted and rejected values, the red
  reason, the rejected value kept as default, the confirm defaults and answers,
  write and read errors, and closed input.

## The project

A CLI to manage and run named sets of terminal commands, with manuals, import
and export in JSON.

- Module: `tartarus` (`go.mod`)
- `go.mod` declares Go 1.26.0
- SQLite uses `github.com/mattn/go-sqlite3` (CGO). The CLI uses Cobra; the
  table and the theme package use Lip Gloss v2. The prompt component uses only
  the standard library.
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

## Current state: rework in progress

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
| `repositories/commands.go` | Live. CRUD for commands and `CommandScript` rows in `command_scripts`. |
| `repositories/manuals.go` | Live. Full CRUD for manuals converted. |
| `repositories/manuals_test.go` | Tests currently commented out; historical assertions are not active coverage. |
| `repositories/commands_test.go` | Tests currently commented out; historical assertions are not active coverage. |
| `backup/import.go` | Live. Rewritten (see below). |
| `backup/export.go` | Live. Rewritten (see below). |
| `backup/export_test.go` | Live. New. |
| `backup/import_test.go` | Tests currently commented out; historical assertions are not active coverage. |
| `backup/backup.go` | Live. Only the JSON structs for the backup format. |
| `backup/restore-legacy.go` | Live. New. Reads a legacy v1.x export (see below). |
| `backup/restore-legacy_test.go` | Tests currently commented out; historical assertions are not active coverage. |
| `backup/testdata/legacy.json` | Live. A real v1.x export, used as the fixture. |
| `helpers/helpers.go` | Live. `GetStorePath` and `SetDevStorePath`. |
| `helpers/helpers_test.go` | Live. New. |
| `components/table.go` | Live. Standalone interactive table; see the contract above. |
| `components/table_test.go` | Commented-out historical command-list tests. |
| `components/prompt.go` | Live. Standalone prompt component; see the contract above. |
| `components/prompt_test.go` | Live. New. `prompt`/`confirm` via injected reader and writer. |
| `views/commands.go` | In-progress Cobra command handlers; subcommand registration is commented out. |
| `main.go` | Live. Opens and migrates the store, renders the logo and executes Cobra. |

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
  `command_scripts.command_id REFERENCES commands(id) ON DELETE CASCADE` real. It
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

`description` was added and `command_items` was renamed to `command_scripts`
by editing migration 1 while `currentVersion` remains 1. Fresh databases receive
the current schema. `RunMigrations` does not replay version 1 for an existing
version-1 database, so it neither adds missing columns nor renames old tables.
Do not assume the schema of an existing `dev.db` from the migration source alone.

`ResetMigrations` is destructive and must not be wired into startup as an
implicit repair. Its current drop script only names `command_scripts`,
`commands`, `manuals` and `migrations`: it does not drop an old `command_items`
table. Resetting a database with that old table is not a complete rename or
cleanup strategy. Preserving existing data requires an explicit migration;
recreating a disposable development database is a separate deliberate action.

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
func (r *Repos) GetCommands(searchTerm string) ([]Command, error)
func (r *Repos) GetCommand(id uint64) (Command, error)
func (r *Repos) InsertCommand(c *Command) error
func (r *Repos) UpdateCommand(c *Command) error
func (r *Repos) DeleteCommand(id uint64) error

// command scripts
func (r *Repos) GetCommandScripts(commandID uint64) ([]CommandScript, error)
func (r *Repos) GetCommandScript(commandID, scriptID uint64) (CommandScript, error)
func (r *Repos) AppendCommandScript(commandID uint64, script *CommandScript) error
func (r *Repos) UpdateCommandScript(commandID uint64, script *CommandScript) error
func (r *Repos) RemoveCommandScript(commandID, scriptID uint64) error
```

`CommandScript` is the repository type for `command_scripts`; the former
`CommandItem` methods are not the current API. `GetCommands(searchTerm)` returns
all matching commands ordered by id, without a limit, offset or filtered count.
A nonempty search is wrapped in `%` and applied with SQLite `LIKE` to name and
`COALESCE(description, '')`. `%` and `_` are not escaped, so they remain wildcard
characters. Backup readers use `GetCommands("")`. The pagination/literal-search
comment above the implementation is stale.

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

`commands`, `command_scripts` and `manuals` each carry a `description TEXT`
column with no `NOT NULL`. The Go field is a plain `string`, matching `Name`,
`Script` and `Body`, and every SELECT wraps the column in
`COALESCE(description, '')`. That wrapper is not cosmetic: `Scan` into a `string`
fails on NULL with `converting NULL to string is unsupported`.

The consequence is that **the repository never writes NULL**. An empty
description reaches the database as `''`, and a NULL row reads back as `""`. If
NULL and empty ever need to be told apart, drop the `COALESCE` and use
`sql.NullString` on the struct — do not add a second code path.

The historical tests write NULL directly through the `nullable` helper in
`commands_test.go`, also used by `manuals_test.go`. Both files are currently
commented out, so those cases are not active coverage. The repository write
path itself cannot produce NULL descriptions.

## Export

`backup/export.go` is converted. The old version called package-level
`repositories.GetCommands()`, which opened a connection per call; there is no
package-level anything in `repositories` anymore, so the store is threaded in:

```go
func Export(r *repositories.Repos, path string) error
```

Three things about it are deliberate:

- **It takes `*repositories.Repos`, not `*store.Store`.** `backup` reads rows through
  the repository layer and does not reach for `store` except to migrate, which is
  what keeps the layering one-way. The cost is that a nil `Repos`, or a `Repos`
  with a nil `Store`, panics on the first call — item 6 in the outstanding work
  below, still open.
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

`Export` is not currently wired into a Cobra command (item 3).

## Import

`backup/import.go` is converted, and takes the same `*repositories.Repos` as
`Export`. The old version called `store.Store.Open()`, `Begin()`, `Commit()` and
`Rollback()`, and called `Open()` **twice** — once itself and once inside
`ResetMigrations` — which leaked a connection. That is gone by construction now:
there is one `Store`, owned by the caller, and `ResetMigrations` is a method on it
that opens no second connection.

Four things about it are deliberate:

- **It destroys the database.** `ResetMigrations` runs before the first insert, so
  an import always replaces and never merges. Importing on top of existing rows
  would collide with the `UNIQUE` on `commands.name`, `manuals.name` and
  `command_scripts.script` anyway.
- **The reset is also what makes a failed import recoverable.** The inserts are
  not atomic — every `repositories` write opens its own transaction, and there is
  no `InsertCommandTx` to compose them into one — so a bad row halfway leaves a
  partial database. Re-running is still safe, because the reset at the top clears
  that partial state before the constraints can collide with it. A single
  transaction would be the better answer and needs new repository methods.
- **The file is validated before the reset, not after.** Every check in
  `validate` is one the database would reject on insert; running them first is the
  difference between a rejected file and a rejected file that also destroyed the
  previous data. The currently disabled `import_test.go` contains historical
  assertions that the pre-existing row survives seven rejection cases. The duplicate checks are there because the constraints
  are global: `command_scripts.script` is unique across the whole table, so two
  different commands cannot hold the same script, and the script set therefore
  cannot be reset per command.
- **Ids and timestamps in the file are ignored.** `Insert*` stamps
  `time.Now().UTC()` and lets SQLite assign ids from `AUTOINCREMENT`, so imported
  rows carry import time and freshly numbered ids. The old code set `CreatedAt`
  on the struct it passed in, which was already dead — the repository overwrites
  it from the row it reads back. The currently disabled round-trip test does not compare
  ids across the two databases; it compares names, bodies, scripts and
  descriptions, and separately asserts that each imported item's `command_id`
  matches the command it is nested under.

That last point is where the old code was actually **wrong**, not just wasteful:
it built each item with `CommandID: cmd.ID`, the id from the file, which after a
reset points at whatever row happens to sit at that id now — silently attaching
items to the wrong command. `Import` passes the `c.ID` that `InsertCommand` just
assigned. `AppendCommandScript` overwrites `item.CommandID` with the row it reads
back regardless, so passing the wrong id would have been caught, but only after a
failed insert, not before.

Like `Export`, `Import` must not be called from inside a `store.WithTx` callback,
but it fails with a **different sentinel**: `Export`'s reads return
`store.ErrUseTx`, while `Import` reaches `Store.WithTx` through `ResetMigrations`
and gets `store.ErrTxActive`. The export assertion is active; the import
assertion is in the currently disabled test file.

Neither backup operation is currently wired into a Cobra command (item 3).

## Legacy restore

`backup/restore-legacy.go` moves data out of the **v1.x** implementation into the
current one. It is the one entry point in the project that is not about a
backup file this codebase wrote.

### What the legacy store is

The v1.x release was a TypeScript CLI (`@mineot/tartarus`, last version 1.2.4)
holding its data in a **PouchDB** database, not SQLite. On disk that is a
`leveldown` store, one folder, `~/.tartarus/db/commands`, made of `*.ldb`
tables, a `MANIFEST-*`, a `CURRENT` and a `*.log`. It cannot be read with
`sqlite3`, and nothing here tries to.

PouchDB stores each document twice, in `levelup` sublevels whose keys are the
sublevel name wrapped in `0xff`:

| Sublevel | Key | Value |
| --- | --- | --- |
| `document-store` | `_id` | revision metadata, `rev_tree`/`rev_map`/`winningRev` |
| `by-sequence` | 16-digit zero-padded sequence number | the document JSON |
| `meta-store` | `_local_uuid`, `_local_doc_count`, `_local_last_update_seq` | counters |

So a document is a `document-store` entry pointing at a `by-sequence` entry, and
the `*.ldb` keys are prefix-compressed, which is why `strings` on the file shows
JSON fragments interrupted by varint bytes. None of that is parsed here.

### The user has to export first

**The v1 binary has to produce the export before the v2 binary is installed.**
That is a documentation instruction, not a code path: the two versions store
data in unrelated formats, and the only thing v2 can read is what v1 wrote out.

```bash
tartarus db export ~/.tartarus/legacy.json   # with the v1 binary still installed
```

The former v2 `-restore-legacy` flag is no longer exposed. The Go entry point
remains `backup.RestoreFromLegacy(r, path)`, without a current Cobra command.

`db export` is PouchDB's `allDocs({include_docs:true})` flattened to a JSON
array, so the file is the winning revision of every document, untyped:

```json
[
  {"instructions": ["rclone config reconnect gdrive-mineot:"], "_id": "cmd:gdrive-refresh-connection", "_rev": "1-8a67..."},
  {"instructions": ["sudo apt update"], "description": "...", "_id": "cmd:linux_upgrade", "_rev": "1-fb74..."},
  {"data": "nano", "updatedAt": "2025-08-16T02:29:38.705Z", "_id": "config:editor", "_rev": "1-c503..."},
  {"content": "teste\n\nteste\n\nteste\n", "updatedAt": "...", "_id": "manual:teste", "_rev": "1-d05c..."}
]
```

The format is a flat array with no type field, so **the prefix of `_id` is the
only discriminator**. `backup/testdata/legacy.json` is a real export, and it is
the fixture used by the currently disabled legacy restore tests.

`~/.tartarus/db/commands` is never read, written or deleted by the new code.
That is the reason the operation is safe to repeat: the user can re-run
`db export` and the restore as many times as they like.

### The mapping

| Legacy document | Becomes | Notes |
| --- | --- | --- |
| `cmd:<name>` | a `commands` row plus one `command_scripts` row per entry of `instructions` | `instructions` is the legacy name for what is now `command_scripts` |
| `cmd:<name>` with `description` | the row's `description` | absent means `""` |
| `manual:<name>` | a `manuals` row | `content` becomes `body`, verbatim |
| anything else, `config:editor` included | nothing | counted as skipped |

`legacyDoc` holds only `_id`, `instructions`, `description` and `content`. `_rev`
and `updatedAt` are absent on purpose: ids and timestamps are ignored the same
way `Import` ignores them, and the repositories stamp `time.Now().UTC()`
themselves. `data` held the editor name and has no destination in the schema.

### Signature, and why there is no marker file

```go
func RestoreFromLegacy(r *repositories.Repos, path string) error
```

Same shape as `Export` and `Import`: the store is owned by the caller and
threaded in, rather than opened here.

**There is no `.imported` marker, on purpose.** The stub asked for one, and it is
the wrong instrument: the legacy folder is never deleted, the user runs the
restore by hand, and the database's own `UNIQUE` constraints can stand in for a
marker. Three things fall out of that:

- **The run is a merge and converges.** It reads the whole existing state first
  (`GetCommands`, `GetManuals`, and one `GetCommandScripts` per command — the same
  N+1 `Export` runs) and only writes what is missing. A command that already
  exists keeps its row, its id and its description, and gains **only the
  instructions it lacks**. A second run is therefore a no-op, and so is a run
  that was interrupted halfway through a command: the second run finishes it.
- **A manual that already exists is left alone.** Deciding which of two bodies
  wins would mean comparing timestamps, which `Import` also ignores. Skipping is
  reported in the output.
- **An item whose script already belongs to another command is skipped and
  named**, not inserted and not fatal. This is the consequence of
  `command_scripts.script` being `UNIQUE` across the whole table (item 5 in the
  outstanding work): a script cannot be shared, so the second command to want it
  cannot have it.

The file is validated **before** the first write, same reasoning as `Import`:
every check is one the schema would refuse anyway, and the export belongs to the
user. `commands.name` and `manuals.name` are `UNIQUE`; `command_scripts.script` is
too, so a script repeated inside the file, or across two documents in it, is
rejected up front rather than halfway through. The currently disabled
`restore-legacy_test.go` contains historical assertions that the pre-existing
row survives eight rejection cases.

Nothing here is atomic: every `repositories` write opens its own transaction and
there is no `InsertCommandTx`. That was a real risk for `Import` and is
tolerable here **only** because of the convergence above — which is the same
reason `Import` can survive its own partial failure, by resetting first.

The report goes to **stdout** as one summary line plus one line per skipped
document or item. The signature returns only an error, and a restore the user
ran by hand deserves an answer in words; if that ever changes shape, return a
struct instead of printing.

### Dispatch and the two sentinels

Legacy restore is not dispatched by the current Cobra CLI. The earlier
`flag.Visit` and `log.Fatal` flow belonged to the removed startup flag.

Like `Export`, `RestoreFromLegacy` must not be called from inside a
`store.WithTx` callback, but it fails with the **read** sentinel:
`readLegacyState` runs first and goes through `Store.Query`, so the error is
`store.ErrUseTx`. `Import` gets `store.ErrTxActive` instead, from
`ResetMigrations`. The historical tests cover these two sentinels, but import
and legacy restore tests are currently disabled.

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

Active test code should contain no explanatory comments; helper rationale
belongs here. Several test files are currently disabled by commenting out their
contents. The helper names below include those historical definitions, not a
claim that they all run today:

| Helper | Why it exists |
| --- | --- |
| `repositories.newTestStore` | the only door into a database from outside `store`, and why it must not run in parallel |
| `backup.newTestStore`, `exportPath`, `readExport` | a third copy of the same door, for the same reason, plus a scratch path and a reparse of what `Export` wrote |
| `backup.legacyPath`, `legacyFixture` | point at `backup/testdata/legacy.json` and restore from it, so the fixture is a real v1.x export rather than an invented one |
| `backup.itemScripts`, `commandNames`, `manualNames` | read one shape back out of the repositories, so an assertion can name a command by its name instead of by id |
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

The application tables are defined in
`store/migrations/0001_create_version_one.up.sql`. The bookkeeping table
`migrations` is created separately by `ensureMigrationsTable` in
`store/migration.go`. `store.New` opens the database without creating this schema;
`RunMigrations` creates it.

| Table | Columns and constraints |
| --- | --- |
| `commands` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; unique, non-null `name`; nullable `description`; non-null `created_at` and `updated_at`. |
| `command_scripts` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; non-null `command_id`; globally unique, non-null `script`; nullable `description`; non-null `created_at` and `updated_at`. |
| `manuals` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; unique, non-null `name`; non-null `body`; nullable `description`; non-null `created_at` and `updated_at`. |
| `migrations` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; non-null `version_date` and `version_number`. |

`command_scripts.command_id` references `commands.id` with both
`ON DELETE CASCADE` and `ON UPDATE CASCADE`. Foreign keys are enabled through
the store DSN. `script` uniqueness is global, not scoped to `command_id`.
The repository additionally validates command names as 1–12 ASCII lowercase
letters/digits; that restriction is not a SQL `CHECK` constraint.

Both the version-1 down migration and `0000_drop_all_tables.sql` drop
`command_scripts` before `commands`. Only the reset script drops `migrations`.
Neither script includes the former `command_items` table; see the existing
version-1 database caveat above.

### Naming and backup compatibility

Use `command_scripts` in current SQL and `CommandScript` / `*CommandScript*`
methods in the repository API. `command_items` is the old SQL name.
The backup format still uses `items` (`json:"items"`) and the private
`commandItem` type in `backup/backup.go`. That serialized field has not been
renamed to `scripts`; do not change the backup format as a side effect of the
SQL/API rename. Legacy `instructions` entries map to `command_scripts` rows.

### Rename follow-up still outstanding

- Active `store/migration_test.go` assertions and SQL still reference
  `command_items`, so its schema and foreign-key tests need alignment with
  `command_scripts` before they can validate the new schema.
- The reset comment in `store/migration.go` and comments in `backup/backup.go`,
  `backup/import.go` and `backup/restore-legacy.go` still use the old SQL name.
- Commented-out repository and backup tests retain old APIs, including
  `CommandItem` methods and the paginated `GetCommands` signature. They need
  deliberate updates before reactivation, not a blanket uncomment operation.

## Known outstanding work

In order of urgency. None of it has been dealt with yet.

Items 1 and 2 used to live here: the double `Open` in `import.go`, and the
unconverted `import.go` itself. Both are resolved — see the `## Import` section.

1. **`Repos.DeleteCommand` deletes `command_scripts` explicitly even though the FK now
   cascades.** It is not wrong, just redundant. Worth deciding whether to keep the
   belt and braces or trust the constraint.
2. **Existing version-1 databases do not receive schema edits.** Missing
   descriptions and the old `command_items` name require an explicit migration
   strategy. The current reset script does not remove the old table; see the
   migration and schema sections above.
3. **Cobra table integration is unfinished.** Wire command handlers to
   `components.Table(opts)` and return its errors from `RunE`. Backup import,
   export and legacy restore also have no current CLI handlers.
4. **`s.Close()` from inside a `WithTx` callback deadlocks.** `WithTx` holds
   `s.mu` for the whole callback and `Close` takes `s.mu`, and `sync.Mutex` is not
   reentrant. There is no guard and the doc comment does not warn about it. This
   is the most dangerous of the list, because the same rollback-on-panic rule that
   keeps the store alive after an aborted callback is what makes `mu` span the
   callback in the first place. Either document it or return `ErrTxActive` from
   `Close` when `inTx` is set.
5. **`command_scripts.script` is `UNIQUE` globally, not per command.** Two different
   commands cannot hold the same script. Almost certainly it should be
   `UNIQUE(command_id, script)`. This is an edit to migration 1, so it inherits
   the caveat in item 2: it will not reach an existing database.
6. **`Repositories` has no interface and no nil guard.** `Repos.Store` is an
   exported `*store.Store`, so `repositories.New(nil)` compiles and panics on the
   first call. Rewriting `backup/` against a small interface is the natural moment
   to fix both halves at once.
7. **Coverage gaps, in rough order of value.** `down()` has never been executed,
   since it needs a database ahead of the binary, and neither have its
   `missing down migration` branch or `up()`'s `missing up migration` branch. Also
   untested: `newAt` with an empty path, `Path()` after `Close` (documented as
   still correct, but nothing holds it there), and a `WithTx` from a second
   goroutine, which currently only exercises the `inTx` fast path and not the
   `mu` window behind it.
8. **Naming and doc gaps in `repositories`.** `manuals.go` uses
    `selectManualQuery` where `commands.go` uses `selectCommand`; the command
    error sentinels lack doc comments, while `Command` and `CommandScript`
    already have them. `GetCommand`/`GetManual` do not reject
    `id == 0` although every `Update*`/`Delete*` does, and `GetCommands` does not
    document the nil slice that `GetManuals` documents.

## Style

- Comments and error messages are in English, in tests too.
- SQL always uses `?` placeholders, never string interpolation.
- Errors are always wrapped or returned; no `panic` on a normal path.
