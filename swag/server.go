package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joncalhoun/twg/form"
	"github.com/joncalhoun/twg/stripe"
	"github.com/joncalhoun/twg/swag/db"
	swaghttp "github.com/joncalhoun/twg/swag/http"
)

var (
	templates struct {
		Orders struct {
			New    *template.Template
			Review *template.Template
		}
		Campaigns struct {
			Show *template.Template
		}
	}
)

const (
	formTemplateHTML = `
		<div class="w-full mb-6">
			<label class="block uppercase tracking-wide text-grey-darker text-xs font-bold mb-2" for="{{.Name}}">
				{{.Label}}
			</label>
			<input class="bg-grey-lighter appearance-none border-2 border-grey-lighter hover:border-orange rounded w-full py-2 px-4 text-grey-darker leading-tight" name="{{.Name}}" type="{{.Type}}" placeholder="{{.Placeholder}}">
		</div>`
)

var (
	stripeSecretKey = "sk_test_qrrEUOnYjJjybMTEsQnABuzE"
	stripePublicKey = "pk_test_pfEqL5GDjl8h4pXjv8CWpi80"
)

func init() {
	formTemplate := template.Must(template.New("").Parse(formTemplateHTML))

	templates.Orders.New = template.Must(template.New("new_order.gohtml").Funcs(template.FuncMap{
		"form_for": func(strct interface{}) (template.HTML, error) {
			return form.HTML(formTemplate, strct)
		},
	}).ParseFiles("./templates/new_order.gohtml"))

	templates.Orders.Review = template.Must(template.ParseFiles("./templates/review_order.gohtml"))

	templates.Campaigns.Show = template.Must(template.ParseFiles("./templates/show_campaign.gohtml"))
}

func main() {
	defer db.DB.Close()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	stripeClient := &stripe.Client{
		Key: stripeSecretKey,
	}
	campaignHandler := &swaghttp.CampaignHandler{}
	campaignHandler.DB = db.DefaultDatabase
	campaignHandler.Logger = logger
	campaignHandler.Templates.Show = templates.Campaigns.Show
	campaignHandler.Templates.Ended = template.Must(template.ParseFiles("./templates/ended_campaign.gohtml"))
	campaignHandler.TimeNow = time.Now

	orderHandler := &swaghttp.OrderHandler{}
	orderHandler.DB = db.DefaultDatabase
	orderHandler.Logger = logger
	orderHandler.Stripe.PublicKey = stripePublicKey
	orderHandler.Stripe.Client = stripeClient
	orderHandler.Templates.New = templates.Orders.New
	orderHandler.Templates.Review = templates.Orders.Review

	db.CreateCampaign(time.Now(), time.Now().Add(time.Hour), 1200)

	router := &swaghttp.Router{
		AssetDir:        "./assets/",
		FaviconDir:      "./assets/img/",
		OrderHandler:    orderHandler,
		CampaignHandler: campaignHandler,
	}

	port := os.Getenv("SWAG_PORT")
	if port == "" {
		port = "3000"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
