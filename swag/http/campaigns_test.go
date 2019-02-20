package http_test

import (
	"html/template"
	"io/ioutil"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joncalhoun/twg/swag/db"
	"github.com/joncalhoun/twg/swag/http"
)

type mockDB struct {
	ActiveCampaignFunc func() (*db.Campaign, error)
}

func (mdb *mockDB) ActiveCampaign() (*db.Campaign, error) {
	return mdb.ActiveCampaignFunc()
}

func TestCampaignHandler_ShowActive_okay(t *testing.T) {
	now := time.Now()
	activeCampaign := &db.Campaign{
		ID:       12,
		StartsAt: now.Add(-1 * time.Hour),
		EndsAt:   now.Add(1 * time.Hour),
		Price:    1000,
	}
	mdb := &mockDB{
		ActiveCampaignFunc: func() (*db.Campaign, error) {
			return activeCampaign, nil
		},
	}

	ch := http.CampaignHandler{}
	ch.DB = mdb
	ch.TimeNow = func() time.Time {
		return now
	}
	ch.Templates.Show = template.Must(template.New("").Parse(`ID: {{.ID}}
Price: {{.Price}}
TimeLeft.Value: {{.TimeLeft.Value}}
TimeLeft.Unit: {{.TimeLeft.Unit}}`))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	ch.ShowActive(w, r)
	res := w.Result()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ReadAll() err = %v; want nil", err)
	}
	defer res.Body.Close()

	got := string(resBody)
	want := `ID: 12
Price: 10
TimeLeft.Value: 1
TimeLeft.Unit: hour(s)`
	if got != want {
		t.Fatalf("CampaignHandler() body = %v; want %v", got, want)
	}
}
