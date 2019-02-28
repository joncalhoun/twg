//+build int

package http_test

import (
	"context"
	"flag"
	"io/ioutil"
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
		stripeClient := &stripe.Client{
			Key: stripeSecretKey,
		}
		stripeCustomerID := ""
		oh.Stripe.Client = &mockStripe{
			CustomerFunc: func(email, token string) (*stripe.Customer, error) {
				cus, err := stripeClient.Customer(email, token)
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
		wantLocPre := "/orders/"
		if !strings.HasPrefix(gotLoc, wantLocPre) {
			t.Fatalf("Redirect location = %s; want prefix %s", gotLoc, wantLocPre)
		}

		gotStripeCusID := gotLoc[len("/orders/"):]
		if gotStripeCusID != stripeCustomerID {
			t.Fatalf("Stripe Customer ID = %s; want %s", gotStripeCusID, stripeCustomerID)
		}
	})

	t.Run("cvc check fail", func(t *testing.T) {
		token := "tok_cvcCheckFail"
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
		stripeClient := &stripe.Client{
			Key: stripeSecretKey,
		}
		oh.Stripe.Client = stripeClient
		oh.Logger = &logRec{}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), "campaign", &db.Campaign{
			ID: 333,
		}))
		oh.Create(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusInternalServerError)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ReadAll() err = %v; want %v", err, nil)
		}
		gotBody := strings.TrimSpace(string(body))
		wantBody := "Something went wrong processing your payment information."
		if !strings.Contains(gotBody, wantBody) {
			t.Fatalf("Body = %v; want substring %v", gotBody, wantBody)
		}
	})
}
