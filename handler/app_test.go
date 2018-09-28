package app_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/joncalhoun/twg/app"
	"golang.org/x/net/publicsuffix"
)

func signedInClient(t *testing.T, baseURL string) *http.Client {
	// Our cookiejar will keep and set cookies for us between requests.
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		t.Fatalf("cookejar.New() err = %s; want nil", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	// Our client has a cookie jar, but it has no session cookie. By logging
	// in we can ensure that it gets set.
	loginURL := baseURL + "/login"
	req, err := http.NewRequest(http.MethodPost, loginURL, nil)
	if err != nil {
		t.Fatalf("NewRequest() err = %s; want nil", err)
	}
	_, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST /login err = %s; want nil", err)
	}
	return client
}

type headerClient struct {
	headers map[string]string
}

func (hc headerClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for hk, hv := range hc.headers {
		req.Header.Set(hk, hv)
	}
	client := http.Client{}
	return client.Do(req)
}

func TestApp(t *testing.T) {
	server := httptest.NewServer(&app.Server{})
	defer server.Close()
	t.Run("cookie based auth", func(t *testing.T) {
		client := signedInClient(t, server.URL)
		res, err := client.Get(server.URL + "/admin")
		if err != nil {
			t.Errorf("GET /admin err = %s; want nil", err)
		}
		if res.StatusCode != 200 {
			t.Errorf("GET /admin code = %d; want %d", res.StatusCode, 200)
		}
		res, err = client.Get(server.URL + "/header-admin")
		if err != nil {
			t.Errorf("GET /header-admin err = %s; want nil", err)
		}
		if res.StatusCode != 403 {
			t.Errorf("GET /header-admin code = %d; want %d", res.StatusCode, 403)
		}
	})

	t.Run("header based auth", func(t *testing.T) {
		client := headerClient{
			headers: map[string]string{"api-key": "fake_api_key"},
		}
		res, err := client.Get(server.URL + "/admin")
		if err != nil {
			t.Errorf("GET /admin err = %s; want nil", err)
		}
		if res.StatusCode != 403 {
			t.Errorf("GET /admin code = %d; want %d", res.StatusCode, 403)
		}
		res, err = client.Get(server.URL + "/header-admin")
		if err != nil {
			t.Errorf("GET /header-admin err = %s; want nil", err)
		}
		if res.StatusCode != 200 {
			t.Errorf("GET /header-admin code = %d; want %d", res.StatusCode, 200)
		}
	})
}
