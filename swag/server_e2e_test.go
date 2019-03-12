//+build e2e

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joncalhoun/twg/swag/db"
	"github.com/joncalhoun/twg/swag/http"
	"github.com/sclevine/agouti"
)

var (
	sleepDuration = 150 * time.Millisecond
	shortSleep    = sleepDuration
	medSleep      = 4 * shortSleep
	longSleep     = 4 * medSleep
)

func init() {
	stripeSecretKey = "sk_test_qrrEUOnYjJjybMTEsQnABuzE"
	stripePublicKey = "pk_test_pfEqL5GDjl8h4pXjv8CWpi80"
}

func resetDB(t *testing.T, sqlDB *sql.DB) {
	_, err := sqlDB.Exec("DELETE FROM orders")
	if err != nil {
		t.Fatalf("DELETE FROM orders err = %v; want nil", err)
	}
	_, err = sqlDB.Exec("DELETE FROM campaigns")
	if err != nil {
		t.Fatalf("DELETE FROM campaigns err = %v; want nil", err)
	}
}

func count(t *testing.T, sqlDB *sql.DB, table string) int {
	var n int
	err := sqlDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
	if err != nil {
		t.Fatalf("Scan() err = %v; want nil", err)
	}
	return n
}

func TestOrder(t *testing.T) {
	sqlDB, err := sql.Open("postgres", psqlURL)
	if err != nil {
		t.Fatalf("sql.Open() err = %v; want %v", err, nil)
	}
	resetDB(t, sqlDB)
	database, err := db.Open(db.WithSqlDB(sqlDB))
	if err != nil {
		t.Fatalf("db.Open() err = %v; want %v", err, nil)
	}
	defer database.Close()
	database.CreateCampaign(time.Now(), time.Now().Add(time.Hour), 1200)

	handler := setupHandler(database)
	port := os.Getenv("SWAG_PORT")
	if port == "" {
		port = "3000"
	}
	addr := fmt.Sprintf(":%s", port)
	go func() {
		log.Fatal(http.ListenAndServe(addr, handler))
	}()
	time.Sleep(medSleep)

	driver := agouti.ChromeDriver()
	if err := driver.Start(); err != nil {
		t.Fatalf("Start() err = %v; want nil", err)
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		t.Fatalf("NewPage() err = %v; want nil", err)
	}
	defer page.CloseWindow()
	page.Size(1840, 1000)

	err = page.Navigate("http://localhost:3000")
	if err != nil {
		t.Fatalf("Navigate() err = %v; want nil", err)
	}
	err = page.FindByLink("BUY NOW").Click()
	if err != nil {
		t.Fatalf("Click() err = %v; want nil", err)
	}

	numOrders := count(t, sqlDB, "orders")

	button := page.FindByButton("Review your order →")
	n, err := button.Count()
	if n != 1 || err != nil {
		t.Fatalf("FindByButton() n = %d, err = %v; want 1, nil", n, err)
	}

	if count(t, sqlDB, "orders") != numOrders {
		t.Fatal("Order count increased on invalid order form")
	}

	err = button.Submit()
	if err != nil {
		t.Fatalf("Submit() err = %v; want nil", err)
	}
	time.Sleep(medSleep)

	wantStripeErr := "Your card number is incomplete."
	stripeErr, err := page.FindByID("card-errors").Text()
	if stripeErr != wantStripeErr || err != nil {
		t.Fatalf("Text() = (%q, %v); want (%q, %v)", stripeErr, err, wantStripeErr, nil)
	}

	data := map[string]string{
		"Name":    "Jon Calhoun",
		"Email":   "test@gopherswag.com",
		"Street1": "123 Fake St",
		"Street2": "Apt 456",
		"City":    "Beverly Hills",
		"State":   "CA",
		"Zip":     "90210",
		"Country": "USA",
	}
	for name, value := range data {
		err := page.FindByName(name).Fill(value)
		if err != nil {
			t.Fatalf("Fill(%q) err = %v; want nil", value, err)
		}
	}

	err = page.FindByName("__privateStripeFrame4").SwitchToFrame()
	if err != nil {
		t.Fatalf("SwitchToFrame() err = %v; want nil", err)
	}
	page.FindByName("cardnumber").Fill("424242")
	time.Sleep(shortSleep)
	page.FindByName("cardnumber").Fill("4242424242424242")
	time.Sleep(shortSleep)
	page.FindByName("exp-date").Fill("12/24")
	time.Sleep(shortSleep)
	page.FindByName("cvc").Fill("123")
	time.Sleep(shortSleep)
	page.FindByName("postal").Fill("15522")
	time.Sleep(shortSleep)

	err = page.SwitchToRootFrame()
	if err != nil {
		t.Fatalf("SwitchToFrame() err = %v; want nil", err)
	}

	err = page.FindByButton("Review your order →").Submit()
	if err != nil {
		t.Fatalf("Submit() err = %v; want nil", err)
	}
	time.Sleep(longSleep)

	diff := count(t, sqlDB, "orders") - numOrders
	if diff != 1 {
		t.Fatalf("Order count should have increased by 1. Instead it changed by %d", diff)
	}

	url, err := page.URL()
	if err != nil {
		t.Fatalf("URL() err = %v; want nil", err)
	}
	pieces := strings.Split(url, "/")
	stripeCusID := pieces[len(pieces)-1]
	t.Logf("URL: %s", url)
	t.Logf("Stripe customer ID: %s", stripeCusID)

	cus, err := database.GetOrderViaPayCus(stripeCusID)
	if err != nil {
		t.Fatalf("GetOrderViaPayCus() err = %v; want nil", err)
	}
	if cus.Payment.ChargeID != "" {
		t.Fatalf("ChargeID = %q; want %q", cus.Payment.ChargeID, "")
	}

	button = page.FindByButton("Complete my order")
	n, err = button.Count()
	if n != 1 || err != nil {
		t.Fatalf("FindByButton() n = %d, err = %v; want 1, nil", n, err)
	}
	err = button.Submit()
	if err != nil {
		t.Fatalf("Submit() err = %v; want nil", err)
	}
	time.Sleep(medSleep)

	wantText := "Your order has been completed successfully!"
	gotText, err := page.HTML()
	if err != nil {
		t.Fatalf("HTML() err = %v; want nil", err)
	}
	if !strings.Contains(gotText, wantText) {
		t.Fatalf("HTML() = %q; want substring %q", gotText, wantText)
	}

	cus, err = database.GetOrderViaPayCus(stripeCusID)
	if err != nil {
		t.Fatalf("GetOrderViaPayCus() err = %v; want nil", err)
	}
	if strings.HasPrefix(cus.Payment.ChargeID, "chg_") {
		t.Fatalf("ChargeID = %q; want prefix %q", cus.Payment.ChargeID, "chg_")
	}
}
