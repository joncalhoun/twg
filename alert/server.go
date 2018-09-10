package alert

import (
	"fmt"
	"net/http"
	"sync"
)

type App struct {
	sync.Once
	mux http.ServeMux
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Once.Do(func() {
		a.mux = http.ServeMux{}
		a.mux.HandleFunc("/", a.Home)
		a.mux.HandleFunc("/alert", a.WithAlert)
		a.mux.HandleFunc("/many", a.ManyAlerts)
	})
	a.mux.ServeHTTP(w, r)
}

func (a App) ManyAlerts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
	<html>
		<head><title>Alert Page</title></head>
		<body>
			<div class="alert alert-primary" role="alert">
				Alert Number 1
			</div>
			<div class="alert alert-primary" role="alert">
				Alert Number 2
			</div>
			<h1>Alert page</h1>
		</body>
	</html>`)
}

func (a App) WithAlert(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
	<html>
		<head><title>Alert Page</title></head>
		<body>
			<div class="alert alert-primary" role="alert">
				Stuff went wrong!
			</div>
			<h1>Alert page</h1>
		</body>
	</html>`)
}

func (a App) Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
	<html>
		<head><title>Home Page</title></head>
		<body>
			<h1>Welcome</h1>
			<p>This is the home page</p>
		</body>
	</html>`)
}
