package http_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/joncalhoun/twg/swag/http"
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
}
