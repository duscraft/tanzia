package main

import (
	"fmt"
	"github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
	"html/template"
	"log"
	"net/http"
	"os"
	"tantieme/domains"
	"tantieme/helpers"
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
	session.InitManager(
		session.SetStore(redis.NewRedisStore(&redis.Options{
			Addr: "127.0.0.1:6379",
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
	http.HandleFunc("GET /dashboard", domains.DashboardHandler)
	http.HandleFunc("POST /login", domains.LoginHandler)
	http.HandleFunc("GET /logout", domains.LogoutHandler)
	http.HandleFunc("GET /", indexHandler)

	_, _ = fmt.Fprint(os.Stdout, "Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
