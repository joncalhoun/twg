package http

import (
	"html/template"
	"net/http"

	"github.com/joncalhoun/twg/swag/db"
)

type orderForm struct {
	Customer struct {
		Name  string `form:"placeholder=Jane Doe"`
		Email string `form:"type=email;placeholder=jane@doe.com;label=Email Address"`
	}
	Address struct {
		Street1 string `form:"placeholder=123 Sticker St;label=Street 1"`
		Street2 string `form:"placeholder=Apt 45;label=Street 2"`
		City    string `form:"placeholder=San Francisco"`
		State   string `form:"label=State (or Province);placeholder=CA"`
		Zip     string `form:"label=Postal Code;placeholder=94139"`
		Country string `form:"placeholder=United States"`
	}
}

type OrderHandler struct {
	StripePublicKey string
	Templates       struct {
		New *template.Template
	}
	Logger Logger
}

func (oh *OrderHandler) New(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// campaign := r.Context().Value("campaign").(*db.Campaign)
	campaign, ok := r.Context().Value("campaign").(*db.Campaign)
	if !ok {
		http.Error(w, "Campaign not provided.", http.StatusInternalServerError)
		return
	}

	data := struct {
		Campaign struct {
			ID    int
			Price int
		}
		OrderForm       orderForm
		StripePublicKey string
	}{}
	data.Campaign.ID = campaign.ID
	data.Campaign.Price = campaign.Price / 100
	data.StripePublicKey = oh.StripePublicKey
	err := oh.Templates.New.Execute(w, data)
	if err != nil {
		oh.Logger.Printf("Error executing the new_order template. err = %v", err)
	}
}
