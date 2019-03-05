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

func testOrderHandler_Show_review(t *testing.T, oh *OrderHandler, campaign *db.Campaign, order *db.Order) {
	tests := map[string]struct {
		tpl  *template.Template
		want func(*db.Order, *db.Campaign) string
	}{
		"order id": {
			tpl: template.Must(template.New("").Parse("{{.Order.ID}}")),
			want: func(order *db.Order, _ *db.Campaign) string {
				return order.Payment.CustomerID
			},
		},
		"order address": {
			tpl: template.Must(template.New("").Parse("{{.Order.Address}}")),
			want: func(order *db.Order, _ *db.Campaign) string {
				return order.Address.Raw
			},
		},
		"campaign price": {
			tpl: template.Must(template.New("").Parse("{{.Campaign.Price}}")),
			want: func(_ *db.Order, campaign *db.Campaign) string {
				return fmt.Sprintf("%d", campaign.Price/100)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			oh.Templates.Review = tc.tpl

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", nil)
			r = r.WithContext(context.WithValue(r.Context(), "order", order))
			oh.Show(w, r)
			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusOK)
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("ReadAll() err = %v; want %v", err, nil)
			}
			gotBody := string(body)
			wantBody := tc.want(order, campaign)
			if gotBody != wantBody {
				t.Fatalf("Body = %v; want %v", gotBody, wantBody)
			}
		})
	}
}

// Another way to run tests similar to the ones in TestOrderHandler_Show
func TestOrderHandler_Show_tableDemo(t *testing.T) {
	tests := map[string]func(*testing.T, *OrderHandler, *db.Campaign, *db.Order){
		"review - campaign found": testOrderHandler_Show_review,
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			oh := &OrderHandler{}
			campaign := &db.Campaign{
				ID:    999,
				Price: 1000,
			}
			order := &db.Order{
				ID:         123,
				CampaignID: campaign.ID,
				Address: db.Address{
					Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
				},
				Payment: db.Payment{
					CustomerID: "cus_abc123",
					Source:     "stripe",
				},
			}
			mdb := &mockDB{
				GetCampaignFunc: func(id int) (*db.Campaign, error) {
					if id == campaign.ID {
						return campaign, nil
					}
					return nil, sql.ErrNoRows
				},
			}
			oh.DB = mdb
			oh.Logger = &logRec{}

			tc(t, oh, campaign, order)
		})
	}
}

func TestOrderHandler_Show(t *testing.T) {
	t.Run("review - campaign found", func(t *testing.T) {
		tests := map[string]struct {
			tpl  *template.Template
			want func(*db.Order, *db.Campaign) string
		}{
			"order id": {
				tpl: template.Must(template.New("").Parse("{{.Order.ID}}")),
				want: func(order *db.Order, _ *db.Campaign) string {
					return order.Payment.CustomerID
				},
			},
			"order address": {
				tpl: template.Must(template.New("").Parse("{{.Order.Address}}")),
				want: func(order *db.Order, _ *db.Campaign) string {
					return order.Address.Raw
				},
			},
			"campaign price": {
				tpl: template.Must(template.New("").Parse("{{.Campaign.Price}}")),
				want: func(_ *db.Order, campaign *db.Campaign) string {
					return fmt.Sprintf("%d", campaign.Price/100)
				},
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				oh := OrderHandler{}
				campaign := &db.Campaign{
					ID:    999,
					Price: 1000,
				}
				order := &db.Order{
					ID:         123,
					CampaignID: campaign.ID,
					Address: db.Address{
						Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
					},
					Payment: db.Payment{
						CustomerID: "cus_abc123",
						Source:     "stripe",
					},
				}
				mdb := &mockDB{
					GetCampaignFunc: func(id int) (*db.Campaign, error) {
						if id == campaign.ID {
							return campaign, nil
						}
						return nil, sql.ErrNoRows
					},
				}
				oh.DB = mdb
				oh.Templates.Review = tc.tpl

				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", nil)
				r = r.WithContext(context.WithValue(r.Context(), "order", order))
				oh.Show(w, r)
				res := w.Result()
				if res.StatusCode != http.StatusOK {
					t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusOK)
				}
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("ReadAll() err = %v; want %v", err, nil)
				}
				gotBody := string(body)
				wantBody := tc.want(order, campaign)
				if gotBody != wantBody {
					t.Fatalf("Body = %v; want %v", gotBody, wantBody)
				}
			})
		}
	})

	t.Run("review - db error", func(t *testing.T) {
		oh := OrderHandler{}
		order := &db.Order{
			ID:         123,
			CampaignID: 999,
		}
		lr := &logRec{}
		oh.Logger = lr
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				return nil, sql.ErrNoRows
			},
		}
		oh.DB = mdb
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", nil)
		r = r.WithContext(context.WithValue(r.Context(), "order", order))
		oh.Show(w, r)
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
		wantBody := "Something went wrong..."
		if gotBody != wantBody {
			t.Fatalf("Body = %v; want %v", gotBody, wantBody)
		}
		if len(lr.logs) != 1 {
			t.Fatalf("len(logs) = %d; want %d", len(lr.logs), 1)
		}
	})

	t.Run("charged", func(t *testing.T) {
		tests := map[string]struct {
			stripeChg *stripe.Charge
			stripeErr error
			wantCode  int
			wantBody  string
		}{
			"succeeded": {
				stripeChg: &stripe.Charge{
					Status: "succeeded",
				},
				stripeErr: nil,
				wantCode:  http.StatusOK,
				wantBody:  "Your order has been completed successfully!",
			},
			"pending": {
				stripeChg: &stripe.Charge{
					Status: "pending",
				},
				stripeErr: nil,
				wantCode:  http.StatusOK,
				wantBody:  "Your payment is still pending.",
			},
			"failed": {
				stripeChg: &stripe.Charge{
					Status: "failed",
				},
				stripeErr: nil,
				wantCode:  http.StatusOK,
				wantBody:  "Your payment failed.",
			},
			"error getting charge": {
				stripeChg: nil,
				stripeErr: &stripe.Error{},
				// This should probably be changed long term to return an error
				// status code.
				wantCode: http.StatusOK,
				wantBody: "Failed to lookup the status of your order.",
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				oh := OrderHandler{}
				campaign := &db.Campaign{
					ID:    999,
					Price: 1000,
				}
				order := &db.Order{
					ID:         123,
					CampaignID: campaign.ID,
					Address: db.Address{
						Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
					},
					Payment: db.Payment{
						ChargeID:   "chg_xyz890",
						CustomerID: "cus_abc123",
						Source:     "stripe",
					},
				}
				mdb := &mockDB{
					GetCampaignFunc: func(id int) (*db.Campaign, error) {
						if id == campaign.ID {
							return campaign, nil
						}
						return nil, sql.ErrNoRows
					},
				}
				oh.DB = mdb
				oh.Stripe.Client = &mockStripe{
					GetChargeFunc: func(id string) (*stripe.Charge, error) {
						return tc.stripeChg, tc.stripeErr
					},
				}
				oh.Logger = &logRec{}

				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", nil)
				r = r.WithContext(context.WithValue(r.Context(), "order", order))
				oh.Show(w, r)
				res := w.Result()
				if res.StatusCode != tc.wantCode {
					t.Fatalf("StatusCode = %d; want %d", res.StatusCode, tc.wantCode)
				}
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("ReadAll() err = %v; want %v", err, nil)
				}
				gotBody := string(body)
				if !strings.Contains(gotBody, tc.wantBody) {
					t.Fatalf("Body = %v; want substring %v", gotBody, tc.wantBody)
				}

			})
		}
	})
}

func TestOrderHandler_Confirm(t *testing.T) {
	// Cases:
	// 1. Error getting campaign
	t.Run("error getting campaign", func(t *testing.T) {
		oh := OrderHandler{}
		lr := &logRec{}
		oh.Logger = lr
		campaign := &db.Campaign{
			ID:    999,
			Price: 1000,
		}
		order := &db.Order{
			ID:         123,
			CampaignID: campaign.ID,
			Address: db.Address{
				Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
			},
			Payment: db.Payment{
				CustomerID: "cus_abc123",
				Source:     "stripe",
			},
		}
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				return nil, sql.ErrNoRows
			},
		}
		oh.DB = mdb

		formData := url.Values{
			"address-raw": []string{order.Address.Raw},
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), "order", order))
		oh.Confirm(w, r)

		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusInternalServerError)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ReadAll() err = %v; want %v", err, nil)
		}
		got := strings.TrimSpace(string(body))
		want := "Something went wrong..."
		if got != want {
			t.Fatalf("Body = %v; want %v", got, want)
		}
		wantLog := "error retrieving order campaign\n"
		if len(lr.logs) != 1 || lr.logs[0] != wantLog {
			t.Fatalf("Logs = %v; want %v", lr.logs, []string{wantLog})
		}
	})
	// 2. Errors creating stripe charge - possibly many, and we should int test this
	// 3. Error confirming in DB
	// 4. All good - same address & new address
	t.Run("same address", func(t *testing.T) {
		paymentChargeID := "chg_123456"
		oh := OrderHandler{}
		campaign := &db.Campaign{
			ID:    999,
			Price: 1000,
		}
		order := &db.Order{
			ID:         123,
			CampaignID: campaign.ID,
			Address: db.Address{
				Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
			},
			Payment: db.Payment{
				CustomerID: "cus_abc123",
				Source:     "stripe",
			},
		}
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				if id == campaign.ID {
					return campaign, nil
				}
				return nil, sql.ErrNoRows
			},
			ConfirmOrderFunc: func(gotOrderID int, gotAddress, gotChargeID string) error {
				if gotOrderID != order.ID {
					t.Fatalf("ConfirmOrder() ID = %d; want %d", gotOrderID, order.ID)
				}
				if gotAddress != order.Address.Raw {
					t.Fatalf("ConfirmOrder() Address = %q; want %q", gotAddress, order.Address.Raw)
				}
				if gotChargeID != paymentChargeID {
					t.Fatalf("ConfirmOrder() ChargeID = %v; want %v", gotChargeID, paymentChargeID)
				}
				return nil
			},
		}
		oh.DB = mdb
		sc := &mockStripe{
			ChargeFunc: func(customerID string, amount int) (*stripe.Charge, error) {
				if customerID == order.Payment.CustomerID {
					return &stripe.Charge{
						ID: paymentChargeID,
					}, nil
				}
				return nil, stripe.Error{}
			},
		}
		oh.Stripe.Client = sc

		formData := url.Values{
			"address-raw": []string{order.Address.Raw},
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/orders/cus_abc123", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), "order", order))
		oh.Confirm(w, r)

		res := w.Result()
		if res.StatusCode != http.StatusFound {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusFound)
		}
		locURL, err := res.Location()
		if err != nil {
			t.Fatalf("Location() err = %v; want %v", err, nil)
		}
		gotLoc := locURL.Path
		wantLoc := fmt.Sprintf("/orders/%s", order.Payment.CustomerID)
		if gotLoc != wantLoc {
			t.Fatalf("Redirect location = %s; want %s", gotLoc, wantLoc)
		}
	})
}
