package builder

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// 0. flag.Parse() if you need flags
	exitCode := run(m)
	os.Exit(exitCode)
}

func run(m *testing.M) int {
	const (
		dropDB          = `DROP DATABASE IF EXISTS test_user_store;`
		createDB        = `CREATE DATABASE test_user_store;`
		createUserTable = `CREATE TABLE users (
												 id SERIAL PRIMARY KEY,
												 name TEXT,
												 email TEXT UNIQUE NOT NULL
											 );`
	)

	builder, err := sql.Open("postgres", "host=localhost port=5432 user=jon sslmode=disable")
	if err != nil {
		panic(fmt.Errorf("sql.Open() err = %s", err))
	}
	defer builder.Close()

	_, err = builder.Exec(dropDB)
	if err != nil {
		panic(fmt.Errorf("builder.Exec() err = %s", err))
	}
	_, err = builder.Exec(createDB)
	if err != nil {
		panic(fmt.Errorf("builder.Exec() err = %s", err))
	}
	// teardown
	defer func() {
		_, err = builder.Exec(dropDB)
		if err != nil {
			panic(fmt.Errorf("builder.Exec() err = %s", err))
		}
	}()

	db, err := sql.Open("postgres", "host=localhost port=5432 user=jon sslmode=disable dbname=test_user_store")
	if err != nil {
		panic(fmt.Errorf("sql.Open() err = %s", err))
	}
	defer db.Close()
	_, err = db.Exec(createUserTable)
	if err != nil {
		panic(fmt.Errorf("db.Exec() err = %s", err))
	}

	return m.Run()
}

func userStore(t *testing.T) (*UserStore, func()) {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=jon sslmode=disable dbname=test_user_store")
	if err != nil {
		t.Fatalf("sql.Open() err = %s", err)
		return nil, nil
	}
	us := &UserStore{
		sql: db,
	}
	return us, func() {
		db.Close()
	}
}

func TestUserStore(t *testing.T) {
	us, teardown := userStore(t)
	defer teardown()
	t.Run("Find", testUserStore_Find(us))
	t.Run("Create", testUserStore_Find(us))
	t.Run("Delete", testUserStore_Find(us))
	t.Run("Subscribe", testUserStore_Find(us))
}

func testUserStore_Find(us *UserStore) func(t *testing.T) {
	return func(t *testing.T) {
		jon := &User{
			Name:  "Jon Calhoun",
			Email: "jon@calhoun.io",
		}
		err := us.Create(jon)
		if err != nil {
			t.Errorf("us.Create() err = %s", err)
		}
		defer func() {
			err := us.Delete(jon.ID)
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
			{"Found", jon.ID, jon, nil},
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

// func newServer(t *testing.T) (*httptest.Server, func()) {
// 	t.Helper()
// 	db := ...
//   cache, err := ...
//   if err != nil {
//     t.Fatalf(...)
//   }
//   h := Handler(db, cache...)
//   server := httptest.NewServer(h)
//   // ...
//   return server, func() {
//     db.Close()
//     cache.Close()
//     server.Close()
//   }
// }

// func TestRoutes(t *testing.T) {
//   server, teardown := newServer(t)
//   defer teardown()

//   t.Run("home", func(t *testing.T) { testGetRoute(t, app, "/") })
//   t.Run("about", func(t *testing.T) { testGetRoute(t, app, "/about") })
//   t.Run("contact", func(t *testing.T) { testGetRoute(t, app, "/contact") })
// }
