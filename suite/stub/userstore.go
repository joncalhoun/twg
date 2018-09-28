package stub

import (
	"github.com/joncalhoun/twg/suite"
)

type UserStore struct{}

func (us *UserStore) Create(user *suite.User) error {
	user.ID = 123
	return nil
}

func (us *UserStore) ByID(id int) (*suite.User, error) {
	if id == 123 {
		return nil, suite.ErrNotFound
	}
	return &suite.User{
		ID:    1,
		Email: "test@test.com",
	}, nil
}

func (us *UserStore) ByEmail(email string) (*suite.User, error) {
	return nil, nil
}

func (us *UserStore) Delete(*suite.User) error {
	return nil
}
