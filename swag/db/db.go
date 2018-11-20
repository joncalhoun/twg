package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB
)

const (
	host    = "localhost"
	port    = "5432"
	user    = "jon"
	dbName  = "swag_dev"
	sslMode = "disable"
)

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", host, port, user, dbName, sslMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// Be sure to close the DB!
	DB = db
}

type Campaign struct {
	ID       int
	StartsAt time.Time
	EndsAt   time.Time
	Price    int
}

func CreateCampaign(start, end time.Time, price int) (*Campaign, error) {
	statement := `
	INSERT INTO campaigns(starts_at, ends_at, price)
	VALUES($1, $2, $3)
	RETURNING id`
	var id int
	err := DB.QueryRow(statement, start, end, price).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Campaign{
		ID:       id,
		StartsAt: start,
		EndsAt:   end,
		Price:    price,
	}, nil
}

func ActiveCampaign() (*Campaign, error) {
	statement := `
	SELECT * FROM campaigns
	WHERE starts_at <= $1
	AND ends_at >= $1`
	row := DB.QueryRow(statement, time.Now())
	var camp Campaign
	err := row.Scan(&camp.ID, &camp.StartsAt, &camp.EndsAt, &camp.Price)
	if err != nil {
		return nil, err
	}
	return &camp, nil
}

func GetCampaign(id int) (*Campaign, error) {
	statement := `
	SELECT * FROM campaigns
	WHERE id = $1`
	row := DB.QueryRow(statement, id)
	var camp Campaign
	err := row.Scan(&camp.ID, &camp.StartsAt, &camp.EndsAt, &camp.Price)
	if err != nil {
		return nil, err
	}
	return &camp, nil
}

type Customer struct {
	Name  string
	Email string
}

type Address struct {
	Street1 string
	Street2 string
	City    string
	State   string
	Zip     string
	Country string
	// In case the format above fails
	Raw string
}

type Payment struct {
	Source     string
	CustomerID string
	ChargeID   string
}

type Order struct {
	ID         int
	CampaignID int
	Customer   Customer
	Address    Address
	Payment    Payment
}

func CreateOrder(order *Order) error {
	statement := `
	INSERT INTO orders (
		campaign_id,
		cus_name, cus_email,
		adr_street1, adr_street2, adr_city, adr_state, adr_zip, adr_country, adr_raw,
		pay_source, pay_customer_id, pay_charge_id
	)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING id`
	err := DB.QueryRow(statement,
		order.CampaignID,
		order.Customer.Name,
		order.Customer.Email,
		order.Address.Street1,
		order.Address.Street2,
		order.Address.City,
		order.Address.State,
		order.Address.Zip,
		order.Address.Country,
		order.Address.Raw,
		order.Payment.Source,
		order.Payment.CustomerID,
		order.Payment.ChargeID,
	).Scan(&order.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetOrderViaPayCus(payCustomerID string) (*Order, error) {
	statement := `
	SELECT * FROM orders
	WHERE pay_customer_id = $1`
	row := DB.QueryRow(statement, payCustomerID)
	var ord Order
	err := row.Scan(
		&ord.ID,
		&ord.CampaignID,
		&ord.Customer.Name,
		&ord.Customer.Email,
		&ord.Address.Street1,
		&ord.Address.Street2,
		&ord.Address.City,
		&ord.Address.State,
		&ord.Address.Zip,
		&ord.Address.Country,
		&ord.Address.Raw,
		&ord.Payment.Source,
		&ord.Payment.CustomerID,
		&ord.Payment.ChargeID,
	)
	if err != nil {
		return nil, err
	}
	return &ord, nil
}
