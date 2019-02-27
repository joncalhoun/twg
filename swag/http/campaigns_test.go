package http_test

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joncalhoun/twg/swag/db"
	. "github.com/joncalhoun/twg/swag/http"
)

func TestCampaignHandler_ShowActive(t *testing.T) {
	now := time.Now()
	timeNow := func() time.Time {
		return now
	}
	type checkFn func(*testing.T, string)
	checks := func(fns ...checkFn) []checkFn {
		return fns
	}
	hasBody := func(want string) func(*testing.T, string) {
		return func(t *testing.T, got string) {
			if got != want {
				t.Fatalf("body = %v; want %v", got, want)
			}
		}
	}

	// each test case returns a campaign handler along with the expected
	// body output as a string
	tests := map[string]func(*testing.T) (*CampaignHandler, []checkFn){
		"ID is 12": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(1 * time.Hour),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.ID}}"))
			return &ch, checks(hasBody("12"))
		},
		"Price is 10": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(1 * time.Hour),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.Price}}"))
			return &ch, checks(hasBody("10"))
		},
		"1 hour left": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(1 * time.Hour),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("1 hour(s)"))
		},
		"2 min left": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(2 * time.Minute),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("2 minute(s)"))
		},
		"3 days left": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(3 * 24 * time.Hour),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("3 day(s)"))
		},
		"25 sec left": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return &db.Campaign{
						ID:       12,
						StartsAt: now.Add(-1 * time.Hour),
						EndsAt:   now.Add(25 * time.Second),
						Price:    1000,
					}, nil
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("25 second(s)"))
		},
		"no active campaign": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return nil, sql.ErrNoRows
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Ended = template.Must(template.New("").Parse("No Active Campaign"))
			return &ch, checks(hasBody("No Active Campaign"))
		},
		"random sql error": func(t *testing.T) (*CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return nil, sql.ErrConnDone
				},
			}
			ch := CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			lr := &logRec{}
			ch.Logger = lr
			return &ch, checks(
				hasBody("Something went wrong..."),
				func(t *testing.T, _ string) {
					if len(lr.logs) != 1 {
						t.Fatalf("len(logs) = %d; want %d", len(lr.logs), 1)
					}
					want := fmt.Sprintf("Error retrieving the active campaign. err = %v", sql.ErrConnDone)
					if lr.logs[0] != want {
						t.Fatalf("log[0] = %v; want %v", lr.logs[0], want)
					}
				},
			)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ch, checks := tc(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ch.ShowActive(w, r)
			res := w.Result()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("ReadAll() err = %v; want nil", err)
			}
			defer res.Body.Close()

			got := strings.TrimSpace(string(resBody))
			for _, check := range checks {
				check(t, got)
			}
		})
	}
}

func TestCampaignHandler_CampaignMw(t *testing.T) {
	failHandler := func(t *testing.T) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("next handler shouldn't have been called by middleware")
		}
	}
	// Three paths:
	// 1. Invalid ID
	// 2. Can't find campaign or DB error
	// 3. Success
	t.Run("invalid id", func(t *testing.T) {
		ch := CampaignHandler{}
		handler := ch.CampaignMw(failHandler(t))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/invalid/id/here", nil)
		handler(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("missing campaign", func(t *testing.T) {
		ch := CampaignHandler{}
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				return nil, sql.ErrNoRows
			},
		}
		ch.DB = mdb
		handler := ch.CampaignMw(failHandler(t))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/123/id/here", nil)
		handler(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("StatusCode = %d; want %d", res.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("campaign found", func(t *testing.T) {
		campaign := &db.Campaign{
			ID:       123,
			StartsAt: time.Now(),
			EndsAt:   time.Now().Add(1 * time.Hour),
			Price:    1200,
		}
		ch := CampaignHandler{}
		mdb := &mockDB{
			GetCampaignFunc: func(id int) (*db.Campaign, error) {
				return campaign, nil
			},
		}
		ch.DB = mdb
		handlerCalled := false
		handler := ch.CampaignMw(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			if r.URL.Path != "/id/here/" {
				t.Fatalf("Path in next handler = %v; want %v", r.URL.Path, "/id/here/")
			}
			got := r.Context().Value("campaign").(*db.Campaign)
			if got != campaign {
				t.Fatalf("Campaign = %v; want %v", got, campaign)
			}
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/123/id/here", nil)
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
