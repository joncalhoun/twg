package main

import (
	"encoding/json"
	"fmt"

	stripe "github.com/joncalhoun/twg/stripe/v0"
)

func main() {
	// curl https://api.stripe.com/v1/charges \
	//    -u sk_test_4eC39HqLyjWDarjtT1zdp7dc: \
	//    -d amount=2000 \
	//    -d currency=usd \
	//    -d source=tok_mastercard \
	//    -d description="Charge for jenny.rosen@example.com"
	c := stripe.Client{
		Key: "sk_test_4eC39HqLyjWDarjtT1zdp7dc",
	}
	charge, err := c.Charge(2000, "tok_mastercard", "Charge for demo purposes.")
	if err != nil {
		panic(err)
	}
	fmt.Println(charge)
	jsonBytes, err := json.Marshal(charge)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}
