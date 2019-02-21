package http

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/joncalhoun/twg/swag/db"
	"github.com/joncalhoun/twg/swag/urlpath"
)

// CampaignHandler is a container for all deps for the web app and has methods
// that can be used as various HTTP handlers.
type CampaignHandler struct {
	DB interface {
		ActiveCampaign() (*db.Campaign, error)
		GetCampaign(int) (*db.Campaign, error)
	}
	Logger    Logger
	Templates struct {
		Show  *template.Template
		Ended *template.Template
	}
	TimeNow func() time.Time
}

// ShowActive is the HTTP handler for showing the current active campaign
func (ch *CampaignHandler) ShowActive(w http.ResponseWriter, r *http.Request) {
	campaign, err := ch.DB.ActiveCampaign()
	switch {
	case err == sql.ErrNoRows:
		err = ch.Templates.Ended.Execute(w, nil)
		if err != nil {
			ch.Logger.Printf("Error executing campaign ended template. err = %v", err)
		}
		return
	case err != nil:
		ch.Logger.Printf("Error retrieving the active campaign. err = %v", err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	var leftValue int
	var leftUnit string
	left := campaign.EndsAt.Sub(ch.TimeNow())
	switch {
	case left >= 24*time.Hour:
		leftValue = int(left / (24 * time.Hour))
		leftUnit = "day(s)"
	case left >= time.Hour:
		leftValue = int(left / time.Hour)
		leftUnit = "hour(s)"
	case left >= time.Minute:
		leftValue = int(left / time.Minute)
		leftUnit = "minute(s)"
	default:
		leftValue = int(left / time.Second)
		leftUnit = "second(s)"
	}
	data := struct {
		ID       int
		Price    int
		TimeLeft struct {
			Value int
			Unit  string
		}
	}{}
	data.ID = campaign.ID
	data.Price = campaign.Price / 100
	data.TimeLeft.Value = leftValue
	data.TimeLeft.Unit = leftUnit
	err = ch.Templates.Show.Execute(w, data)
	if err != nil {
		ch.Logger.Printf("Error executing campaign show template. err = %v", err)
	}
}

// CampaignMw will parse the campaign ID from the URL, lookup that
// campaign in the DB, then finally set that campaign to the request
// context before finally calling the next http.HanderFunc. It will also
// remove the ID of the campaign from the path and call the next handler
// with the updated path set to the URL so that the next handler can
// proceed as if it was never part of the URL path.
//
// If there is an error - such as an invalid ID or that campaign doesn't
// exist - then the http.NotFound handler will be called.
func (ch *CampaignHandler) CampaignMw(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr, path := urlpath.Split(r.URL.Path)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		campaign, err := ch.DB.GetCampaign(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "campaign", campaign)
		r = r.WithContext(ctx)
		r.URL.Path = path
		next(w, r)
	}
}
