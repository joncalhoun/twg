package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var (
	DB *tempDB

	DefaultDatabase = &Database{}
)

type tempDB struct{}

func (tdb *tempDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return DefaultDatabase.sqlDB.Exec(query, args...)
}

func (tdb *tempDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return DefaultDatabase.sqlDB.QueryRow(query, args...)
}

func (tdb *tempDB) Close() error {
	return DefaultDatabase.sqlDB.Close()
}

const (
	defaultURL = "postgres://postgres@127.0.0.1:5432/swag_dev?sslmode=disable"
)

func init() {
	Open(defaultURL)
}

// Open will open a database connection using the provided postgres URL.
// Be sure to close this using the db.DB.Close function.
func Open(psqlURL string) error {
	db, err := sql.Open("postgres", psqlURL)
	if err != nil {
		return err
	}
	DefaultDatabase.sqlDB = db
	return nil
}

type Campaign struct {
	ID       int
	StartsAt time.Time
	EndsAt   time.Time
	Price    int
}

type Database struct {
	sqlDB *sql.DB
}

func CreateCampaign(start, end time.Time, price int) (*Campaign, error) {
	return DefaultDatabase.CreateCampaign(start, end, price)
}

func (db *Database) CreateCampaign(start, end time.Time, price int) (*Campaign, error) {
	statement := `
	INSERT INTO campaigns(starts_at, ends_at, price)
	VALUES($1, $2, $3)
	RETURNING id`
	var id int
	err := db.sqlDB.QueryRow(statement, start, end, price).Scan(&id)
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

var timeNow = time.Now

func ActiveCampaign() (*Campaign, error) {
	return DefaultDatabase.ActiveCampaign()
}

func (db *Database) ActiveCampaign() (*Campaign, error) {
	statement := `
	SELECT * FROM campaigns
	WHERE starts_at <= $1
	AND ends_at >= $1`
	row := db.sqlDB.QueryRow(statement, timeNow())
	var camp Campaign
	err := row.Scan(&camp.ID, &camp.StartsAt, &camp.EndsAt, &camp.Price)
	if err != nil {
		return nil, err
	}
	return &camp, nil
}

func GetCampaign(id int) (*Campaign, error) {
	return DefaultDatabase.GetCampaign(id)
}

func (db *Database) GetCampaign(id int) (*Campaign, error) {
	statement := `
	SELECT * FROM campaigns
	WHERE id = $1`
	row := db.sqlDB.QueryRow(statement, id)
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
	return DefaultDatabase.CreateOrder(order)
}

func (db *Database) CreateOrder(order *Order) error {
	statement := `
	INSERT INTO orders (
		campaign_id,
		cus_name, cus_email,
		adr_street1, adr_street2, adr_city, adr_state, adr_zip, adr_country, adr_raw,
		pay_source, pay_customer_id, pay_charge_id
	)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING id`
	err := db.sqlDB.QueryRow(statement,
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
	return DefaultDatabase.GetOrderViaPayCus(payCustomerID)
}

func (db *Database) GetOrderViaPayCus(payCustomerID string) (*Order, error) {
	statement := `
	SELECT * FROM orders
	WHERE pay_customer_id = $1`
	row := db.sqlDB.QueryRow(statement, payCustomerID)
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
