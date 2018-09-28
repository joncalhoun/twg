package suite

import "errors"

type User struct {
	ID    int
	Email string
}

// Errors returned by the UserStore
var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email is taken")
)

type UserStore interface {
	Create(*User) error
	ByID(id int) (*User, error)
	ByEmail(email string) (*User, error)
	Delete(*User) error
}
