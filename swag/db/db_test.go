package db_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joncalhoun/twg/swag/db"
)

const defaultURL = "postgres://postgres@127.0.0.1:5432/swag_test?sslmode=disable"

var (
	testURL string
)

func init() {
	testURL = os.Getenv("PSQL_URL")
	if testURL == "" {
		testURL = defaultURL
	}
	if db.DB != nil {
		db.DB.Close()
	}
	db.Open(testURL)
}

func TestCampaigns(t *testing.T) {
	setup := reset
	teardown := reset

	t.Run("Create", func(t *testing.T) {
		testCreateCampaign(t, setup, teardown)
	})
}

func testCreateCampaign(t *testing.T, setup, teardown func(*testing.T)) {
	tests := map[string]*db.Campaign{
		"active": &db.Campaign{
			StartsAt: time.Now(),
			EndsAt:   time.Now().Add(2 * time.Hour),
			Price:    1000,
		},
		"expired": &db.Campaign{
			StartsAt: time.Now().Add(-2 * time.Hour),
			EndsAt:   time.Now().Add(-1 * time.Hour),
			Price:    1000,
		},
	}
	for name, campaign := range tests {
		t.Run(name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			nBefore := count(t, "campaigns")

			start := campaign.StartsAt
			end := campaign.EndsAt
			price := campaign.Price
			created, err := db.CreateCampaign(start, end, price)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			if created.ID <= 0 {
				t.Errorf("ID = %d; want > 0", created.ID)
			}
			if !created.StartsAt.Equal(start) {
				t.Errorf("StartsAt = %v; want %v", created.StartsAt, start)
			}

			nAfter := count(t, "campaigns")
			if diff := nAfter - nBefore; diff != 1 {
				t.Fatalf("campaign count difference = %d; want %d", diff, 1)
			}

			got, err := db.GetCampaign(created.ID)
			if err != nil {
				t.Fatalf("GetCampaign() err = %v; want nil", err)
			}
			if got.ID <= 0 {
				t.Errorf("ID = %d; want > 0", got.ID)
			}
			if !got.StartsAt.Equal(start) {
				t.Errorf("StartsAt = %v; want %v", got.StartsAt, start)
			}

			db.DB.Exec("DELETE FROM campaigns")
		})
	}
}

func reset(t *testing.T) {
	_, err := db.DB.Exec("DELETE FROM orders")
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	_, err = db.DB.Exec("DELETE FROM campaigns")
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
}

func count(t *testing.T, table string) int {
	var n int
	err := db.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
	if err != nil {
		t.Fatalf("Scan() err = %v; want nil", err)
	}
	return n
}
