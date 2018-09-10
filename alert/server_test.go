package alert_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/alert"
	"golang.org/x/net/html"
)

func getText(n *html.Node) string {
	var ret []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			ret = append(ret, strings.TrimSpace(c.Data))
		case html.ElementNode:
			getText(c)
		}
	}
	return strings.Join(ret, " ")
}

func findNodes(body, tag, class string) ([]string, error) {
	// Source adapted from the html package's examples
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	var ret []string
	var find func(n *html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			for _, a := range n.Attr {
				if a.Key == "class" {
					classes := strings.Fields(a.Val)
					for _, c := range classes {
						if c == class {
							ret = append(ret, getText(n))
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	return ret, nil
}

// See https://golang.org/src/net/http/httptest/recorder_test.go for more
// examples
func TestApp(t *testing.T) {
	app := alert.App{}

	type checkFn func(r *http.Response, body string) error
	hasNoAlerts := func() checkFn {
		return func(r *http.Response, body string) error {
			nodes, err := findNodes(body, "div", "alert")
			if err != nil {
				return fmt.Errorf("findNodes() err = %s", err)
			}
			if len(nodes) != 0 {
				return fmt.Errorf("len(alerts)=%d; want len(alerts)=0", len(nodes))
			}
			return nil
		}
	}
	hasAlert := func(msg string) checkFn {
		return func(r *http.Response, body string) error {
			nodes, err := findNodes(body, "div", "alert")
			if err != nil {
				return fmt.Errorf("findNodes() err = %s", err)
			}
			for _, node := range nodes {
				if node == msg {
					return nil
				}
			}
			return fmt.Errorf("missing alert: %q", msg)
		}
	}

	tests := []struct {
		method string
		path   string
		body   io.Reader
		checks []checkFn
	}{
		{http.MethodGet, "/", nil, []checkFn{hasNoAlerts()}},
		{http.MethodGet, "/alert", nil, []checkFn{hasAlert("Stuff went wrong!")}},
		{http.MethodGet, "/many", nil, []checkFn{hasAlert("Alert Number 1"), hasAlert("Alert Number 2")}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tc.method, tc.path, tc.body)
			if err != nil {
				t.Fatalf("http.NewRequest() err = %s", err)
			}
			app.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()
			var sb strings.Builder
			io.Copy(&sb, res.Body)

			for _, check := range tc.checks {
				if err := check(res, sb.String()); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
