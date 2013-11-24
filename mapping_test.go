package db

import (
	"database/sql"
	"testing"
)

func TestMappingSave(t *testing.T) {
	withConnection(t, func(db *sql.DB) {
		session := Use(db, Sqlite3Dialect)
		users := session.Table("users")

		user := &User{Name: "jim"}
		created, err := users.Save(&user)
		if err != nil {
			t.Fatalf("cannot save jim user: %s", err)
		}
		if !created {
			t.Fatal("user jim was saved, but not created")
		}
		if user.Id != 4 {
			t.Fatalf("<user jim>.Id should be 4, but was set to %d", user.Id)
		}

		user.Name = "jimmy"
		created, err = users.Save(&user)
		if err != nil {
			t.Fatalf("cannot save jim user: %s", err)
		}
		if created {
			t.Fatal("user jim was saved, but not updated")
		}

		rows, err := session.Query("SELECT name FROM users WHERE id = ?", user.Id)
		if err != nil {
			t.Fatalf("cannot query users table: %s", err)
		}
		rows.Next()
		var name string
		rows.Scan(&name)
		rows.Close()
		if name != "jimmy" {
			t.Fatalf("expected user name to be 'jimmy', but got '%s'", name)
		}
	})
}

func TestMappingDelete(t *testing.T) {
	withConnection(t, func(db *sql.DB) {
		session := Use(db, Sqlite3Dialect)

		user := &User{Id: 1}

		if err := session.Table("users").Delete(&user); err != nil {
			t.Fatalf("cannot delete first user: %s", err)
		}

		if err := session.Table("users").Delete(&user); err != ErrNotFound {
			t.Fatalf("deleting non existing row should fail: %s", err)
		}
	})
}
