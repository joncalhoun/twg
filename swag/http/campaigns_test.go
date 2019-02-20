package http_test

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joncalhoun/twg/swag/db"
	"github.com/joncalhoun/twg/swag/http"
)

type logRec struct {
	logs []string
}

func (lr *logRec) Printf(format string, v ...interface{}) {
	lr.logs = append(lr.logs, fmt.Sprintf(format, v...))
}

type mockDB struct {
	ActiveCampaignFunc func() (*db.Campaign, error)
}

func (mdb *mockDB) ActiveCampaign() (*db.Campaign, error) {
	return mdb.ActiveCampaignFunc()
}

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
	tests := map[string]func(*testing.T) (*http.CampaignHandler, []checkFn){
		"ID is 12": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.ID}}"))
			return &ch, checks(hasBody("12"))
		},
		"Price is 10": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.Price}}"))
			return &ch, checks(hasBody("10"))
		},
		"1 hour left": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("1 hour(s)"))
		},
		"2 min left": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("2 minute(s)"))
		},
		"3 days left": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("3 day(s)"))
		},
		"25 sec left": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
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
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Show = template.Must(template.New("").Parse("{{.TimeLeft.Value}} {{.TimeLeft.Unit}}"))
			return &ch, checks(hasBody("25 second(s)"))
		},
		"no active campaign": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return nil, sql.ErrNoRows
				},
			}
			ch := http.CampaignHandler{}
			ch.DB = mdb
			ch.TimeNow = timeNow
			ch.Templates.Ended = template.Must(template.New("").Parse("No Active Campaign"))
			return &ch, checks(hasBody("No Active Campaign"))
		},
		"random sql error": func(t *testing.T) (*http.CampaignHandler, []checkFn) {
			mdb := &mockDB{
				ActiveCampaignFunc: func() (*db.Campaign, error) {
					return nil, sql.ErrConnDone
				},
			}
			ch := http.CampaignHandler{}
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
			r := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
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
