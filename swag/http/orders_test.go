package http_test

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			oh := OrderHandler{
				StripePublicKey: "sk_pub_123abc",
			}
			oh.Templates.New = template.Must(template.New("").Parse("{{.StripePublicKey}}"))
			return &oh, &db.Campaign{
				Price: 1200,
			}, checks(hasBody(oh.StripePublicKey))
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
