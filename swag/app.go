package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/joncalhoun/twg/swag/db"
)

type App struct {
	DB        *db.Database
	Logger    *log.Logger
	Templates struct {
		Campaigns struct {
			Show  *template.Template
			Ended *template.Template
		}
	}
	TimeNow func() time.Time
}

func (app *App) ShowActiveCampaign(w http.ResponseWriter, r *http.Request) {
	campaign, err := app.DB.ActiveCampaign()
	switch {
	case err == sql.ErrNoRows:
		err = app.Templates.Campaigns.Ended.Execute(w, nil)
		if err != nil {
			app.Logger.Printf("Error executing campaign ended template. err = %v", err)
		}
		return
	case err != nil:
		app.Logger.Printf("Error retrieving the active campaign. err = %v", err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	var leftValue int
	var leftUnit string
	left := app.TimeNow().Sub(campaign.EndsAt)
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
	err = app.Templates.Campaigns.Show.Execute(w, data)
	if err != nil {
		app.Logger.Printf("Error executing campaign show template. err = %v", err)
	}
}
