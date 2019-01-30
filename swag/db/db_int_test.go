package db

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const testDefaultURL = "postgres://postgres@127.0.0.1:5432/swag_test?sslmode=disable"

var (
	testURL string
)

func init() {
	testURL = os.Getenv("PSQL_URL")
	if testURL == "" {
		testURL = testDefaultURL
	}
}

// TestActiveCampaign_specificTiming covers tests that require specific
// timing such as a campaign that starts at exactly now or ends at exactly
// now.
func TestActiveCampaign_specificTiming(t *testing.T) {
	// Makes sure we don't accidentally use the global DefaultDatabase
	tmp := DefaultDatabase
	DefaultDatabase = nil
	defer func() {
		DefaultDatabase = tmp
	}()

	database, err := Open(testURL)
	if err != nil {
		t.Fatalf("Open() err = %v; want nil", err)
	}
	defer database.Close()

	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()

	// each test case returns the campaign and error it wants from a call
	// to ActiveCampaign
	tests := map[string]func(*testing.T) (*Campaign, error){
		"just started": func(t *testing.T) (*Campaign, error) {
			want, err := database.CreateCampaign(now, now.Add(1*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
		"nearly ended": func(t *testing.T) (*Campaign, error) {
			want, err := database.CreateCampaign(now.Add(-1*time.Hour), now, 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			database.TestReset(t)
			want, wantErr := setup(t)
			campaign, err := database.ActiveCampaign()
			if err := campaignEq(campaign, want); err != nil {
				t.Errorf("ActiveCampaign() %v", err)
			}
			if err != wantErr {
				t.Fatalf("ActiveCampaign() err = %v; want %v", err, wantErr)
			}
		})
	}
	database.TestReset(t)
}

func campaignEq(got, want *Campaign) error {
	// nil == nil
	if got == want {
		return nil
	}
	if got == nil {
		return fmt.Errorf("got nil; want %v", want)
	}
	if want == nil {
		return fmt.Errorf("got %v; want nil", got)
	}
	if got.ID != want.ID {
		return fmt.Errorf("got.ID = %d; want %d", got.ID, want.ID)
	}
	if !got.StartsAt.Equal(want.StartsAt) {
		return fmt.Errorf("got.StartsAt = %v; want %v", got.StartsAt, want.StartsAt)
	}
	if !got.EndsAt.Equal(want.EndsAt) {
		return fmt.Errorf("got.StartsAt = %v; want %v", got.EndsAt, want.EndsAt)
	}
	return nil
}
