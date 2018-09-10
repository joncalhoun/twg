package race

import (
	"database/sql"

	"github.com/pkg/errors"
)

// Common errors that you will likely want to account for in your code.
// Any other errors are wrapped with context vai the github.com/pkg/errors
// package and returned but are harder to use an if/switch to match.
var (
	ErrNotFound = errors.New("race: resource could not be located")
)

// User is an example user model. This typically wouldn't be defined in
// this package but is done here for simplicity.
type User struct {
	ID      int
	Name    string
	Email   string
	Balance int
}

type UserStore interface {
	Tx(func(UserStore) error) error
	Find(id int) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id int) error
}

// PsqlUserStore is used to interact with our user store.
type PsqlUserStore struct {
	tx interface {
		Begin() (*sql.Tx, error)
	}
	sql interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	}
}

func (pus *PsqlUserStore) Tx(fn func(us UserStore) error) error {
	tx, err := pus.tx.Begin()
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "race: failed to begin transaction")
	}
	txStore := &PsqlUserStore{
		sql: tx,
	}
	err = fn(txStore)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "race: failed to commit transaction")
	}
	return nil
}

// Find will retrieve a user with the provided ID or return ErrNotFount
// if the user isn't located. Other errors are wrapped with context
// but are otherwise wrapped as-is.
func (pus *PsqlUserStore) Find(id int) (*User, error) {
	const query = `SELECT id, name, email, balance FROM users WHERE id=$1;`
	row := pus.sql.QueryRow(query, id)
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Balance)
	switch err {
	case sql.ErrNoRows:
		return nil, ErrNotFound
	case nil:
		return &user, nil
	default:
		return nil, errors.Wrap(err, "race: error querying for user by id")
	}
}

// Create will create a new user in the DB using the provided user and
// will update the ID of the provided user. If there is an error it will
// be wrapped and returned.
func (pus *PsqlUserStore) Create(user *User) error {
	const query = `INSERT INTO users (name, email, balance) VALUES ($1, $2, $3) RETURNING id`
	err := pus.sql.QueryRow(query, user.Name, user.Email, user.Balance).Scan(&user.ID)
	if err != nil {
		return errors.Wrap(err, "race: error creating new user")
	}
	return nil
}

// Update will update a user in the DB with the provided info.
func (pus *PsqlUserStore) Update(user *User) error {
	const query = `UPDATE users SET name=$2, email=$3, balance=$4 WHERE id=$1`
	_, err := pus.sql.Exec(query, user.ID, user.Name, user.Email, user.Balance)
	if err != nil {
		return errors.Wrap(err, "race: error updating user")
	}
	return nil
}

// Delete will delete a user form the DB. If there is an error it will
// be wrapped and returned.
func (pus *PsqlUserStore) Delete(id int) error {
	const query = `DELETE FROM users WHERE id=$1;`
	_, err := pus.sql.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "race: error deleting user")
	}
	return nil
}

func Spend(tx interface {
	Tx(func(UserStore) error) error
}, userID int, amount int) error {
	return tx.Tx(func(us UserStore) error {
		user, err := us.Find(userID)
		if err != nil {
			return err
		}
		user.Balance -= amount
		return us.Update(user)
	})
}
