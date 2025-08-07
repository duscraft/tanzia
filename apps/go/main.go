package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"tanzia/apps/go/domains"
	"tanzia/apps/go/helpers"

	"github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("apps/go/templates/404.html", "apps/go/templates/base-layout.html")
	w.WriteHeader(http.StatusNotFound)
	err := t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	domains.LogUserConnection(w, r, "website")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	t, _ := template.ParseFiles("apps/go/templates/index.html", "apps/go/templates/base-layout.html")
	err := t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	domains.LogUserConnection(w, r, "website")
}

func cgvHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("apps/go/templates/cgv.html", "apps/go/templates/base-layout.html")
	err := t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	domains.LogUserConnection(w, r, "website")
}

func legalsHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("apps/go/templates/legals.html", "apps/go/templates/base-layout.html")
	err := t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	domains.LogUserConnection(w, r, "website")
}

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	if len(redisUrl) == 0 {
		redisUrl = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if len(redisPort) == 0 {
		redisPort = "6379"
	}
	session.InitManager(
		session.SetStore(redis.NewRedisStore(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", redisUrl, redisPort),
			DB:   0,
		})),
	)

	connManager := helpers.GetConnectionManager()

	_, _ = connManager.AddConnection("postgres")
	defer func() {
		if err := connManager.CloseConnection(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	http.HandleFunc("GET /persons", domains.PersonHandler)
	http.HandleFunc("POST /persons", domains.AddPersonHandler)
	http.HandleFunc("GET /bills", domains.BillsHandler)
	http.HandleFunc("POST /bills", domains.AddBillHandler)
	http.HandleFunc("GET /provisions", domains.ProvisionsHandler)
	http.HandleFunc("POST /provisions", domains.AddProvisionHandler)
	http.HandleFunc("GET /dashboard", domains.DashboardHandler)
	http.HandleFunc("POST /login", domains.LoginHandler)
	http.HandleFunc("GET /logout", domains.LogoutHandler)
	http.HandleFunc("POST /signup", domains.SignupHandler)
	http.HandleFunc("GET /cgv", cgvHandler)
	http.HandleFunc("GET /legals", legalsHandler)
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("GET /", indexHandler)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	_, _ = fmt.Fprintf(os.Stdout, "Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
