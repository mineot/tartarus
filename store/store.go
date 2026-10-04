package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"tartarus/helpers"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db  *sql.DB
	tx  *sql.Tx
	ctx context.Context
}

func (s *Store) ensureOpen() error {
	if s.db == nil {
		return errors.New("store is not open")
	}

	return nil
}

func (s *Store) Close() error {
	var rollbackErr error
	var closeErr error

	if s.tx != nil {
		rollbackErr = s.tx.Rollback()
		s.tx = nil

		if errors.Is(rollbackErr, sql.ErrTxDone) {
			rollbackErr = nil
		}
	}

	if s.db != nil {
		closeErr = s.db.Close()
		s.db = nil
	}

	s.ctx = nil

	return errors.Join(rollbackErr, closeErr)
}

func (s *Store) Open() error {
	if s.db != nil {
		return errors.New("store is already open")
	}

	var ctx context.Context
	var db *sql.DB
	var err error
	var path string

	if helpers.IsProduction() {
		if path, err = helpers.GetProductionStorePath("tartarus.db"); err != nil {
			return err
		}
	} else {
		if path, err = helpers.GetDevelopmentStorePath("dev.db"); err != nil {
			return err
		}
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if db, err = sql.Open("sqlite3", path); err != nil {
		return err
	}

	ctx = context.Background()

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}

	s.db = db
	s.ctx = ctx
	return nil
}

func (s *Store) Query(query string, args ...any) (*sql.Rows, error) {
	if err := s.ensureOpen(); err != nil {
		return nil, err
	}

	if s.tx != nil {
		return s.tx.QueryContext(s.ctx, query, args...)
	}

	return s.db.QueryContext(s.ctx, query, args...)
}

func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	if err := s.ensureOpen(); err != nil {
		return nil, err
	}

	if s.tx != nil {
		return s.tx.ExecContext(s.ctx, query, args...)
	}

	return s.db.ExecContext(s.ctx, query, args...)
}

func (s *Store) Begin() error {
	if err := s.ensureOpen(); err != nil {
		return err
	}

	if s.tx != nil {
		return errors.New("transaction is already active")
	}

	tx, err := s.db.BeginTx(s.ctx, nil)

	if err != nil {
		return err
	}

	s.tx = tx

	return nil
}

func (s *Store) Commit() error {
	if s.tx == nil {
		return nil
	}

	err := s.tx.Commit()
	s.tx = nil

	return err
}

func (s *Store) Rollback() error {
	if s.tx == nil {
		return nil
	}

	err := s.tx.Rollback()
	s.tx = nil

	return err
}
