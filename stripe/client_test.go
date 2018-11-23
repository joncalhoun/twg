package stripe_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/stripe"
)

var (
	apiKey string
	update bool
)

const (
	tokenAmex               = "tok_amex"
	tokenVisaDebit          = "tok_visa_debit"
	tokenMastercardPrepaid  = "tok_mastercard_prepaid"
	tokenInvalid            = "tok_alsdkjfa"
	tokenExpiredCard        = "tok_chargeDeclinedExpiredCard"
	tokenIncorrectCVC       = "tok_chargeDeclinedIncorrectCvc"
	tokenInsufficientFunds  = "tok_chargeDeclinedInsufficientFunds"
	tokenChargeCustomerFail = "tok_chargeCustomerFail"
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the Stripe API. If present, integration tests will be run using this key.")
	flag.BoolVar(&update, "update", false, "Set this flag to update the responses used in local tests. This requires that the key flag is set so that we can interact with the Stripe API.")
}

func TestClient_Local(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, `{
			"id": "cus_Dy08A71mTUtap1",
			"object": "customer",
			"account_balance": 0,
			"created": 1542124116,
			"currency": "usd",
			"default_source": null,
			"delinquent": false,
			"description": null,
			"discount": null,
			"email": null,
			"invoice_prefix": "1C04993",
			"livemode": false,
			"metadata": {
			},
			"shipping": null,
			"sources": {
				"object": "list",
				"data": [

				],
				"has_more": false,
				"total_count": 0,
				"url": "/v1/customers/cus_Dy08A71mTUtap1/sources"
			},
			"subscriptions": {
				"object": "list",
				"data": [

				],
				"has_more": false,
				"total_count": 0,
				"url": "/v1/customers/cus_Dy08A71mTUtap1/subscriptions"
			},
			"tax_info": null,
			"tax_info_verification": null
		}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	c := stripe.Client{
		Key:     "gibberish-key",
		BaseURL: server.URL,
	}
	_, err := c.Customer("random token", "random email")
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func stripeClient(t *testing.T) (*stripe.Client, func()) {
	teardown := make([]func(), 0)
	c := stripe.Client{
		Key: apiKey,
	}
	if apiKey == "" {
		count := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			resp := readResponse(t, count)
			w.WriteHeader(resp.StatusCode)
			w.Write(resp.Body)
			count++
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		c.BaseURL = server.URL
		teardown = append(teardown, server.Close)
	}
	if update {
		rc := &recorderClient{}
		c.HttpClient = rc
		teardown = append(teardown, func() {
			for i, res := range rc.responses {
				recordResponse(t, res, i)
			}
		})
	}
	return &c, func() {
		for _, fn := range teardown {
			fn()
		}
	}
}

func responsePath(t *testing.T, count int) string {
	return filepath.Join("testdata", filepath.FromSlash(fmt.Sprintf("%s.%d.json", t.Name(), count)))
}

func readResponse(t *testing.T, count int) response {
	var resp response
	path := responsePath(t, count)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read the response file: %s. err = %v", path, err)
	}
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		t.Fatalf("failed to json unmarshal the response file: %s. err = %v", path, err)
	}
	return resp
}

func recordResponse(t *testing.T, resp response, count int) {
	path := responsePath(t, count)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		t.Fatalf("failed to create the response dir: %s. err = %v", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON for response file: %s. err = %v", path, err)
	}
	_, err = f.Write(jsonBytes)
	if err != nil {
		t.Fatalf("failed to write json bytes for response file: %s. err = %v", path, err)
	}
}

func TestClient_Customer(t *testing.T) {
	if apiKey == "" {
		t.Log("No API key provided. Running unit tests using recorded responses. Be sure to run against the real API before commiting.")
	}

	type checkFn func(*testing.T, *stripe.Customer, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, cus *stripe.Customer, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}
	hasErrType := func(typee string) checkFn {
		return func(t *testing.T, cus *stripe.Customer, err error) {
			se, ok := err.(stripe.Error)
			if !ok {
				t.Fatalf("err isn't a stripe.Error")
			}
			if se.Type != typee {
				t.Errorf("err.Type = %s; want %s", se.Type, typee)
			}
		}
	}
	hasIDPrefix := func() checkFn {
		return func(t *testing.T, cus *stripe.Customer, err error) {
			if !strings.HasPrefix(cus.ID, "cus_") {
				t.Errorf("ID = %s; want prefix %q", cus.ID, "cus_")
			}
		}
	}
	hasCardDefaultSource := func() checkFn {
		return func(t *testing.T, cus *stripe.Customer, err error) {
			if !strings.HasPrefix(cus.DefaultSource, "card_") {
				t.Errorf("DefaultSource = %s; want prefix %q", cus.DefaultSource, "card_")
			}
		}
	}
	hasEmail := func(email string) checkFn {
		return func(t *testing.T, cus *stripe.Customer, err error) {
			if cus.Email != email {
				t.Errorf("Email = %s; want %s", cus.Email, email)
			}
		}
	}

	tests := map[string]struct {
		token  string
		email  string
		checks []checkFn
	}{
		"valid customer with amex": {
			token:  tokenAmex,
			email:  "test@testwithgo.com",
			checks: check(hasNoErr(), hasIDPrefix(), hasCardDefaultSource(), hasEmail("test@testwithgo.com")),
		},
		"invalid token": {
			token:  tokenInvalid,
			email:  "test@testwithgo.com",
			checks: check(hasErrType(stripe.ErrTypeInvalidRequest)),
		},
		"expired card": {
			token:  tokenExpiredCard,
			email:  "test@testwithgo.com",
			checks: check(hasErrType(stripe.ErrTypeCardError)),
		},
		"incorrect cvc": {
			token:  tokenIncorrectCVC,
			email:  "test@testwithgo.com",
			checks: check(hasErrType(stripe.ErrTypeCardError)),
		},
		"insufficient funds": {
			token:  tokenInsufficientFunds,
			email:  "test@testwithgo.com",
			checks: check(hasErrType(stripe.ErrTypeCardError)),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c, teardown := stripeClient(t)
			defer teardown()
			cus, err := c.Customer(tc.token, tc.email)
			for _, check := range tc.checks {
				check(t, cus, err)
			}
		})
	}
}

func TestClient_CustomerMeta(t *testing.T) {
	if apiKey == "" {
		t.Log("No API key provided. Running unit tests using recorded responses. Be sure to run against the real API before commiting.")
	}

	tests := map[string]struct {
		token string
		email string
	}{
		"valid customer with amex": {
			token: tokenAmex,
			email: "test@testwithgo.com",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c, teardown := stripeClient(t)
			defer teardown()
			cus, err := c.Customer(tc.token, tc.email)
			if err != nil {
				t.Fatalf("Customer() err = %v; want nil", err)
			}
			if len(cus.Meta) != 0 {
				t.Fatalf("len(meta) = %d; want 0", len(cus.Meta))
			}
			address := `MR PERSON
123 FAKE ST
SOME TOWN, CA 12345
UNITED STATES`
			cus, err = c.CustomerMeta(cus.ID, map[string]string{
				"address": address,
			})
			if err != nil {
				t.Fatalf("CustomerMeta() err = %v; want nil", err)
			}
			if len(cus.Meta) != 1 {
				t.Fatalf("len(meta) = %d; want 1", len(cus.Meta))
			}
			if cus.Meta["address"] != address {
				t.Fatalf("meta[address] = %q; want %q", cus.Meta["address"], address)
			}
		})
	}
}

func TestClient_Charge(t *testing.T) {
	if apiKey == "" {
		t.Log("No API key provided. Running unit tests using recorded responses. Be sure to run against the real API before commiting.")
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
	hasStatus := func(status string) checkFn {
		return func(t *testing.T, charge *stripe.Charge, err error) {
			if charge.Status != status {
				t.Errorf("Status = %s; want %s", charge.Status, status)
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
	hasMeta := func(key, value string) checkFn {
		return func(t *testing.T, charge *stripe.Charge, err error) {
			if charge.Meta[key] != value {
				t.Errorf("metadata[%s] = %q; want %q", key, charge.Meta[key], value)
			}
		}
	}

	customerViaToken := func(token string) func(*testing.T, *stripe.Client) string {
		return func(t *testing.T, c *stripe.Client) string {
			email := "test@testwithgo.com"
			cus, err := c.Customer(token, email)
			if err != nil {
				t.Fatalf("err creating customer with token %s. err = %v; want nil", token, err)
			}
			return cus.ID
		}
	}

	chargeViaToken := func(token string, amount int) func(*testing.T, *stripe.Client) string {
		return func(t *testing.T, c *stripe.Client) string {
			email := "test@testwithgo.com"
			cus, err := c.Customer(token, email)
			if err != nil {
				t.Fatalf("err creating customer with token %s. err = %v; want nil", token, err)
			}
			chg, err := c.Charge(cus.ID, amount, nil)
			if err != nil {
				t.Fatalf("err creating charge with token %s. err = %v; want nil", token, err)
			}
			return chg.ID
		}
	}

	t.Run("Create charge", func(t *testing.T) {
		tests := map[string]struct {
			customerID func(*testing.T, *stripe.Client) string
			amount     int
			meta       map[string]string
			checks     []checkFn
		}{
			"valid charge with amex": {
				customerID: customerViaToken(tokenAmex),
				amount:     1234,
				checks:     check(hasNoErr(), hasAmount(1234)),
			},
			"valid charge with visa debit": {
				customerID: customerViaToken(tokenVisaDebit),
				amount:     8787,
				checks:     check(hasNoErr(), hasAmount(8787)),
			},
			"valid charge with mastercard prepaid": {
				customerID: customerViaToken(tokenMastercardPrepaid),
				amount:     98765,
				checks:     check(hasNoErr(), hasAmount(98765)),
			},
			"invalid customer id": {
				customerID: func(*testing.T, *stripe.Client) string {
					return "cus_missing"
				},
				amount: 1234,
				checks: check(hasErrType(stripe.ErrTypeInvalidRequest)),
			},
			"charge failure": {
				customerID: customerViaToken(tokenChargeCustomerFail),
				amount:     5555,
				checks:     check(hasErrType(stripe.ErrTypeCardError)),
			},
			"valid charge with metadata": {
				customerID: customerViaToken(tokenAmex),
				amount:     888,
				meta: map[string]string{
					"address": `123 Easy St
Some Town, CA  90210
UNITED STATES`,
				},
				checks: check(hasNoErr(), hasAmount(888), hasMeta("address", `123 Easy St
Some Town, CA  90210
UNITED STATES`)),
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				c, teardown := stripeClient(t)
				defer teardown()
				cusID := tc.customerID(t, c)
				charge, err := c.Charge(cusID, tc.amount, tc.meta)
				for _, check := range tc.checks {
					check(t, charge, err)
				}
			})
		}
	})

	t.Run("Get charge", func(t *testing.T) {
		tests := map[string]struct {
			chargeID func(*testing.T, *stripe.Client) string
			checks   []checkFn
		}{
			"successful charge": {
				chargeID: chargeViaToken(tokenAmex, 1234),
				checks:   check(hasNoErr(), hasAmount(1234), hasStatus("succeeded")),
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				c, teardown := stripeClient(t)
				defer teardown()
				chgID := tc.chargeID(t, c)
				charge, err := c.GetCharge(chgID)
				for _, check := range tc.checks {
					check(t, charge, err)
				}
			})
		}
	})
}
