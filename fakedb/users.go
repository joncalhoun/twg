package fakedb

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("fakedb: resource could not be located")
)

type User struct {
	ID    int
	Email string
}

func NewUserDB() *UserDB {
	return &UserDB{
		store:  make(map[string]int, 0),
		nextID: 1,
	}
}

type UserDB struct {
	store  map[string]int
	nextID int
}

func (udb *UserDB) Create(user *User) error {
	if id, ok := udb.store[user.Email]; ok {
		return fmt.Errorf("Email %s is taken by the user with the ID: %d", user.Email, id)
	}
	user.ID = udb.nextID
	udb.nextID++
	return nil
}

func (udb *UserDB) FindByEmail(email string) (*User, error) {
	if id, ok := udb.store[email]; ok {
		return &User{
			ID:    id,
			Email: email,
		}, nil
	}
	return nil, ErrNotFound
}
