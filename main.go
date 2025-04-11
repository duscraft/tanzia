package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"tantieme/domains"
	"tantieme/helpers"

	"github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/404.html")
	w.WriteHeader(http.StatusNotFound)
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	t, _ := template.ParseFiles("templates/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	connManager.AddConnection("sqlite3", "auth")
	defer connManager.CloseConnection("auth")

	http.HandleFunc("GET /persons", domains.PersonHandler)
	http.HandleFunc("POST /persons", domains.AddPersonHandler)
	http.HandleFunc("GET /bills", domains.BillsHandler)
	http.HandleFunc("POST /bills", domains.AddBillHandler)
	http.HandleFunc("GET /provisions", domains.ProvisionsHandler)
	http.HandleFunc("POST /provisions", domains.AddProvisionHandler)
	http.HandleFunc("GET /dashboard", domains.DashboardHandler)
	http.HandleFunc("POST /login", domains.LoginHandler)
	http.HandleFunc("GET /logout", domains.LogoutHandler)
	http.HandleFunc("GET /", indexHandler)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	_, _ = fmt.Fprintf(os.Stdout, "Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
