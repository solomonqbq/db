package db

import (
	"database/sql"
	"testing"
)

func TestMappingQuery(t *testing.T) {
	withConnection(t, func(db *sql.DB) {
		session := Use(db, Sqlite3Dialect)

		// single fetch
		func () {
			user := &User{}

			if err := session.Table("users").Query().One(&user); err != ErrMultipleRowsFound {
				t.Fatalf("should not be able to query users without limit: %s", err)
			}
			if err := session.Table("users").Query().Limit(1).One(&user); err != nil {
				t.Fatalf("cannot query single user: %s", err)
			}
			if user.Id == 0 || user.Name == "" {
				t.Fatalf("not all fields were fetched: %#v", user)
			}
		}()

		// multiple fetch, simple query
		func () {
			users := make([]*User, 0, 2)
			if err := session.Table("users").Query().All(&users); err != nil {
				t.Fatalf("cannot query all users: %s", err)
			}
			if len(users) != 3 {
				t.Fatalf("expected to fetch 3 users, got %d", len(users))
			}
			for _, user := range users {
				if user.Id == 0 || user.Name == "" {
					t.Fatalf("not all fields were fetched: %#v", user)
				}
			}
		}()

		// multiple fetch, filter
		func () {
			users := make([]*User, 0, 2)
			if err := session.Table("users").Query().Where("name =", "bob").All(&users); err != nil {
				t.Fatalf("cannot query all users: %s", err)
			}
			if len(users) != 1 {
				t.Fatalf("expected filter to fetch one user, got %d", len(users))
			}
			user := users[0]
			if user.Id != 1 || user.Name != "bob" {
				t.Fatalf("fetched data not fully propagated: %#v", user)
			}
		}()

		// multiple fetch, limit, offset, order
		func () {
			users := make([]*User, 0, 2)
			q := session.Table("users").Query().Limit(2).Offset(1).OrderBy("id")
			if err := q.All(&users); err != nil {
				t.Fatalf("cannot query all users: %s", err)
			}
			if len(users) != 2 {
				t.Fatalf("expected to get 2 users, got %d", len(users))
			}
			if users[0].Id != 2 || users[1].Id != 3 {
				t.Fatalf("expected ascending order, got: [%d, %d]", users[0].Id, users[1].Id)
			}
		}()
	})
}
