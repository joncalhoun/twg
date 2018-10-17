package timing

import (
	"testing"
	"time"
)

func TestSaveUser(t *testing.T) {
	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()

	user := User{}
	SaveUser(&user)
	if user.UpdatedAt != now {
		t.Errorf("user.UpdatedAt = %v, want ~%v", user.UpdatedAt, now)
	}
}

func TestUserSaver_Save(t *testing.T) {
	now := time.Now()
	us := UserSaver{
		now: func() time.Time {
			return now
		},
	}
	user := User{}
	us.Save(&user)
	if user.UpdatedAt != now {
		t.Errorf("user.UpdatedAt = %v, want ~%v", user.UpdatedAt, now)
	}
}
