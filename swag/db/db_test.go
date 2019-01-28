package db_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"

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

func TestCreateCampaign(t *testing.T) {
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
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			defer reset(t)

			nBefore := count(t, "campaigns")

			start := want.StartsAt
			end := want.EndsAt
			price := want.Price
			created, err := db.CreateCampaign(start, end, price)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			if created.ID <= 0 {
				t.Errorf("CreateCampaign() ID = %d; want > 0", created.ID)
			}
			want.ID = created.ID
			if err := campaignEq(created, want); err != nil {
				t.Errorf("CreateCampaign() %v", err)
			}

			nAfter := count(t, "campaigns")
			if diff := nAfter - nBefore; diff != 1 {
				t.Fatalf("CreateCampaign() increased campaign count by %d; want %d", diff, 1)
			}

			got, err := db.GetCampaign(created.ID)
			if err != nil {
				t.Fatalf("GetCampaign() err = %v; want nil", err)
			}
			if err := campaignEq(got, want); err != nil {
				t.Errorf("GetCampaign() %v", err)
			}
		})
	}
}

func TestActiveCampaign(t *testing.T) {
	reset(t)

	// each test case returns the campaign and error it wants from a call
	// to ActiveCampaign
	tests := map[string]func(*testing.T) (*db.Campaign, error){
		"just started": func(t *testing.T) (*db.Campaign, error) {
			// This test case is less than perfect... Can we fix it?
			want, err := db.CreateCampaign(time.Now(), time.Now().Add(1*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
		// This test case fails, but we'd like it to pass. How do we fix it?
		// "nearly ended": func(t *testing.T) (*db.Campaign, error) {
		// 	want, err := db.CreateCampaign(time.Now().Add(-1*time.Hour), time.Now(), 900)
		// 	if err != nil {
		// 		t.Fatalf("CreateCampaign() err = %v; want nil", err)
		// 	}
		// 	return want, nil
		// },
		"mid campaign": func(t *testing.T) (*db.Campaign, error) {
			// This test case is less than perfect... Can we fix it?
			want, err := db.CreateCampaign(time.Now().Add(-1*time.Hour), time.Now().Add(time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return want, nil
		},
		"none": func(t *testing.T) (*db.Campaign, error) {
			return nil, sql.ErrNoRows
		},
		"expired recently": func(t *testing.T) (*db.Campaign, error) {
			_, err := db.CreateCampaign(time.Now().Add(-7*24*time.Hour), time.Now().Add(-1*time.Second), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return nil, sql.ErrNoRows
		},
		"future": func(t *testing.T) (*db.Campaign, error) {
			_, err := db.CreateCampaign(time.Now().Add(7*24*time.Hour), time.Now().Add(10*24*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return nil, sql.ErrNoRows
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			want, wantErr := setup(t)
			defer reset(t)
			campaign, err := db.ActiveCampaign()
			if err := campaignEq(campaign, want); err != nil {
				t.Errorf("ActiveCampaign() %v", err)
			}
			if err != wantErr {
				t.Fatalf("ActiveCampaign() err = %v; want %v", err, wantErr)
			}
		})
	}
}

func TestGetCampaign(t *testing.T) {
	reset(t)

	// each test case returns the id to search along with the campaign and
	// error it wants from a call to GetCampaign
	tests := map[string]func(*testing.T) (int, *db.Campaign, error){
		"missing": func(t *testing.T) (int, *db.Campaign, error) {
			return 123, nil, sql.ErrNoRows
		},
		"expired recently": func(t *testing.T) (int, *db.Campaign, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(-7*24*time.Hour), time.Now().Add(-1*time.Second), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return campaign.ID, campaign, nil
		},
		"future": func(t *testing.T) (int, *db.Campaign, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(7*24*time.Hour), time.Now().Add(10*24*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return campaign.ID, campaign, nil
		},
		"active": func(t *testing.T) (int, *db.Campaign, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(-7*24*time.Hour), time.Now().Add(10*24*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return campaign.ID, campaign, nil
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			id, want, wantErr := setup(t)
			defer reset(t)
			campaign, err := db.GetCampaign(id)
			if err := campaignEq(campaign, want); err != nil {
				t.Errorf("GetCampaign() %v", err)
			}
			if err != wantErr {
				t.Fatalf("GetCampaign() err = %v; want %v", err, wantErr)
			}
		})
	}
}

func TestCreateOrder_valid(t *testing.T) {
	reset(t)

	tests := map[string]func(*testing.T) db.Order{
		"valid": func(t *testing.T) db.Order {
			campaign, err := db.CreateCampaign(time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour), 1000)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			return db.Order{
				CampaignID: campaign.ID,
				Customer:   testCustomer(),
				Address:    testAddress(),
				Payment:    testPayment(),
			}
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			want := setup(t)
			defer reset(t)

			nBefore := count(t, "orders")

			created := want
			err := db.CreateOrder(&created)
			if err != nil {
				t.Fatalf("CreateOrder() err = %v; want nil", err)
			}
			if created.ID <= 0 {
				t.Errorf("CreateOrder() ID = %d; want > 0", created.ID)
			}
			want.ID = created.ID
			if created != want {
				t.Errorf("CreateOrder() = %v; want %v", created, want)
			}

			nAfter := count(t, "orders")
			if diff := nAfter - nBefore; diff != 1 {
				t.Fatalf("CreateOrder() increased order count by %d; want %d", diff, 1)
			}

			got, err := db.GetOrderViaPayCus(want.Payment.CustomerID)
			if err != nil {
				t.Fatalf("GetOrderViaPayCus() err = %v; want nil", err)
			}
			if *got != want {
				t.Errorf("CreateOrder() = %v; want %v", *got, want)
			}
		})
	}
}

const (
	// Ideally our package would return a better error here, but since we
	// aren't refactoring right now I'm just trying to capture whatever the
	// current backend uses as best as I can.
	pqForeignKeyCode = "23503"
)

func TestCreateOrder_invalid(t *testing.T) {
	reset(t)
	type checkFn func(error) error

	checkPqError := func(code pq.ErrorCode) func(error) error {
		return func(err error) error {
			if err == nil {
				return fmt.Errorf("got nil; want *pq.Error")
			}
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code != code {
					return fmt.Errorf("pq.Error.Code = %s; want %s", pqErr.Code, code)
				}
				return nil
			}
			return fmt.Errorf("got %T; want *pq.Error", err)
		}
	}

	tests := map[string]func(*testing.T) (db.Order, []checkFn){
		"missing campaign": func(t *testing.T) (db.Order, []checkFn) {
			return db.Order{
				Customer: testCustomer(),
				Address:  testAddress(),
				Payment:  testPayment(),
			}, []checkFn{checkPqError(pqForeignKeyCode)}
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			order, checks := setup(t)
			defer reset(t)

			nBefore := count(t, "orders")
			created := order
			createErr := db.CreateOrder(&created)
			for _, check := range checks {
				if err := check(createErr); err != nil {
					t.Errorf("CreateOrder() %v", err)
				}
			}
			nAfter := count(t, "orders")
			if diff := nAfter - nBefore; diff != 0 {
				t.Fatalf("CreateOrder() increased order count by %d; want %d", diff, 0)
			}

			got, err := db.GetOrderViaPayCus(order.Payment.CustomerID)
			if err != sql.ErrNoRows {
				t.Fatalf("GetOrderViaPayCus() err = %v; want %v", err, sql.ErrNoRows)
			}
			if got != nil {
				t.Fatalf("GetOrderViaPayCus() = %v; want nil", got)
			}
		})
	}
}

func TestGetOrderViaPayCus(t *testing.T) {
	reset(t)

	// each test case returns the id to search along with the campaign and
	// error it wants from a call to GetCampaign
	tests := map[string]func(*testing.T) (string, *db.Order, error){
		"missing": func(t *testing.T) (string, *db.Order, error) {
			return "fake_id", nil, sql.ErrNoRows
		},
		"expired campaign": func(t *testing.T) (string, *db.Order, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(-7*24*time.Hour), time.Now().Add(-1*time.Second), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			order := db.Order{
				CampaignID: campaign.ID,
				Customer:   testCustomer(),
				Address:    testAddress(),
				Payment:    testPayment(),
			}
			order.Payment.CustomerID = "cus_123abc"
			err = db.CreateOrder(&order)
			if err != nil {
				t.Fatalf("CreateOrder() err = %v; want nil", err)
			}
			return order.Payment.CustomerID, &order, nil
		},
		"future campaign": func(t *testing.T) (string, *db.Order, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(7*24*time.Hour), time.Now().Add(10*24*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			order := db.Order{
				CampaignID: campaign.ID,
				Customer:   testCustomer(),
				Address:    testAddress(),
				Payment:    testPayment(),
			}
			order.Payment.CustomerID = "cus_888zzz"
			err = db.CreateOrder(&order)
			if err != nil {
				t.Fatalf("CreateOrder() err = %v; want nil", err)
			}
			return order.Payment.CustomerID, &order, nil
		},
		"active campaign": func(t *testing.T) (string, *db.Order, error) {
			campaign, err := db.CreateCampaign(time.Now().Add(-7*24*time.Hour), time.Now().Add(10*24*time.Hour), 900)
			if err != nil {
				t.Fatalf("CreateCampaign() err = %v; want nil", err)
			}
			order := db.Order{
				CampaignID: campaign.ID,
				Customer:   testCustomer(),
				Address:    testAddress(),
				Payment:    testPayment(),
			}
			order.Payment.CustomerID = "non_cus_prefixed_string"
			err = db.CreateOrder(&order)
			if err != nil {
				t.Fatalf("CreateOrder() err = %v; want nil", err)
			}
			return order.Payment.CustomerID, &order, nil
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			id, want, wantErr := setup(t)
			defer reset(t)
			order, err := db.GetOrderViaPayCus(id)
			if err != wantErr {
				t.Fatalf("GetOrderViaPayCus() err = %v; want %v", err, wantErr)
			}
			if order == nil && want == nil {
				return
			}
			if order == nil || want == nil {
				t.Fatalf("GetOrderViaPayCus() = %v; want %v", order, want)
			}
			if *order != *want {
				t.Fatalf("GetOrderViaPayCus() = %+v; want %+v", *order, *want)
			}
		})
	}
}

func campaignEq(got, want *db.Campaign) error {
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

func testCustomer() db.Customer {
	return db.Customer{
		Name:  "Johnny Test Person",
		Email: "johnny.test@gopherswag.com",
	}
}

func testAddress() db.Address {
	return db.Address{
		Street1: "123 Test St",
		Street2: "Apt 456",
		City:    "Beverly Hills",
		State:   "CA",
		Zip:     "90210",
		Country: "United States",
		Raw: `JOHNNY TEST PERSON
123 TEST ST
APT 456
BEVERLY HILLS CA  90210
UNITED STATES`,
	}
}

func testPayment() db.Payment {
	return db.Payment{
		Source:     "stripe",
		CustomerID: "cus_123abc",
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
