package store

import (
	"context"
	"database/sql"
)

// Tx is a transaction in progress. It exists only inside the WithTx callback and
// cannot be used outside of it.
//
// It is not safe for concurrent use: every operation has to come from the same
// goroutine that called WithTx.
type Tx struct {
	tx  *sql.Tx
	ctx context.Context
}

// Exec runs a statement inside the transaction.
func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

// Query runs a query inside the transaction.
//
// Unlike a query on the Store, the rows have to be closed before the commit, and
// the commit fails with sql.ErrTxDone if they are not. Iterating to the end
// closes them on its own; for partial iteration, use defer rows.Close() inside
// the callback.
func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

// QueryRow runs a query that returns at most one row.
func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}
