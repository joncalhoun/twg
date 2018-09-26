package emailapp

import "strings"

// Or SendGridClient - I just happen to use Mailgun.
type MailgunClient struct {
	// stuff here
}

func (mc *MailgunClient) Welcome(name, email string) error {
	// send out a welcome email to the user!
	return nil
}

// this is all fake just to make the demo work
type User struct{}
type UserStore struct{}

func (us *UserStore) Create(name, email string) (*User, error) {
	// pretend to add user to DB
	return &User{}, nil
}

type EmailClient interface {
	Welcome(name, email string) error
}

func Signup(name, email string, ec EmailClient, us *UserStore) (*User, error) {
	email = strings.ToLower(email)
	user, err := us.Create(name, email)
	if err != nil {
		return nil, err
	}
	err = ec.Welcome(name, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
