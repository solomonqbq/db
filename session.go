package db

import (
	"bufio"
	"database/sql"
	"io"
	"os"
)

type Session struct {
	db      *sql.DB
	tx      *sql.Tx
	log     *logger
	dialect Dialect
}

func Use(db *sql.DB, dialect Dialect) *Session {
	return &Session{
		db:      db,
		log:     standardLogger(),
		dialect: dialect,
	}
}

func ExecFile(db *sql.DB, filepath string) error {
	fd, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer fd.Close()

	commit := true
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func () {
		if commit {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	rd := bufio.NewReader(fd)
	for {
		line, err := rd.ReadString(';')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			commit = false
			return err
		}
		if _, err := tx.Exec(line); err != nil {
			commit = false
			return err
		}
	}
}

func (s *Session) transaction() (*sql.Tx, error) {
	if s.tx == nil {
		tx, err := s.db.Begin()
		if err != nil {
			return nil, err
		}
		s.tx = tx
	}
	return s.tx, nil
}

func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}
	s.log.Debug("Exec: %s  =>  %#v", query, args)
	res, err := tx.Exec(query, args...)
	if err != nil {
		s.log.Warn("Exec error: %s", err)
	}
	return res, err
}

func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}
	s.log.Debug("Query: %s  =>  %#v", query, args)
	rows, err := tx.Query(query, args...)
	if err != nil {
		s.log.Warn("Query error: %s", err)
	}
	return rows, err
}

func (s *Session) Commit() error {
	if s.tx == nil {
		return nil
	}
	err := s.tx.Commit()
	s.tx = nil
	if err != nil {
		s.log.Warn("Commit error: %s", err)
	}
	return err
}

func (s *Session) Rollback() error {
	if s.tx == nil {
		return nil
	}
	err := s.tx.Rollback()
	s.tx = nil
	if err != nil {
		s.log.Warn("Rollback error: %s", err)
	}
	return err
}

func (s *Session) Table(name string) *TableMapping {
	return &TableMapping{
		session: s,
		name:    name,
	}
}
