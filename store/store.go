package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"tartarus/helpers"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Open() error {
	var err error
	var path string

	if helpers.IsProduction() {
		path, err = helpers.GetProductionStorePath("tartarus.db")

		if err != nil {
			return err
		}
	} else {
		path, err = helpers.GetDevelopmentStorePath("dev.db")

		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	s.db = db

	return nil
}

func (s *Store) Query(query string, args ...any) (*sql.Rows, error) {
	rows, err := s.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return rows, nil
}

func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	result, err := s.db.Exec(query, args...)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) Begin() {
	s.Begin()
}

func (s *Store) Rollback() {
	s.Rollback()
}

func (s *Store) Commit() {
	s.Commit()
}
