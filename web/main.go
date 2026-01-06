package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/duscraft/tanzia/lib/domains"
	"github.com/duscraft/tanzia/lib/helpers"

	"github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/404.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "website")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	t, err := template.ParseFiles("web/templates/index.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "website")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/login.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "app")
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/signup.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "app")
}

func cgvHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/cgv.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "website")
}

func legalsHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/legals.html", "web/templates/base-layout.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains.LogUserConnection(w, r, "website")
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if len(redisURL) == 0 {
		redisURL = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if len(redisPort) == 0 {
		redisPort = "6379"
	}
	session.InitManager(
		session.SetStore(redis.NewRedisStore(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", redisURL, redisPort),
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
	http.HandleFunc("GET /login", loginHandler)
	http.HandleFunc("GET /signup", signupHandler)
	http.HandleFunc("POST /login", domains.LoginHandler)
	http.HandleFunc("GET /logout", domains.LogoutHandler)
	http.HandleFunc("POST /signup", domains.SignupHandler)
	http.HandleFunc("GET /cgv", cgvHandler)
	http.HandleFunc("GET /legals", legalsHandler)
	http.HandleFunc("GET /export/pdf", domains.ExportPDFHandler)
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	http.HandleFunc("GET /", indexHandler)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
