package db

import (
	"fmt"
	"testing"
	"time"
)

// TestActiveCampaign_specificTiming covers tests that require specific
// timing such as a campaign that starts at exactly now or ends at exactly
// now.
func TestActiveCampaign_specificTiming(t *testing.T) {
	reset(t)
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
			want, err := CreateCampaign(now, now.Add(1*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
		"nearly ended": func(t *testing.T) (*Campaign, error) {
			want, err := CreateCampaign(now.Add(-1*time.Hour), now, 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			want, wantErr := setup(t)
			defer reset(t)
			campaign, err := ActiveCampaign()
			if err := campaignEq(campaign, want); err != nil {
				t.Errorf("ActiveCampaign() %v", err)
			}
			if err != wantErr {
				t.Fatalf("ActiveCampaign() err = %v; want %v", err, wantErr)
			}
		})
	}
}

func reset(t *testing.T) {
	_, err := DB.Exec("DELETE FROM orders")
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	_, err = DB.Exec("DELETE FROM campaigns")
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
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
