package stripe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	Version = "2018-09-24"
)

type Customer struct {
	ID string `json:"id"`
}

type Client struct {
	Key string
}

func (c *Client) Customer(token string) (*Customer, error) {
	endpoint := "https://api.stripe.com/v1/customers"
	v := url.Values{}
	v.Set("source", token)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Stripe-Version", Version)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.Key, "")
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))
	var cus Customer
	err = json.Unmarshal(body, &cus)
	if err != nil {
		return nil, err
	}
	return &cus, nil
}
