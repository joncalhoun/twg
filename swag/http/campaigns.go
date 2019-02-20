package http

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/joncalhoun/twg/swag/db"
)

// CampaignHandler is a container for all deps for the web app and has methods
// that can be used as various HTTP handlers.
type CampaignHandler struct {
	DB interface {
		ActiveCampaign() (*db.Campaign, error)
	}
	Logger interface {
		Printf(format string, v ...interface{})
	}
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
