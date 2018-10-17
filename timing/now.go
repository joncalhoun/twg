package timing

import (
	"time"
)

var (
	timeNow = time.Now
)

type User struct {
	UpdatedAt time.Time
}

func SaveUser(user *User) {
	t := timeNow()
	user.UpdatedAt = t
	// ... save the user
}

type UserSaver struct {
	now func() time.Time
}

func (us *UserSaver) Save(user *User) {
	t := us.now()
	user.UpdatedAt = t
	// ... save the user
}
