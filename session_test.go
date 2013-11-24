package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DatabaseInitSQL = []string{
	`CREATE TABLE users(id INTEGER NOT NULL PRIMARY KEY, name STRING UNIQUE)`,
	`INSERT INTO users(id, name) VALUES (1, 'bob'), (2, 'mike'), (3, 'john')`,
}

type User struct {
	Id   int64
	Name string
}

func withConnection(t *testing.T, fn func(*sql.DB)) {
	dbpath := fmt.Sprintf(os.TempDir()+"/sqlite3.%d.db", time.Now().UnixNano())
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		t.Fatalf("cannot create database: %s", err)
	}

	for _, sql := range DatabaseInitSQL {
		if _, err := db.Exec(sql); err != nil {
			t.Fatalf("cannot execute init sql: %s", err)
		}
	}
	fn(db)
	defer db.Close()
	defer os.Remove(dbpath)
}

func TestSessionExec(t *testing.T) {
	withConnection(t, func(db *sql.DB) {
		session := Use(db, Sqlite3Dialect)
		_, err := session.Exec("INSERT INTO users(name) VALUES(?)", "garry")
		if err != nil {
			t.Fatalf("cannot insert user: %s", err)
		}
		_, err = session.Exec("INSERT INTO users(name) VALUES(?)", "garry")
		if err == nil {
			t.Fatalf("cannot insert user: %s", err)
		}

		if err := session.Rollback(); err != nil {
			t.Fatalf("cannot rollback session: %s", err)
		}

		_, err = session.Exec("INSERT INTO users(name) VALUES(?)", "garry")
		if err != nil {
			t.Fatalf("cannot insert user: %s", err)
		}

		if err := session.Commit(); err != nil {
			t.Fatalf("cannot commit session: %s", err)
		}

		_, err = session.Exec("INSERT INTO users(name) VALUES(?)", "garry")
		if err == nil {
			t.Fatalf("cannot insert user: %s", err)
		}
	})
}

func TestSessionQuery(t *testing.T) {
	withConnection(t, func(db *sql.DB) {
		session := Use(db, Sqlite3Dialect)
		rows, err := session.Query("SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("cannot query database: %s", err)
		}
		users := make([]*User, 0, 3)
		for rows.Next() {
			user := &User{}
			if err := rows.Scan(&user.Id, &user.Name); err != nil {
				t.Fatalf("cannot scan row: %s", err)
			}
			users = append(users, user)
		}
		if len(users) != 3 {
			t.Fatalf("expected 3 users, got %d", len(users))
		}
	})
}
