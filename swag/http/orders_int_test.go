//+build int

package http_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
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

func TestOrderHandler_Show_stripeInt(t *testing.T) {
	if stripeSecretKey == "" {
		t.Skip("stripe secret key not provided")
	}

	t.Run("charged", func(t *testing.T) {
		price := 1000

		tests := map[string]struct {
			chgID    func(*testing.T, *stripe.Client) string
			wantCode int
			wantBody string
		}{
			"succeeded": {
				chgID: func(t *testing.T, sc *stripe.Client) string {
					cus, err := sc.Customer("tok_visa", "success@gopherswag.com")
					if err != nil {
						t.Fatalf("Customer() err = %v; want %v", err, nil)
					}
					chg, err := sc.Charge(cus.ID, price)
					if err != nil {
						t.Fatalf("Charge() err = %v; want %v", err, nil)
					}
					return chg.ID
				},
				wantCode: http.StatusOK,
				wantBody: "Your order has been completed successfully!",
			},
			"error getting charge": {
				chgID: func(*testing.T, *stripe.Client) string {
					return "chg_fake_id"
				},
				// This should probably be changed long term to return an error
				// status code.
				wantCode: http.StatusOK,
				wantBody: "Failed to lookup the status of your order.",
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				oh := OrderHandler{}
				sc := &stripe.Client{
					Key: stripeSecretKey,
				}
				oh.Stripe.Client = sc
				campaign := &db.Campaign{
					ID:    999,
					Price: price,
				}
				chgID := tc.chgID(t, sc)
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
						ChargeID:   chgID,
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

func TestOrderHandler_Confirm_stripeInt(t *testing.T) {
	if stripeSecretKey == "" {
		t.Skip("stripe secret key not provided")
	}

	type checkFn func(*testing.T, *http.Response)
	hasStatus := func(code int) checkFn {
		return func(t *testing.T, res *http.Response) {
			if res.StatusCode != code {
				t.Fatalf("StatusCode = %d; want %d", res.StatusCode, code)
			}
		}
	}
	hasBody := func(want string) checkFn {
		return func(t *testing.T, res *http.Response) {
			defer res.Body.Close()
			bodyBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("ReadAll() err = %v; want %v", err, nil)
			}
			got := strings.TrimSpace(string(bodyBytes))
			if got != want {
				t.Fatalf("Body = %v; want %v", got, want)
			}
		}
	}
	hasLocation := func(want string) checkFn {
		return func(t *testing.T, res *http.Response) {
			locURL, err := res.Location()
			if err != nil {
				t.Fatalf("Location() err = %v; want %v", err, nil)
			}
			gotLoc := locURL.Path
			if gotLoc != want {
				t.Fatalf("Redirect location = %s; want %s", gotLoc, want)
			}
		}
	}
	testOrder := func(campaignID int, customerID string) *db.Order {
		return &db.Order{
			ID:         123,
			CampaignID: campaignID,
			Address: db.Address{
				Raw: `JON CALHOUN
PO BOX 295
BEDFORD PA  15522
UNITED STATES`,
			},
			Payment: db.Payment{
				CustomerID: customerID,
				Source:     "stripe",
			},
		}
	}

	runTest := func(t *testing.T, oh *OrderHandler, formData url.Values, order *db.Order, checks ...checkFn) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders/cus_abc123", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), "order", order))
		oh.Confirm(w, r)

		res := w.Result()
		for _, check := range checks {
			check(t, res)
		}
	}

	t.Run("stripe error when creating charge", func(t *testing.T) {
		oh := OrderHandler{}
		campaign := &db.Campaign{
			ID:    999,
			Price: 1000,
		}
		sc := &stripe.Client{
			Key: stripeSecretKey,
		}
		oh.Stripe.Client = sc
		stripeCus, err := sc.Customer("tok_chargeCustomerFail", "fail@gopherswag.com")
		if err != nil {
			t.Fatalf("Customer() err = %v; want %v", err, nil)
		}
		order := testOrder(campaign.ID, stripeCus.ID)
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				if id == campaign.ID {
					return campaign, nil
				}
				return nil, sql.ErrNoRows
			},
		}
		oh.DB = mdb

		formData := url.Values{
			"address-raw": []string{order.Address.Raw},
		}

		runTest(t, &oh, formData, order,
			hasStatus(http.StatusOK),
			hasBody("Your card was declined."),
		)
	})

	t.Run("visa card", func(t *testing.T) {
		oh := OrderHandler{}
		campaign := &db.Campaign{
			ID:    999,
			Price: 1000,
		}
		sc := &stripe.Client{
			Key: stripeSecretKey,
		}
		oh.Stripe.Client = sc
		stripeCus, err := sc.Customer("tok_visa", "same@gopherswag.com")
		if err != nil {
			t.Fatalf("Customer() err = %v; want %v", err, nil)
		}
		order := testOrder(campaign.ID, stripeCus.ID)
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				if id == campaign.ID {
					return campaign, nil
				}
				return nil, sql.ErrNoRows
			},
			ConfirmOrderFunc: func(gotOrderID int, gotAddress, gotChargeID string) error {
				return nil
			},
		}
		oh.DB = mdb

		formData := url.Values{
			"address-raw": []string{order.Address.Raw},
		}

		runTest(t, &oh, formData, order,
			hasStatus(http.StatusFound),
			hasLocation(fmt.Sprintf("/orders/%s", order.Payment.CustomerID)),
		)
	})

	t.Run("discover card", func(t *testing.T) {
		newAddress := `NEW ADDRESS HERE
123 NEW STREET
SOME TOWN NY  12345
UNITED STATES`
		oh := OrderHandler{}
		campaign := &db.Campaign{
			ID:    999,
			Price: 1000,
		}
		sc := &stripe.Client{
			Key: stripeSecretKey,
		}
		oh.Stripe.Client = sc
		stripeCus, err := sc.Customer("tok_discover", "diff@gopherswag.com")
		if err != nil {
			t.Fatalf("Customer() err = %v; want %v", err, nil)
		}
		order := testOrder(campaign.ID, stripeCus.ID)
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				if id == campaign.ID {
					return campaign, nil
				}
				return nil, sql.ErrNoRows
			},
			ConfirmOrderFunc: func(gotOrderID int, gotAddress, gotChargeID string) error {
				return nil
			},
		}
		oh.DB = mdb

		formData := url.Values{
			"address-raw": []string{newAddress},
		}

		runTest(t, &oh, formData, order,
			hasStatus(http.StatusFound),
			hasLocation(fmt.Sprintf("/orders/%s", order.Payment.CustomerID)),
		)
	})
}
