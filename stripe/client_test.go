package stripe_test

import (
	"flag"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/stripe"
)

var (
	apiKey string
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the Stripe API. If present, integration tests will be run using this key.")
}

func TestClient_Customer(t *testing.T) {
	if apiKey == "" {
		t.Skip("No API key provided")
	}

	c := stripe.Client{
		Key: apiKey,
	}
	tok := "tok_amex"
	email := "test@testwithgo.com"
	cus, err := c.Customer(tok, email)
	if err != nil {
		t.Errorf("Customer() err = %v; want %v", err, nil)
	}
	if cus == nil {
		t.Fatalf("Customer() = nil; want non-nil value")
	}
	if !strings.HasPrefix(cus.ID, "cus_") {
		t.Errorf("Customer() ID = %s; want prefix %q", cus.ID, "cus_")
	}
	if !strings.HasPrefix(cus.DefaultSource, "card_") {
		t.Errorf("Customer() DefaultSource = %s; want prefix %q", cus.DefaultSource, "card_")
	}
	if cus.Email != email {
		t.Errorf("Customer() Email = %s; want %s", cus.Email, email)
	}
}

func TestClient_Charge(t *testing.T) {
	if apiKey == "" {
		t.Skip("No API key provided")
	}

	type checkFn func(*testing.T, *stripe.Charge, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, charge *stripe.Charge, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}
	hasAmount := func(amount int) checkFn {
		return func(t *testing.T, charge *stripe.Charge, err error) {
			if charge.Amount != amount {
				t.Errorf("Amount = %d; want %d", charge.Amount, amount)
			}
		}
	}
	hasErrType := func(typee string) checkFn {
		return func(t *testing.T, charge *stripe.Charge, err error) {
			se, ok := err.(stripe.Error)
			if !ok {
				t.Fatalf("err isn't a stripe.Error")
			}
			if se.Type != typee {
				t.Errorf("err.Type = %s; want %s", se.Type, typee)
			}
		}
	}

	c := stripe.Client{
		Key: apiKey,
	}
	// Create a customer for the test
	tok := "tok_amex"
	email := "test@testwithgo.com"
	cus, err := c.Customer(tok, email)
	if err != nil {
		t.Fatalf("Customer() err = %v; want nil", err)
	}

	tests := map[string]struct {
		customerID string
		amount     int
		checks     []checkFn
	}{
		"valid charge": {
			customerID: cus.ID,
			amount:     1234,
			checks:     check(hasNoErr(), hasAmount(1234)),
		},
		"invalid customer id": {
			customerID: "cus_missing",
			amount:     1234,
			checks:     check(hasErrType("invalid_request_error")),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			charge, err := c.Charge(tc.customerID, tc.amount)
			for _, check := range tc.checks {
				check(t, charge, err)
			}
		})
	}
}
