package app

import (
	"fmt"
	"net/http"
	"sync"
)

const (
	fakeToken  = "fake_session_token"
	fakeAPIKey = "fake_api_key"
)

type Server struct {
	mux  *http.ServeMux
	once sync.Once
}

func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
		a.mux = http.NewServeMux()
		a.mux.HandleFunc("/", a.home)
		a.mux.HandleFunc("/login", a.login)
		a.mux.HandleFunc("/admin", cookieAuthMw(a.admin))
		a.mux.HandleFunc("/header-admin", headerAuthMw(a.admin))
	})
	a.mux.ServeHTTP(w, r)
}

func (a *Server) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Welcome!</h1>")
}

func (a *Server) login(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:  "session",
		Value: fakeToken,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func cookieAuthMw(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}
		// The way we use tokens here is insecure so don't copy this exactly
		// Get in touch if you have more questions about actually using
		// secure tokens - jon@calhoun.io
		if c.Value != fakeToken {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func headerAuthMw(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("api-key")
		if key != fakeAPIKey {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}
		next(w, r)
		next(w, r)
	}
}

func (a *Server) admin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Welcome to the admin page!</h1>")
}
