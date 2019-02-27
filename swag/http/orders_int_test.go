//+build int

package http_test

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/stripe"
	"github.com/joncalhoun/twg/swag/db"
	. "github.com/joncalhoun/twg/swag/http"
)

var (
	stripeSecretKey = ""
)

func init() {
	flag.StringVar(&stripeSecretKey, "stripe", "", "stripe secret key for integration testing")
	flag.Parse()
}

func TestOrderHandler_Create_stripeInt(t *testing.T) {
	if stripeSecretKey == "" {
		t.Skip("stripe secret key not provided")
	}
	t.Run("visa", func(t *testing.T) {
		token := "tok_visa"
		oh := OrderHandler{}
		oh.DB = &mockDB{
			CreateOrderFunc: func(order *db.Order) error {
				order.ID = 123
				return nil
			},
		}
		formData := url.Values{
			"Name":         []string{"Jon Calhoun"},
			"Email":        []string{"jon@calhoun.io"},
			"stripe-token": []string{token},
		}
		stripeCustomerID := ""
		stripeClient := &stripe.Client{
			Key: stripeSecretKey,
		}
		oh.Stripe.Client = &mockStripe{
			CustomerFunc: func(token, email string) (*stripe.Customer, error) {
				cus, err := stripeClient.Customer(token, email)
				if cus != nil {
					stripeCustomerID = cus.ID
				}
				return cus, err
			},
		}

		oh.Logger = &logFail{t}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), "campaign", &db.Campaign{
			ID: 333,
		}))
		oh.Create(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusFound {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusFound)
		}
		locURL, err := res.Location()
		if err != nil {
			t.Fatalf("Location() err = %v; want %v", err, nil)
		}
		gotLoc := locURL.Path
		if !strings.HasPrefix(gotLoc, "/orders/") {
			t.Fatalf("Redirect location = %s; want prefix %s", gotLoc, "/orders/")
		}
		gotCustomerID := gotLoc[len("/orders/"):]
		if gotCustomerID != stripeCustomerID {
			t.Fatalf("Customer ID = %s; want %s", gotCustomerID, stripeCustomerID)
		}
	})
}
