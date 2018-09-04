package psql

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func TestUserStore(t *testing.T) {
	// queries
	const (
		createDB        = `CREATE DATABASE test_user_store;`
		dropDB          = `DROP DATABASE IF EXISTS test_user_store;`
		createUserTable = `CREATE TABLE users (
												 id SERIAL PRIMARY KEY,
												 name TEXT,
												 email TEXT UNIQUE NOT NULL
											 );`
	)

	psql, err := sql.Open("postgres", "host=localhost port=5432 user=jon sslmode=disable")
	if err != nil {
		t.Fatalf("sql.Open() err = %s", err)
	}
	defer psql.Close()

	_, err = psql.Exec(dropDB)
	if err != nil {
		t.Fatalf("psql.Exec() err = %s", err)
	}
	_, err = psql.Exec(createDB)
	if err != nil {
		t.Fatalf("psql.Exec() err = %s", err)
	}
	// teardown
	defer func() {
		_, err = psql.Exec(dropDB)
		if err != nil {
			t.Errorf("psql.Exec() err = %s", err)
		}
	}()

	db, err := sql.Open("postgres", "host=localhost port=5432 user=jon sslmode=disable dbname=test_user_store")
	if err != nil {
		t.Fatalf("sql.Open() err = %s", err)
	}
	defer db.Close()
	_, err = db.Exec(createUserTable)
	if err != nil {
		t.Errorf("db.Exec() err = %s", err)
	}

	us := &UserStore{
		sql: db,
	}
	t.Run("Find", testUserStore_Find(us))
	t.Run("Find", testUserStore_Find(us))
	t.Run("Find", testUserStore_Find(us))
	// t.Run("Create", testUserStore_Find(us))

}

func testUserStore_Find(us *UserStore) func(t *testing.T) {
	return func(t *testing.T) {
		user := &User{
			Name:  "Jon Calhoun",
			Email: "jon@calhoun.io",
		}
		err := us.Create(user)
		if err != nil {
			t.Errorf("us.Create() err = %s", err)
		}
		defer func() {
			err := us.Delete(user.ID)
			if err != nil {
				t.Errorf("us.Delete() err = %s", err)
			}
		}()

		tests := []struct {
			name    string
			id      int
			want    *User
			wantErr error
		}{
			{"Found", user.ID, user, nil},
			{"Not Found", -1, nil, ErrNotFound},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := us.Find(tc.id)
				if err != tc.wantErr {
					t.Errorf("us.Find() err = %s", err)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("us.Find() = %+v, want %+v", got, tc.want)
				}
			})
		}
	}
}
