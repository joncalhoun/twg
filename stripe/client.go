package stripe

type Customer struct {
	ID string
}

type Client struct{}

func (c *Client) Customer(token string) (*Customer, error) {
	return nil, nil
}
