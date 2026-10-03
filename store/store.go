package store

import (
	"context"
	"database/sql"
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

func (s *Store) Close() error {
	if s.tx != nil {
		s.tx.Rollback()
		s.tx = nil
	}

	if s.ctx != nil {
		s.ctx.Done()
	}

	if s.db == nil {
		return nil
	}

	return s.db.Close()
}

func (s *Store) Open() error {
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
	if s.tx != nil {
		return s.tx.QueryContext(s.ctx, query, args...)
	}

	return s.db.QueryContext(s.ctx, query, args...)
}

func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.ExecContext(s.ctx, query, args...)
	}

	return s.db.ExecContext(s.ctx, query, args...)
}

func (s *Store) Begin() error {
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
