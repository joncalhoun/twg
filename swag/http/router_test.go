package http_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/joncalhoun/twg/swag/http"
	"github.com/joncalhoun/twg/swag/urlpath"
)

func TestRouter(t *testing.T) {
	t.Run("assets", func(t *testing.T) {
		tests := []string{
			"img/test.txt",
			"css/test.css",
		}
		for _, tc := range tests {
			t.Run(tc, func(t *testing.T) {
				file, err := os.Open(fmt.Sprintf("testdata/%s", tc))
				if err != nil {
					t.Fatalf("Open() err = %v; want %v", err, nil)
				}
				fileBytes, err := ioutil.ReadAll(file)
				if err != nil {
					t.Fatalf("ReadAll() err = %v; want %v", err, nil)
				}
				want := string(fileBytes)

				router := &Router{
					AssetDir:        "testdata/",
					CampaignHandler: &CampaignHandler{},
					OrderHandler:    &OrderHandler{},
				}
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc), nil)
				router.ServeHTTP(w, r)
				res := w.Result()
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("ReadAll() err= %v; want %v", err, nil)
				}
				got := string(body)

				if got != want {
					t.Fatalf("body contents = %v; want %v", got, want)
				}
			})
		}
	})

	idMw := func(t *testing.T, idWant string) func(http.HandlerFunc) http.HandlerFunc {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				id, path := urlpath.Split(r.URL.Path)
				if idWant != "" && id != idWant {
					t.Fatalf("ID = %v; want %v", id, idWant)
				}
				r.URL.Path = path
				next(w, r)
			}
		}
	}
	failMw := func(t *testing.T) func(http.HandlerFunc) http.HandlerFunc {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("Mw called unexpectedly")
			}
		}
	}

	t.Run("successes", func(t *testing.T) {
		want := "SUCCESS"
		tests := map[string]func(*testing.T) (r *Router, method, path string){
			"show order": func(t *testing.T) (*Router, string, string) {
				orderID := "ord_abc123"
				return &Router{
					AssetDir: "testdata/",
					CampaignHandler: &RouterCampaignHandlerMock{
						CampaignMwFunc: failMw(t),
					},
					OrderHandler: &RouterOrderHandlerMock{
						ShowFunc: func(w http.ResponseWriter, r *http.Request) {
							fmt.Fprint(w, want)
						},
						OrderMwFunc: idMw(t, orderID),
					},
				}, http.MethodGet, fmt.Sprintf("/orders/%v", orderID)
			},
			"confirm order": func(t *testing.T) (*Router, string, string) {
				orderID := "ord_abc123"
				return &Router{
					AssetDir: "testdata/",
					CampaignHandler: &RouterCampaignHandlerMock{
						CampaignMwFunc: failMw(t),
					},
					OrderHandler: &RouterOrderHandlerMock{
						ConfirmFunc: func(w http.ResponseWriter, r *http.Request) {
							fmt.Fprint(w, want)
						},
						OrderMwFunc: idMw(t, orderID),
					},
				}, http.MethodPost, fmt.Sprintf("/orders/%v/confirm", orderID)
			},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				router, method, path := tc(t)

				w := httptest.NewRecorder()
				r := httptest.NewRequest(method, path, nil)
				router.ServeHTTP(w, r)
				res := w.Result()
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("ReadAll() err= %v; want %v", err, nil)
				}
				got := string(body)

				if got != want {
					t.Fatalf("body contents = %v; want %v", got, want)
				}
			})
		}
	})
}
