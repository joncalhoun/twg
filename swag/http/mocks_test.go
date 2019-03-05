package http_test

import (
	"fmt"
	"testing"

	"github.com/joncalhoun/twg/stripe"
	"github.com/joncalhoun/twg/swag/db"
)

type logRec struct {
	logs []string
}

func (lr *logRec) Printf(format string, v ...interface{}) {
	lr.logs = append(lr.logs, fmt.Sprintf(format, v...))
}

type logFail struct {
	t *testing.T
}

func (lf *logFail) Printf(format string, v ...interface{}) {
	lf.t.Fatalf("Printf() called unexpectedly: %s", fmt.Sprintf(format, v...))
}

type mockDB struct {
	ActiveCampaignFunc    func() (*db.Campaign, error)
	GetCampaignFunc       func(int) (*db.Campaign, error)
	CreateOrderFunc       func(*db.Order) error
	GetOrderViaPayCusFunc func(string) (*db.Order, error)
	ConfirmOrderFunc      func(orderID int, addressRaw, paymentChargeID string) error
}

func (mdb *mockDB) ActiveCampaign() (*db.Campaign, error) {
	return mdb.ActiveCampaignFunc()
}

func (mdb *mockDB) GetCampaign(id int) (*db.Campaign, error) {
	return mdb.GetCampaignFunc(id)
}

func (mdb *mockDB) CreateOrder(order *db.Order) error {
	return mdb.CreateOrderFunc(order)
}

func (mdb *mockDB) GetOrderViaPayCus(payCustomerID string) (*db.Order, error) {
	return mdb.GetOrderViaPayCusFunc(payCustomerID)
}

func (mdb *mockDB) ConfirmOrder(orderID int, addressRaw, paymentChargeID string) error {
	return mdb.ConfirmOrderFunc(orderID, addressRaw, paymentChargeID)
}

type mockStripe struct {
	CustomerFunc  func(token, email string) (*stripe.Customer, error)
	GetChargeFunc func(chargeID string) (*stripe.Charge, error)
	ChargeFunc    func(customerID string, amount int) (*stripe.Charge, error)
}

func (ms *mockStripe) Customer(token, email string) (*stripe.Customer, error) {
	return ms.CustomerFunc(token, email)
}

func (ms *mockStripe) GetCharge(chargeID string) (*stripe.Charge, error) {
	return ms.GetChargeFunc(chargeID)
}

func (ms *mockStripe) Charge(customerID string, amount int) (*stripe.Charge, error) {
	return ms.ChargeFunc(customerID, amount)
}
