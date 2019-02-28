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
	type checkFn func(*testing.T, *http.Response)
	hasCode := func(want int) checkFn {
		return func(t *testing.T, res *http.Response) {
			if res.StatusCode != want {
				t.Fatalf("StatusCode = %d; want %d", res.StatusCode, want)
			}
		}
	}
	bodyContains := func(want string) checkFn {
		return func(t *testing.T, res *http.Response) {
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("ReadAll() err = %v; want %v", err, nil)
			}
			gotBody := strings.TrimSpace(string(body))
			if !strings.Contains(gotBody, want) {
				t.Fatalf("Body = %v; want substring %v", gotBody, want)
			}
		}
	}
	hasLocationPrefix := func(want string) checkFn {
		return func(t *testing.T, res *http.Response) {
			locURL, err := res.Location()
			if err != nil {
				t.Fatalf("Location() err = %v; want %v", err, nil)
			}
			gotLoc := locURL.Path
			if !strings.HasPrefix(gotLoc, want) {
				t.Fatalf("Redirect location = %s; want prefix %s", gotLoc, want)
			}
		}
	}
	hasCustomerID := func(customerID *string) checkFn {
		return func(t *testing.T, res *http.Response) {
			locURL, err := res.Location()
			if err != nil {
				t.Fatalf("Location() err = %v; want %v", err, nil)
			}
			gotLoc := locURL.Path
			gotStripeCusID := gotLoc[len("/orders/"):]
			stripeCustomerID := *customerID
			if gotStripeCusID != stripeCustomerID {
				t.Fatalf("Stripe Customer ID = %s; want %s", gotStripeCusID, stripeCustomerID)
			}
		}
	}
	hasLogs := func(logger *logRec, logs ...string) checkFn {
		return func(t *testing.T, _ *http.Response) {
			if len(logger.logs) != len(logs) {
				t.Fatalf("len(logs) = %d; want %d", len(logger.logs), len(logs))
			}
			for i, log := range logs {
				if !strings.HasPrefix(logger.logs[i], log) {
					t.Fatalf("log[%d] = %s; want prefix %s", i, logger.logs[i], log)
				}
			}
		}
	}
	stripeClientAndIDCapture := func(stripeClient interface {
		Customer(email, token string) (*stripe.Customer, error)
	}) (*mockStripe, *string) {
		stripeCustomerID := ""
		return &mockStripe{
			CustomerFunc: func(email, token string) (*stripe.Customer, error) {
				cus, err := stripeClient.Customer(email, token)
				if cus != nil {
					stripeCustomerID = cus.ID
				}
				return cus, err
			},
		}, &stripeCustomerID
	}

	tests := map[string]func(*testing.T, *OrderHandler) (string, []checkFn){
		"visa": func(t *testing.T, oh *OrderHandler) (string, []checkFn) {
			stripeClient, stripeCustomerID := stripeClientAndIDCapture(oh.Stripe.Client)
			oh.Stripe.Client = stripeClient
			oh.Logger = &logFail{t}

			return "tok_visa", []checkFn{
				hasCode(http.StatusFound),
				hasLocationPrefix("/orders/"),
				hasCustomerID(stripeCustomerID),
			}
		},
		"cvc check failure": func(t *testing.T, oh *OrderHandler) (string, []checkFn) {
			lr := &logRec{}
			oh.Logger = lr

			return "tok_cvcCheckFail", []checkFn{
				hasCode(http.StatusInternalServerError),
				bodyContains("Something went wrong processing your payment information."),
				hasLogs(lr, "Error creating stripe customer."),
			}
		},
		"amex": func(t *testing.T, oh *OrderHandler) (string, []checkFn) {
			stripeClient, stripeCustomerID := stripeClientAndIDCapture(oh.Stripe.Client)
			oh.Stripe.Client = stripeClient
			oh.Logger = &logFail{t}

			return "tok_amex", []checkFn{
				hasCode(http.StatusFound),
				hasLocationPrefix("/orders/"),
				hasCustomerID(stripeCustomerID),
			}
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			oh := OrderHandler{}
			oh.DB = &mockDB{
				CreateOrderFunc: func(order *db.Order) error {
					order.ID = 123
					return nil
				},
			}
			oh.Stripe.Client = &stripe.Client{
				Key: stripeSecretKey,
			}
			oh.Logger = &logRec{}

			token, checks := tc(t, &oh)

			formData := url.Values{
				"Name":         []string{"Jon Calhoun"},
				"Email":        []string{"jon@calhoun.io"},
				"stripe-token": []string{token},
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(formData.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r = r.WithContext(context.WithValue(r.Context(), "campaign", &db.Campaign{
				ID: 333,
			}))
			oh.Create(w, r)
			res := w.Result()
			for _, check := range checks {
				check(t, res)
			}
		})
	}
}
