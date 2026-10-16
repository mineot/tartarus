# AGENTS.md

State of the Tartarus project. Read this before touching anything.

## Terminal table component

`components/table.go` is the self-contained table renderer. Keep its fields and
rendering in that file. It is a one-shot `fmt`-style renderer built on
`text/tabwriter`, not a Bubble Tea program.

```go
type Table struct {
	Header   []string
	Rows     [][]string
	EmptyMsg string
	Title    string
	tw       *tabwriter.Writer
}

func (t *Table) Render()
```

The type and its four data fields are exported and set directly; `tw` is private
and created inside `Render`. There is no constructor, no options struct and no
error return. `Render` writes to `os.Stdout` and drops write errors.

### Rendering

- `Title` is printed first through Lip Gloss in `themes.PrimaryColor`, followed
  by a blank line.
- `Header` is optional. When present it is written first, then a separator row
  with one run of `-` per header, then the data rows. Each run is
  `strings.Repeat("-", len(h))` — the header's **byte** length, not its terminal
  cell width, so a multi-byte header under-lines.
- Cells are joined with `\t`; `text/tabwriter` (minwidth 1, tabwidth 1, padding
  2, space padchar) does the column alignment.
- An empty or nil `Rows` prints `EmptyMsg` and returns; the header and separator
  are not printed in that case. `EmptyMsg` has no default, so an unset one prints
  an empty line.
- The table is not interactive: there is no validation, no width measurement or
  truncation, no paging, no keyboard or mouse handling and no `q` to quit. The
  table receives an in-memory data set and performs no database queries.

### Historical UI

The former interactive table — `TableOpts`, `Widths`, `Selectable`, ten visible
rows, scrolling, the footer, `Terminal too small`, mouse reporting and the
Bubble Tea program lifecycle — does not describe this component.
`components/table_test.go` covers the renderer described above.

## Terminal prompt component

`components/prompt.go` is the self-contained prompt component. Keep its options,
line reading and rendering in that file. The prompt loop and line reading use
only the standard library; Lip Gloss is used only by `PromptTitle`.

```go
var ErrInterrupted = errors.New("interrupted")

func Prompt(opts PromptOpts) (string, error) // package components
func Confirm(opts ConfirmOpts) (bool, error) // package components
func PromptTitle(title string)               // package components
```

`PromptOpts` carries `Label`, `Default` and `Validate`. `ConfirmOpts` carries
`Question` and `YesDefault`. `Prompt` and `Confirm` read from `os.Stdin` and
write to `os.Stderr`, keeping stdout free for the command's output;
`PromptTitle` writes to `os.Stdout` in `themes.PrimaryColor`, followed by a blank
line. There is no multi-field model: the caller sequences calls to build a
multi-step flow.

### Prompt

- `Prompt` writes `Label: `, reads one line and returns the value trimmed of
  surrounding whitespace. An empty line returns the current default.
- A nonempty default is shown between brackets: `Label [default]: `. There is no
  editable inline buffer, which would need raw mode; the user types the new value
  or submits the empty line to keep the current one.
- A line without a trailing newline (the last line of a pipe) is returned as-is.
  Carriage returns are dropped.
- `Validate` runs on the value about to be returned, after the default has been
  substituted. On error the prompt does not advance: the error is printed in red
  (`\x1b[31m`) and the rejected value becomes the new default, so it is shown in
  brackets and can be adjusted.
- Ctrl+C is intercepted by the public `Prompt` and `Confirm` through
  `os/signal`, which covers only the duration of a prompt. On interrupt the
  prompt returns `ErrInterrupted`, wrapped as `components: prompt: %w` and
  matchable with `errors.Is`; `main.go` maps it to exit code 130 and prints a
  newline so the shell prompt starts on a fresh line. The internal `prompt` and
  `confirm` used by tests do not intercept signals. The terminal still echoes
  `^C` before the signal is delivered, which is normal and cannot be removed
  without raw mode. Outside a prompt the process keeps the default SIGINT
  behaviour. Because the prompt is read in a goroutine while the signal is
  watched, an interrupt abandons a goroutine blocked on stdin; that is safe
  only because a `ErrInterrupted` caller is expected to exit, so do not ignore
  the error and keep reading.

### Confirm

- `Confirm` writes `Question (y/N): ` or `Question (Y/n): `, depending on
  `YesDefault`, and reads one line.
- It accepts `y` and `yes` for true and `no` for false, case-insensitively. An
  empty line returns `YesDefault`.
- Any other answer — `n` included — prints the red hint `replay y ou n` and asks
  again. There is no `s`/`sim`/`nao` handling.

### Errors, line reading and tests

- Every error is wrapped as `components: prompt: %w`.
- `readLine` drops `\r`, terminates on `\n`, and returns the bytes read so far on
  EOF. A closed input with no data returns `io.ErrUnexpectedEOF`, matchable
  through `errors.Is`. There is no context parameter and no `ErrAborted`
  sentinel.
- `components/prompt_test.go` drives the unexported `prompt` and `confirm`,
  which take an `io.Reader` and an `io.Writer`, so no TTY is needed. It covers
  valid, trimmed, newline-less, defaulted and rejected values, the red reason,
  the rejected value kept as default, the confirm defaults and answers, write and
  read errors, and closed input; the public `Prompt`/`Confirm`/`PromptTitle`
  wrappers are covered by swapping `os.Stdin`/`os.Stderr`/`os.Stdout`.

## CLI and views

`main.go` opens and migrates the store (`store.New`, then `RunMigrations`),
creates the repository with `repositories.New`, renders the logo with
`themes.RenderLogo`, and calls `views.New(r).ExecuteContext(ctx)`. The store is
closed by `defer` before `main` reports an execution error. `main` also
recognizes `components.ErrInterrupted` and exits with code 130 after printing a
newline, instead of reporting it as a failure. The entry point has no legacy
restore flag.

`views/views.go` defines the root `tartarus` command (version `2.0.0`, usage and
errors silenced) and registers `views.NewCommands(r)` under `cmds`.
`views/commands.go` implements `cmds` with these handlers:

| Subcommand | Behaviour |
| --- | --- |
| `ls` | Reads `GetCommands(filter)` and renders a `Table` with `Name`/`Description`. `-f/--filter` is an optional literal substring; an empty or whitespace-only value is rejected in `PreRunE`. |
| `v <name>` | `GetCommand(name)`, renders a headerless `Table` (Name/Description), then `GetCommandScripts(command.ID)` in a `Table` with `Name`/`Script`/`Description`. |
| `c` | Prompts interactively for name and description with `components.Prompt`, validates the name with `ValidateCommandName`, and calls `InsertCommand`. |
| `e <name>` | `GetCommand(name)`, re-prompts with the current values as `Default`, and calls `UpdateCommand`. |
| `d <name>` | `GetCommand(name)`, asks for confirmation with `components.Confirm`, and calls `DeleteCommand`. |

The three item handlers `appendCommandItem`, `changeCommandItem` and
`removeCommandItem` (`ap`/`ch`/`rm`) remain commented out at the bottom of
`views/commands.go` and still reference the old `command_items` names. Do not
uncomment, fix or "repair" them unless asked; they are the only remaining
`CommandItem` spelling in the tree.

## The project

A CLI to manage and run named sets of terminal commands, with manuals, import
and export in JSON.

- Module: `tartarus` (`go.mod`)
- `go.mod` declares Go 1.26.0
- SQLite uses `github.com/mattn/go-sqlite3` (CGO). The CLI uses Cobra; the
  table, the theme package and `components.PromptTitle` use Lip Gloss v2. The
  prompt loop uses only the standard library.
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

Tests exist in `store/`, `helpers/`, `repositories/`, `backup/`, `components/`,
`themes/` and `views/`, plus `main_test.go`.

## Current state: rework in progress

This repository is being reworked. The former test files have been reactivated
and adjusted to the current APIs, and the store, repositories and backup layers
have been converted. The command views are implemented and active; only the
`ap`/`ch`/`rm` item handlers at the bottom of `views/commands.go` are commented
out on purpose, and still reference the old `command_items` API. Do not
uncomment, fix or "repair" them unless asked.

| File | State |
| --- | --- |
| `store/store.go` | Live. Rewritten (see below). |
| `store/tx.go` | Live. New. |
| `store/store_test.go` | Live. New. |
| `store/migration.go` | Live. Rewritten for `WithTx` (see below). |
| `store/migration_test.go` | Live. New; uses `command_scripts`. |
| `repositories/commands.go` | Live. CRUD for commands and `CommandScript` rows in `command_scripts`. |
| `repositories/manuals.go` | Live. Full CRUD for manuals converted. |
| `repositories/manuals_test.go` | Live. Reactivated and adjusted. |
| `repositories/commands_test.go` | Live. Reactivated and adjusted. |
| `repositories/command_scripts_test.go` | Live. New; `CommandScript` CRUD, search and validation. |
| `backup/import.go` | Live. Rewritten (see below). |
| `backup/export.go` | Live. Rewritten (see below). |
| `backup/export_test.go` | Live. New. |
| `backup/import_test.go` | Live. Reactivated and adjusted. |
| `backup/backup.go` | Live. Only the JSON structs for the backup format. |
| `backup/restore-legacy.go` | Live. New. Reads a legacy v1.x export (see below). |
| `backup/restore-legacy_test.go` | Live. Reactivated and adjusted. |
| `backup/testdata/legacy.json` | Live. A real v1.x export, used as the fixture. |
| `helpers/helpers.go` | Live. `GetStorePath` and `SetDevStorePath`. |
| `helpers/helpers_test.go` | Live. New. |
| `components/table.go` | Live. Tabwriter-based table renderer; see the contract above. |
| `components/table_test.go` | Live. New. |
| `components/prompt.go` | Live. Standalone prompt component; see the contract above. |
| `components/prompt_test.go` | Live. New. |
| `themes/theme.go` | Live. Lip Gloss text helpers and the logo. |
| `themes/theme_test.go` | Live. New. |
| `views/views.go` | Live. Root `tartarus` Cobra command; registers `cmds`. |
| `views/commands.go` | Live. `cmds` with `ls`/`v`/`c`/`e`/`d`; the `ap`/`ch`/`rm` item handlers are commented out. |
| `views/commands_test.go` | Live. New. |
| `main.go` | Live. Opens and migrates the store, renders the logo and executes Cobra. |
| `main_test.go` | Live. New. |

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
func (r *Repos) GetCommand(name string) (Command, error)
func (r *Repos) InsertCommand(c *Command) error
func (r *Repos) UpdateCommand(c *Command) error
func (r *Repos) DeleteCommand(id uint64) error

// command scripts
func (r *Repos) GetCommandScripts(commandID uint64) ([]CommandScript, error)
func (r *Repos) GetCommandScriptsByName(commandID uint64, searchTerm string) ([]CommandScript, error)
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

`CommandScript` carries a required `Name`, and so do `AppendCommandScript` and
`UpdateCommandScript`. Its name is validated with `ValidateCommandScriptName`,
which applies the same rule as `ValidateCommandName` — 1–12 ASCII lowercase
letters/digits — and returns `ErrCommandScriptNameRequired`,
`ErrCommandScriptNameTooLong` or `ErrCommandScriptNameInvalid`.
`GetCommandScriptsByName` mirrors `GetCommands(searchTerm)` for the scripts
of one command: an empty search matches every script of the command, a nonempty
one is wrapped in `%` and matched with `LIKE` against `name`, and the result is
ordered by id; `commandID == 0` returns `ErrCommandIDRequired` before any query.

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

The tests write NULL directly through the `nullable` helper in
`commands_test.go`, also used by `manuals_test.go`. Both files are active
coverage. The repository write path itself cannot produce NULL descriptions.

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
- **It runs one query per command for that command's scripts**, an N+1 over a pool of
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
  would collide with the `UNIQUE` on `commands.name`, `command_scripts.name`,
  `command_scripts.script` and `manuals.name` anyway.
- **The reset is also what makes a failed import recoverable.** The inserts are
  not atomic — every `repositories` write opens its own transaction, and there is
  no `InsertCommandTx` to compose them into one — so a bad row halfway leaves a
  partial database. Re-running is still safe, because the reset at the top clears
  that partial state before the constraints can collide with it. A single
  transaction would be the better answer and needs new repository methods.
- **The file is validated before the reset, not after.** Every check in
  `validate` is one the database or the repository would refuse on insert — the
  database for the `UNIQUE` constraints, the repository for the script name rule
  — and running them first is the
  difference between a rejected file and a rejected file that also destroyed the
  previous data. `import_test.go` asserts that the pre-existing row survives seven
  rejection cases. The duplicate checks are there because the constraints
  are global: `command_scripts.name` and `command_scripts.script` are unique
  across the whole table, so two different commands cannot hold the same name or
  the same script, and the script set therefore cannot be reset per command.
- **Ids and timestamps in the file are ignored.** `Insert*` stamps
  `time.Now().UTC()` and lets SQLite assign ids from `AUTOINCREMENT`, so imported
  rows carry import time and freshly numbered ids. The old code set `CreatedAt`
  on the struct it passed in, which was already dead — the repository overwrites
  it from the row it reads back. The round-trip test does not compare
  ids across the two databases; it compares names, bodies, scripts and
  descriptions, and separately asserts that each imported script's `command_id`
  matches the command it is nested under.

That last point is where the old code was actually **wrong**, not just wasteful:
it built each script with `CommandID: cmd.ID`, the id from the file, which after a
reset points at whatever row happens to sit at that id now — silently attaching
scripts to the wrong command. `Import` passes the `c.ID` that `InsertCommand` just
assigned. `AppendCommandScript` overwrites `script.CommandID` with the row it reads
back regardless, so passing the wrong id would have been caught, but only after a
failed insert, not before.

Like `Export`, `Import` must not be called from inside a `store.WithTx` callback,
but it fails with a **different sentinel**: `Export`'s reads return
`store.ErrUseTx`, while `Import` reaches `Store.WithTx` through `ResetMigrations`
and gets `store.ErrTxActive`. Both assertions are active.

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
the fixture used by the legacy restore tests.

`~/.tartarus/db/commands` is never read, written or deleted by the new code.
That is the reason the operation is safe to repeat: the user can re-run
`db export` and the restore as many times as they like.

### The mapping

| Legacy document | Becomes | Notes |
| --- | --- | --- |
| `cmd:<name>` | a `commands` row plus one `command_scripts` row per entry of `instructions` | a legacy name can break the command rule (underscores, hyphens, more than 12 characters), so the row is created under a sanitized, truncated name that satisfies `ValidateCommandName`; the mapping from legacy name to command name is derived from the file alone, so a rerun derives the same name and finds the row instead of inserting a duplicate. `instructions` is the legacy name for what is now `command_scripts`; a legacy instruction carries no name, so one is generated from the sanitized, truncated command name plus the instruction position, must satisfy `ValidateCommandScriptName`, and is made unique against every `command_scripts.name` |
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
- **A legacy instruction whose script already belongs to another command is
  skipped and named**, not inserted and not fatal. This is the consequence of
  `command_scripts.script` being `UNIQUE` across the whole table (item 5 in the
  outstanding work): a script cannot be shared, so the second command to want it
  cannot have it.

The file is validated **before** the first write, same reasoning as `Import`:
every check is one the schema would refuse anyway, and the export belongs to the
user. `commands.name` and `manuals.name` are `UNIQUE`; `command_scripts.script` is
too, so a script repeated inside the file, or across two documents in it, is
rejected up front rather than halfway through. `restore-legacy_test.go` asserts
that the pre-existing row survives eight rejection cases.

Nothing here is atomic: every `repositories` write opens its own transaction and
there is no `InsertCommandTx`. That was a real risk for `Import` and is
tolerable here **only** because of the convergence above — which is the same
reason `Import` can survive its own partial failure, by resetting first.

The report goes to **stdout** as one summary line plus one line per skipped
document or script. The signature returns only an error, and a restore the user
ran by hand deserves an answer in words; if that ever changes shape, return a
struct instead of printing.

### Dispatch and the two sentinels

Legacy restore is not dispatched by the current Cobra CLI. The earlier
`flag.Visit` and `log.Fatal` flow belonged to the removed startup flag.

Like `Export`, `RestoreFromLegacy` must not be called from inside a
`store.WithTx` callback, but it fails with the **read** sentinel:
`readLegacyState` runs first and goes through `Store.Query`, so the error is
`store.ErrUseTx`. `Import` gets `store.ErrTxActive` instead, from
`ResetMigrations`. Both assertions are active.

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
belongs here. All test files are active. The helper names below list the current
definitions:

| Helper | Why it exists |
| --- | --- |
| `repositories.newTestStore` | the only door into a database from outside `store`, and why it must not run in parallel |
| `backup.newTestStore`, `exportPath`, `readExport` | a third copy of the same door, for the same reason, plus a scratch path and a reparse of what `Export` wrote |
| `backup.legacyPath`, `legacyFixture` | point at `backup/testdata/legacy.json` and restore from it, so the fixture is a real v1.x export rather than an invented one |
| `backup.scriptTexts`, `commandNames`, `manualNames` | read one shape back out of the repositories, so an assertion can name a command by its name instead of by id |
| `repositories.nullable` | the only way to write SQL NULL, which the repository's own write path cannot produce |
| `repositories.insertTestCommand`, `insertTestCommandScript`, `insertManual` | set up rows the repository API would not allow |
| `store.newEmptyStore` | `store.newTestStore` creates an `items` table, which gets in the way of asserting on what `RunMigrations` builds |
| `helpers.setVersion`, `markRoot` | swap the link-time version and plant a `go.mod` marker |
| `components.captureStdout`, `withStdin` | capture `os.Stdout`, and point `os.Stdin`/`os.Stderr` at pipes, for `Render` and the public prompt wrappers |
| `components.failingReader`, `failingWriter` | force the read and write error paths the internal `prompt`/`confirm` wrap |
| `themes.captureStdout` | capture `RenderLogo`'s output, a second copy for the same reason as the components one |
| `views.newTestRepos`, `withStdin`, `captureStdout`, `runCommands`, `insertTestCommand` | a store-backed `Repos`, the same two stdio swaps, a driver that builds a fresh `NewCommands` and executes it, and a row the Cobra handlers can act on |
| `main.captureStdout` | capture `run`'s logo and help output, a third copy of the same helper |

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
| `command_scripts` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; non-null `command_id`; globally unique, non-null `name`; globally unique, non-null `script`; nullable `description`; non-null `created_at` and `updated_at`. |
| `manuals` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; unique, non-null `name`; non-null `body`; nullable `description`; non-null `created_at` and `updated_at`. |
| `migrations` | `id INTEGER PRIMARY KEY AUTOINCREMENT`; non-null `version_date` and `version_number`. |

`command_scripts.command_id` references `commands.id` with both
`ON DELETE CASCADE` and `ON UPDATE CASCADE`. Foreign keys are enabled through
the store DSN. `name` and `script` uniqueness are global, not scoped to
`command_id`.
The repository additionally validates command names and command script names as
1–12 ASCII lowercase letters/digits; that restriction is not a SQL `CHECK`
constraint.

Both the version-1 down migration and `0000_drop_all_tables.sql` drop
`command_scripts` before `commands`. Only the reset script drops `migrations`.
Neither script includes the former `command_items` table; see the existing
version-1 database caveat above.

### Naming and backup compatibility

Use `command_scripts` in current SQL and `CommandScript` / `*CommandScript*`
methods in the repository API. `command_items` is the old SQL name.
The backup format was renamed to match: `backup/backup.go` now nests
`Scripts []commandScript` under `json:"scripts"`, and the private type is
`commandScript`. This is an intentional change to the serialized format, not a
side effect of the SQL/API rename. There is **no backward compatibility**: a file
written by an older build carries `"items"`, which `json.Unmarshal` ignores, so
`Import` reads the command with an empty `Scripts` slice and silently drops every
script. Old backups must be re-exported with a current build, or the `items` key
renamed to `scripts` by hand. `commandScript` carries a `name` field, matching the
SQL column, and a script name is required on import. Legacy `instructions` entries
map to `command_scripts` rows.

### Rename follow-up

The rename is complete in code and the tests are active:

- `store/migration_test.go` now asserts against `command_scripts`, including the
  foreign-key reject and cascade cases.
- The reset comment in `store/migration.go` names `command_scripts`.
- `repositories/commands_test.go` uses `CommandScript` and the current
  `GetCommands(searchTerm)` signature.
- `backup/import_test.go` and `backup/restore-legacy_test.go` use
  `CommandScript`/`scripts`, pass the required script `name`, and no longer use
  the paginated `GetCommands` signature.

The only remaining `command_items`/`CommandItem` spelling is in the disabled
`views/commands.go` item handlers (`appendCommandItem`, `changeCommandItem`,
`removeCommandItem`), which are left switched off on purpose.

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
3. **Backup has no CLI handlers.** `backup.Export`, `backup.Import` and
   `backup.RestoreFromLegacy` are implemented and tested but not registered as
   Cobra commands. The command views are already wired to `components.Table`.
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
    already have them. `GetCommand` looks up by name and returns a zero `Command`
    with no error when nothing matches, while `GetManual` looks up by id and does
    not reject `id == 0` although every `Update*`/`Delete*` does. `GetCommands`
    does not document the nil slice that `GetManuals` documents.

## Style

- Comments and error messages are in English, in tests too.
- SQL always uses `?` placeholders, never string interpolation.
- Errors are always wrapped or returned; no `panic` on a normal path.
