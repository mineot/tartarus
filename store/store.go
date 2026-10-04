package store

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	// ErrClosed is returned by any operation on a Store that is already closed.
	ErrClosed = errors.New("store: closed")

	// ErrTxActive is returned when WithTx is called from inside another WithTx.
	ErrTxActive = errors.New("store: a transaction is already active")

	// ErrUseTx is returned by the Store's Exec and Query while a transaction is
	// in progress.
	ErrUseTx = errors.New("store: inside WithTx use the *Tx handle, not the *Store")

	// ErrNoFunc is returned when WithTx is given a nil function.
	ErrNoFunc = errors.New("store: WithTx requires a function")
)

const (
	busyTimeout = 5 * time.Second
	dirPerm     = 0755
)

// Store owns the lifecycle of the connection to the database.
//
// Concurrency: a Store may be shared across goroutines, but the way to reach the
// database goes through WithTx. While a transaction is in progress, Exec and
// Query return ErrUseTx instead of waiting for the connection: with a single
// connection in the pool, waiting there would serialize access without changing
// the outcome. Inside the WithTx callback, use only the *Tx.
//
// Transactions are callback scoped and never instance state. The *Tx exists only
// between the begin and the commit or rollback, which stops one operation's
// transaction from leaking into the next.
type Store struct {
	db     *sql.DB
	ctx    context.Context
	mu     sync.Mutex
	inTx   atomic.Bool
	closed atomic.Bool
}

// New opens path and returns a Store ready for use, creating the database
// directory if it does not exist yet.
//
// The context is kept as the parent of every operation, so it should live as
// long as the process rather than as long as a request: a request context
// cancelled here takes the whole Store down with it.
func New(ctx context.Context, path string) (*Store, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return nil, errors.New("store: path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dsn(path))

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(0)

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db, ctx: ctx}, nil
}

// dsn builds the SQLite connection parameters.
//
// _busy_timeout gives a transaction room to wait for the connection in use
// instead of failing right away with SQLITE_BUSY.
//
// _journal_mode lets one reader and one writer make progress at the same time.
// The mode is written to the file header, so it applies to every future open of
// the database.
//
// _txlock swaps the deferred BEGIN for a BEGIN IMMEDIATE, which takes the write
// lock up front. Without it a transaction that only reads can fail when it is
// promoted to a write, and _busy_timeout does not cover that case.
func dsn(path string) string {
	var q = url.Values{}

	q.Set("_busy_timeout", strconv.FormatInt(busyTimeout.Milliseconds(), 10))
	q.Set("_journal_mode", "WAL")
	q.Set("_txlock", "immediate")

	return path + "?" + q.Encode()
}

// guard stops a Store operation from using the connection while it belongs to a
// transaction.
func (s *Store) guard() error {
	if s.closed.Load() {
		return ErrClosed
	}

	if s.inTx.Load() {
		return ErrUseTx
	}

	return nil
}

// Exec runs a statement outside a transaction.
func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}

	return s.db.ExecContext(s.ctx, query, args...)
}

// Query runs a query outside a transaction.
//
// Rows have to be closed before the next operation or they hold on to the
// connection. Iterating to the end closes them on its own; for partial
// iteration, use defer rows.Close().
func (s *Store) Query(query string, args ...any) (*sql.Rows, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}

	return s.db.QueryContext(s.ctx, query, args...)
}

// WithTx opens a transaction, hands a *Tx to fn, and decides the outcome from
// the return value: nil commits, an error rolls back and is returned.
//
// One transaction at a time per Store. Concurrent calls to WithTx get ErrTxActive
// instead of waiting.
//
// Any number of operations can run inside the same callback, and they all commit
// or roll back together:
//
//	err := store.WithTx(func(tx *store.Tx) error {
//		if _, err := tx.Exec(insertCommand, ...); err != nil {
//			return err
//		}
//
//		return nil
//	})
//
// If fn panics, the transaction is rolled back before the panic propagates.
// Otherwise the connection would stay checked out and, with a single connection
// in the pool, it would lock up the whole Store.
func (s *Store) WithTx(fn func(*Tx) error) error {
	if fn == nil {
		return ErrNoFunc
	}

	if s.closed.Load() {
		return ErrClosed
	}

	if s.inTx.Load() {
		return ErrTxActive
	}

	s.mu.Lock()

	defer s.mu.Unlock()

	if s.closed.Load() {
		return ErrClosed
	}

	tx, err := s.db.BeginTx(s.ctx, nil)

	if err != nil {
		return err
	}

	s.inTx.Store(true)

	defer s.inTx.Store(false)

	committed := false

	defer func() {
		if !committed {
			ignoreTxDone(tx.Rollback())
		}
	}()

	if err = fn(&Tx{tx: tx, ctx: s.ctx}); err != nil {
		return errors.Join(err, ignoreTxDone(tx.Rollback()))
	}

	committed = true

	return tx.Commit()
}

// Close closes the connection. It is idempotent: closing twice returns nil, so a
// defer does not have to check.
func (s *Store) Close() error {
	s.mu.Lock()

	defer s.mu.Unlock()

	if s.closed.Load() {
		return nil
	}

	s.closed.Store(true)

	return s.db.Close()
}

// ignoreTxDone discards sql.ErrTxDone, which only means the transaction had
// already finished before the rollback arrived.
func ignoreTxDone(err error) error {
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}

	return err
}
