package http

import (
	"net/http"
	"sync"

	"github.com/joncalhoun/twg/swag/urlpath"
)

type Router struct {
	AssetDir     string
	FaviconDir   string
	OrderHandler interface {
		New(w http.ResponseWriter, r *http.Request)
		Create(w http.ResponseWriter, r *http.Request)
		Show(w http.ResponseWriter, r *http.Request)
		Confirm(w http.ResponseWriter, r *http.Request)
		OrderMw(next http.HandlerFunc) http.HandlerFunc
	}
	CampaignHandler interface {
		ShowActive(w http.ResponseWriter, r *http.Request)
		CampaignMw(next http.HandlerFunc) http.HandlerFunc
	}

	once    sync.Once
	handler http.Handler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.once.Do(func() {
		mux := http.NewServeMux()
		resourceMux := http.NewServeMux()
		fs := http.FileServer(http.Dir(router.AssetDir))
		mux.Handle("/img/", fs)
		mux.Handle("/css/", fs)
		mux.Handle("/favicon.ico", http.FileServer(http.Dir(router.FaviconDir)))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = urlpath.Clean(r.URL.Path)
			resourceMux.ServeHTTP(w, r)
		})
		resourceMux.HandleFunc("/", router.CampaignHandler.ShowActive)
		resourceMux.Handle("/campaigns/", http.StripPrefix("/campaigns", campaignsMux(router.CampaignHandler.CampaignMw, router.OrderHandler.New, router.OrderHandler.Create)))
		resourceMux.Handle("/orders/", http.StripPrefix("/orders", ordersMux(router.OrderHandler.OrderMw, router.OrderHandler.Show, router.OrderHandler.Confirm)))
		router.handler = mux
	})

	router.handler.ServeHTTP(w, r)
}

func ordersMux(orderMw func(http.HandlerFunc) http.HandlerFunc, showOrder, confirmOrder http.HandlerFunc) http.Handler {
	// The order mux expects the order to be set in the context
	// and the ID to be trimmed from the path.
	ordMux := http.NewServeMux()
	ordMux.HandleFunc("/confirm/", confirmOrder)
	ordMux.HandleFunc("/", showOrder)
	// ordMux.HandleFunc("/confirm/", confirmOrder)

	// Trim the ID from the path, set the campaign in the ctx, and call
	// the cmpMux.
	return orderMw(ordMux.ServeHTTP)
}

func campaignsMux(campaignMw func(http.HandlerFunc) http.HandlerFunc, newOrder, createOrder http.HandlerFunc) http.Handler {
	// Paths like /campaigns/:id/orders/new are handled here, but most of
	// that path - the /campaigns/:id/orders part - is stripped and
	// processed beforehand.
	cmpOrdMux := http.NewServeMux()
	cmpOrdMux.HandleFunc("/new/", newOrder)
	cmpOrdMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createOrder(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// The campaign mux expects the campaign to be set in the context
	// and the ID to be trimmed from the path.
	cmpMux := http.NewServeMux()
	cmpMux.Handle("/orders/", http.StripPrefix("/orders", cmpOrdMux))

	// Trim the ID from the path, set the campaign in the ctx, and call
	// the cmpMux.
	return campaignMw(cmpMux.ServeHTTP)
}
