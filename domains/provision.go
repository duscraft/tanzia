package domains

import (
	"html/template"
	"net/http"
	"strconv"
	"tantieme/helpers"
)

type Provision struct {
	Label  string
	Amount float64
}

func AddProvisionHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := GetAuthenticatedUsername(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	_, err := db.Exec("INSERT INTO provisions (label, amount) VALUES (?, ?)", r.FormValue("label"), amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#provision_added", http.StatusFound)
}

func ProvisionsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUsername(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("templates/edit-provisions.html")
	err := t.Execute(w, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
