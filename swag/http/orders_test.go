package http_test

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
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

func TestOrderHandler_New(t *testing.T) {
	type checkFn func(*testing.T, *http.Response)
	checks := func(fns ...checkFn) []checkFn {
		return fns
	}
	hasBody := func(want string) func(*testing.T, *http.Response) {
		return func(t *testing.T, got *http.Response) {
			body, err := ioutil.ReadAll(got.Body)
			if err != nil {
				t.Fatalf("ReadAll() err = %v; want %v", err, nil)
			}
			gotBody := strings.TrimSpace(string(body))
			if gotBody != want {
				t.Fatalf("body = %v; want %v", gotBody, want)
			}
		}
	}
	hasStatus := func(code int) func(*testing.T, *http.Response) {
		return func(t *testing.T, got *http.Response) {
			if got.StatusCode != code {
				t.Fatalf("StatusCode = %d; want %d", got.StatusCode, code)
			}
		}
	}

	// each test case returns a campaign handler along with the expected
	// body output as a string
	tests := map[string]func(*testing.T) (*OrderHandler, *db.Campaign, []checkFn){
		"campaign id": func(t *testing.T) (*OrderHandler, *db.Campaign, []checkFn) {
			oh := OrderHandler{}
			oh.Templates.New = template.Must(template.New("").Parse("{{.Campaign.ID}}"))
			return &oh, &db.Campaign{
				ID: 123,
			}, checks(hasBody("123"))
		},
		"campaign price": func(t *testing.T) (*OrderHandler, *db.Campaign, []checkFn) {
			oh := OrderHandler{}
			oh.Templates.New = template.Must(template.New("").Parse("{{.Campaign.Price}}"))
			return &oh, &db.Campaign{
				Price: 1200,
			}, checks(hasBody("12"))
		},
		"stripe public key": func(t *testing.T) (*OrderHandler, *db.Campaign, []checkFn) {
			oh := OrderHandler{}
			oh.Stripe.PublicKey = "sk_pub_123abc"
			oh.Templates.New = template.Must(template.New("").Parse("{{.StripePublicKey}}"))
			return &oh, &db.Campaign{
				Price: 1200,
			}, checks(hasBody(oh.Stripe.PublicKey))
		},
		"campaign is not set": func(t *testing.T) (*OrderHandler, *db.Campaign, []checkFn) {
			oh := OrderHandler{}
			return &oh, nil, checks(hasBody("Campaign not provided."), hasStatus(http.StatusInternalServerError))
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			oh, campaign, checks := tc(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if campaign != nil {
				r = r.WithContext(context.WithValue(r.Context(), "campaign", campaign))
			}
			oh.New(w, r)
			res := w.Result()
			defer res.Body.Close()
			for _, check := range checks {
				check(t, res)
			}
		})
	}
}

func TestOrderHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
			"stripe-token": []string{"fake-stripe-token"},
		}
		stripeCustomerID := "cus_abc123"
		oh.Stripe.Client = &mockStripe{
			CustomerFunc: func(token, email string) (*stripe.Customer, error) {
				if token != formData.Get("stripe-token") {
					t.Fatalf("token = %s; want %s", token, formData.Get("stripe-token"))
				}
				if email != formData.Get("Email") {
					t.Fatalf("email = %s; want %s", email, formData.Get("Email"))
				}
				return &stripe.Customer{
					ID: stripeCustomerID,
				}, nil
			},
		}

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
		wantLoc := fmt.Sprintf("/orders/%s", stripeCustomerID)
		if gotLoc != wantLoc {
			t.Fatalf("Redirect location = %s; want %s", gotLoc, wantLoc)
		}
	})
}

func TestOrderHandler_OrderMw(t *testing.T) {
	failHandler := func(t *testing.T) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("next handler shouldn't have been called by middleware")
		}
	}
	// Three paths:
	// 1. Can't find order or DB error
	// 2. Success
	t.Run("missing order", func(t *testing.T) {
		oh := OrderHandler{}
		mdb := &mockDB{
			GetOrderViaPayCusFunc: func(id string) (*db.Order, error) {
				return nil, sql.ErrNoRows
			},
		}
		oh.DB = mdb
		handler := oh.OrderMw(failHandler(t))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/cus_abc123/id/here", nil)
		handler(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("order found", func(t *testing.T) {
		order := &db.Order{
			ID: 123,
			Payment: db.Payment{
				CustomerID: "cus_abc123",
				Source:     "stripe",
			},
		}
		oh := OrderHandler{}
		mdb := &mockDB{
			GetOrderViaPayCusFunc: func(id string) (*db.Order, error) {
				if id == order.Payment.CustomerID {
					return order, nil
				}
				return nil, sql.ErrNoRows
			},
		}
		oh.DB = mdb
		handlerCalled := false
		handler := oh.OrderMw(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			if r.URL.Path != "/remaining/path/" {
				t.Fatalf("Path in next handler = %v; want %v", r.URL.Path, "/remaining/path/")
			}
			got := r.Context().Value("order").(*db.Order)
			if got != order {
				t.Fatalf("Order = %v; want %v", got, order)
			}
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/remaining/path", order.Payment.CustomerID), nil)
		handler(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusOK)
		}
		if !handlerCalled {
			t.Fatalf("next handler not called")
		}
	})
}
