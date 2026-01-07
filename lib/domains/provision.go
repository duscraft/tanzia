package domains

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/duscraft/tanzia/lib/helpers"
)

type Provision struct {
	Label  string
	Amount float64
}

func AddProvisionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	canUserCreateProvision, err := helpers.CanUserCreateProvision(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !canUserCreateProvision {
		http.Redirect(w, r, "/dashboard#limit-provisions", http.StatusFound)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount value", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO provisions (label, amount, userId) VALUES ($1, $2, $3)", r.FormValue("label"), amount, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard#provision_added", http.StatusFound)
}

func ProvisionsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, err := template.ParseFiles("lib/templates/edit-provisions.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
