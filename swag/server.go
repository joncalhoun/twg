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
	"github.com/joncalhoun/twg/swag/urlpath"
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
	campaignHandler := swaghttp.CampaignHandler{}
	campaignHandler.DB = db.DefaultDatabase
	campaignHandler.Logger = logger
	campaignHandler.Templates.Show = templates.Campaigns.Show
	campaignHandler.Templates.Ended = template.Must(template.ParseFiles("./templates/ended_campaign.gohtml"))
	campaignHandler.TimeNow = time.Now

	orderHandler := swaghttp.OrderHandler{}
	orderHandler.DB = db.DefaultDatabase
	orderHandler.Logger = logger
	orderHandler.Stripe.PublicKey = stripePublicKey
	orderHandler.Stripe.Client = stripeClient
	orderHandler.Templates.New = templates.Orders.New
	orderHandler.Templates.Review = templates.Orders.Review

	db.CreateCampaign(time.Now(), time.Now().Add(time.Hour), 1200)

	mux := http.NewServeMux()
	resourceMux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets/"))
	mux.Handle("/img/", fs)
	mux.Handle("/css/", fs)
	mux.Handle("/favicon.ico", http.FileServer(http.Dir("./assets/img/")))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = urlpath.Clean(r.URL.Path)
		resourceMux.ServeHTTP(w, r)
	})
	resourceMux.HandleFunc("/", campaignHandler.ShowActive)
	resourceMux.Handle("/campaigns/", http.StripPrefix("/campaigns", campaignsMux(campaignHandler.CampaignMw, orderHandler.New, orderHandler.Create)))
	resourceMux.Handle("/orders/", http.StripPrefix("/orders", ordersMux(orderHandler.OrderMw, orderHandler.Show, orderHandler.Confirm)))

	port := os.Getenv("SWAG_PORT")
	if port == "" {
		port = "3000"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func ordersMux(orderMw func(http.HandlerFunc) http.HandlerFunc, showOrder, confirmOrder http.HandlerFunc) http.Handler {
	// The order mux expects the order to be set in the context
	// and the ID to be trimmed from the path.
	ordMux := http.NewServeMux()
	ordMux.HandleFunc("/confirm/", confirmOrder)
	ordMux.HandleFunc("/", showOrder)
	// ordMux.HandleFunc("/confirm/", confirmOrder)

	// Trim the ID from the path, set the campaign in the ctx, and call
	// the cmpMux.
	return orderMw(ordMux.ServeHTTP)
}

func campaignsMux(campaignMw func(http.HandlerFunc) http.HandlerFunc, newOrder, createOrder http.HandlerFunc) http.Handler {
	// Paths like /campaigns/:id/orders/new are handled here, but most of
	// that path - the /campaigns/:id/orders part - is stripped and
	// processed beforehand.
	cmpOrdMux := http.NewServeMux()
	cmpOrdMux.HandleFunc("/new/", newOrder)
	cmpOrdMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createOrder(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// The campaign mux expects the campaign to be set in the context
	// and the ID to be trimmed from the path.
	cmpMux := http.NewServeMux()
	cmpMux.Handle("/orders/", http.StripPrefix("/orders", cmpOrdMux))

	// Trim the ID from the path, set the campaign in the ctx, and call
	// the cmpMux.
	return campaignMw(cmpMux.ServeHTTP)
}
