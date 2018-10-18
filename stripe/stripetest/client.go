package stripetest

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joncalhoun/twg/stripe"
)

func Client() (*stripe.Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client := &stripe.Client{
		BaseURL: server.URL,
	}
	return client, mux, func() {
		server.Close()
	}
}

func Customer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, customerJSON)
	}
}

const (
	customerJSON = `{
	"id": "cus_Do2qkvBR8OMXq3",
	"object": "customer",
	"account_balance": 0,
	"created": 1539827792,
	"currency": null,
	"default_source": "card_1DMQkW2eZvKYlo2CInhXngdF",
	"delinquent": false,
	"description": null,
	"discount": null,
	"email": null,
	"invoice_prefix": "E30F1CD",
	"livemode": false,
	"metadata": {},
	"shipping": null,
	"sources": {
		"object": "list",
		"data": [
			{
				"id": "card_1DMQkW2eZvKYlo2CInhXngdF",
				"object": "card",
				"address_city": null,
				"address_country": null,
				"address_line1": null,
				"address_line1_check": null,
				"address_line2": null,
				"address_state": null,
				"address_zip": null,
				"address_zip_check": null,
				"brand": "American Express",
				"country": "US",
				"customer": "cus_Do2qkvBR8OMXq3",
				"cvc_check": null,
				"dynamic_last4": null,
				"exp_month": 10,
				"exp_year": 2019,
				"fingerprint": "EdFCik9NII3EjtXE",
				"funding": "credit",
				"last4": "8431",
				"metadata": {},
				"name": null,
				"tokenization_method": null
			}
		],
		"has_more": false,
		"total_count": 1,
		"url": "/v1/customers/cus_Do2qkvBR8OMXq3/sources"
	},
	"subscriptions": {
		"object": "list",
		"data": [],
		"has_more": false,
		"total_count": 0,
		"url": "/v1/customers/cus_Do2qkvBR8OMXq3/subscriptions"
	},
	"tax_info": null,
	"tax_info_verification": null
}`
)
