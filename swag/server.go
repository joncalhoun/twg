package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/joncalhoun/twg/form"
	"github.com/joncalhoun/twg/stripe"
	"github.com/joncalhoun/twg/swag/db"
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

	db.CreateCampaign(time.Now(), time.Now().Add(time.Hour), 1200)

	mux := http.NewServeMux()
	resourceMux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets/"))
	mux.Handle("/img/", fs)
	mux.Handle("/css/", fs)
	mux.Handle("/favicon.ico", http.FileServer(http.Dir("./assets/img/")))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = cleanPath(r.URL.Path)
		resourceMux.ServeHTTP(w, r)
	})
	resourceMux.HandleFunc("/", showActiveCampaign)
	resourceMux.Handle("/campaigns/", http.StripPrefix("/campaigns", campaignsMux()))
	resourceMux.Handle("/orders/", http.StripPrefix("/orders", ordersMux()))

	port := os.Getenv("SWAG_PORT")
	if port == "" {
		port = "3000"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func ordersMux() http.Handler {
	// The order mux expects the order to be set in the context
	// and the ID to be trimmed from the path.
	ordMux := http.NewServeMux()
	ordMux.HandleFunc("/confirm/", confirmOrder)
	ordMux.HandleFunc("/", showOrder)
	// ordMux.HandleFunc("/confirm/", confirmOrder)

	// Trim the ID from the path, set the campaign in the ctx, and call
	// the cmpMux.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payCusID, path := splitPath(r.URL.Path)
		order, err := db.GetOrderViaPayCus(payCusID)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "order", order)
		r = r.WithContext(ctx)
		r.URL.Path = path
		ordMux.ServeHTTP(w, r)
	})
}

func campaignsMux() http.Handler {
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr, path := splitPath(r.URL.Path)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		campaign, err := db.GetCampaign(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "campaign", campaign)
		r = r.WithContext(ctx)
		r.URL.Path = path
		cmpMux.ServeHTTP(w, r)
	})
}

func cleanPath(pth string) string {
	pth = path.Clean("/" + pth)
	if pth[len(pth)-1] != '/' {
		pth = pth + "/"
	}
	return pth
}

func splitPath(pth string) (head, tail string) {
	pth = cleanPath(pth)
	parts := strings.SplitN(pth[1:], "/", 2)
	if len(parts) < 2 {
		parts = append(parts, "/")
	}
	return parts[0], cleanPath(parts[1])
}

func showActiveCampaign(w http.ResponseWriter, r *http.Request) {
	campaign, err := db.ActiveCampaign()
	switch {
	case err == sql.ErrNoRows:
		showCampaignEnded(w, r)
		return
	case err != nil:
		log.Printf("Error retrieving the active campaign. err = %v", err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	var leftValue int
	var leftUnit string
	left := time.Now().Sub(campaign.EndsAt)
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
	err = templates.Campaigns.Show.Execute(w, data)
	if err != nil {
		log.Printf("Error executing show_campaign template. err = %v", err)
	}
}

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

func newOrder(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	campaign := r.Context().Value("campaign").(*db.Campaign)

	data := struct {
		Campaign struct {
			ID    int
			Price int
		}
		OrderForm       orderForm
		StripePublicKey string
	}{}
	data.Campaign.ID = campaign.ID
	data.Campaign.Price = campaign.Price
	data.StripePublicKey = stripePublicKey
	err := templates.Orders.New.Execute(w, data)
	if err != nil {
		log.Printf("Error executing the new_order template. err = %v", err)
	}
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	campaign := r.Context().Value("campaign").(*db.Campaign)
	formData := struct {
		Name    string
		Email   string
		Street1 string
		Street2 string
		City    string
		State   string
		Zip     string
		Country string
	}{}
	r.ParseForm()
	schema.NewDecoder().Decode(&formData, r.PostForm)
	fmt.Println(formData)
	if formData.Email == "" {
		panic("email wasn't parsed!")
	}
	stripeClient := &stripe.Client{
		Key: stripeSecretKey,
	}
	cus, err := stripeClient.Customer(r.PostForm.Get("stripe-token"), formData.Email)
	if err != nil {
		log.Printf("Error creating stripe customer. email = %s, err = %v", formData.Email, err)
		http.Error(w, "Something went wrong processing your payment information. Try again, or contact me - jon@calhoun.io - if the problem persists.", http.StatusInternalServerError)
		return
	}
	var order db.Order
	order.CampaignID = campaign.ID
	// Customer
	order.Customer.Name = formData.Name
	order.Customer.Email = formData.Email
	// Address
	order.Address.Street1 = formData.Street1
	order.Address.Street2 = formData.Street2
	order.Address.City = formData.City
	order.Address.State = formData.State
	order.Address.Zip = formData.Zip
	order.Address.Country = formData.Country
	order.Address.Raw = fmt.Sprintf(`%s
%s
%s
%s %s  %s
%s`, order.Customer.Name,
		order.Address.Street1,
		order.Address.Street2,
		order.Address.City, order.Address.State, order.Address.Zip,
		order.Address.Country)
	order.Address.Raw = strings.Replace(order.Address.Raw, "\n\n", "\n", 1)
	order.Address.Raw = strings.ToUpper(order.Address.Raw)

	// Payment info
	order.Payment.Source = "stripe"
	order.Payment.CustomerID = cus.ID
	err = db.CreateOrder(&order)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%s", order.Payment.CustomerID), http.StatusFound)
}

func showOrder(w http.ResponseWriter, r *http.Request) {
	order := r.Context().Value("order").(*db.Order)
	campaign, err := db.GetCampaign(order.CampaignID)
	if err != nil {
		log.Println("error retrieving order campaign")
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	if order.Payment.ChargeID != "" {
		stripeClient := &stripe.Client{
			Key: stripeSecretKey,
		}
		chg, err := stripeClient.GetCharge(order.Payment.ChargeID)
		if err != nil {
			log.Printf("error looking up a customer's charge where chg.ID = %s; err = %v", order.Payment.ChargeID, err)
			fmt.Fprintln(w, "Failed to lookup the status of your order. Please try again, or contact me if this persists - jon@calhoun.io")
			return
		}
		switch chg.Status {
		case "succeeded":
			fmt.Fprintln(w, "Your order has been completed successfully! You will be contacted when it ships.")
		case "pending":
			fmt.Fprintln(w, "Your payment is still pending.")
		case "failed":
			fmt.Fprintln(w, "Your payment failed. :( Please create a new order with a new card if you want to try again.")
		}
		return
	}
	data := struct {
		Order struct {
			ID      string
			Address string
		}
		Campaign struct {
			Price int
		}
	}{}
	data.Order.ID = order.Payment.CustomerID
	data.Order.Address = order.Address.Raw
	data.Campaign.Price = campaign.Price / 100
	templates.Orders.Review.Execute(w, data)
}

func confirmOrder(w http.ResponseWriter, r *http.Request) {
	order := r.Context().Value("order").(*db.Order)
	campaign, err := db.GetCampaign(order.CampaignID)
	if err != nil {
		log.Println("error retrieving order campaign")
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	order.Address.Raw = r.PostFormValue("address-raw")
	stripeClient := &stripe.Client{
		Key: stripeSecretKey,
	}
	chg, err := stripeClient.Charge(order.Payment.CustomerID, campaign.Price)
	if err != nil {
		if se, ok := err.(stripe.Error); ok {
			fmt.Fprint(w, se.Message)
			return
		}
		http.Error(w, "Something went wrong processing your card. Please contact me for support - jon@calhoun.io", http.StatusInternalServerError)
		return
	}
	order.Payment.ChargeID = chg.ID
	statement := `UPDATE orders
	SET adr_raw = $2, pay_charge_id = $3
	WHERE id = $1`
	_, err = db.DB.Exec(statement, order.ID, order.Address.Raw, order.Payment.ChargeID)
	if err != nil {
		http.Error(w, "You were charged, but something went wrong saving your data. Please contact me for support - jon@calhoun.io", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%s", order.Payment.CustomerID), http.StatusFound)
}

func showCampaignEnded(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("./assets/html/campaign-ended.html")
	if err != nil {
		log.Printf("Error opening the campaign ended file. err = %v", err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	io.Copy(w, f)
}
