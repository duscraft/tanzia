package domains

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/duscraft/tanzia/lib/helpers"
)

type Person struct {
	Name     string
	Tantieme int
}

func PersonHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, err := template.ParseFiles("lib/templates/edit-persons.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AddPersonHandler(w http.ResponseWriter, r *http.Request) {
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

	canUserCreatePerson, err := helpers.CanUserCreatePerson(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !canUserCreatePerson {
		http.Redirect(w, r, "/dashboard#limit-persons", http.StatusFound)
		return
	}

	tantieme, err := strconv.Atoi(r.FormValue("tantieme"))
	if err != nil {
		http.Error(w, "Invalid tantieme value", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO persons (name, tantieme, userId) VALUES ($1, $2, $3)", r.FormValue("name"), tantieme, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard#person_added", http.StatusFound)
}

func (person *Person) CalculateDue(totalTantiemes int, bill Bill) float64 {
	return float64(person.Tantieme) / float64(totalTantiemes) * bill.Amount
}

func (person *Person) CalculateProvision(totalTantiemes int, provision Provision) float64 {
	return float64(person.Tantieme) / float64(totalTantiemes) * provision.Amount
}

func (person *Person) CalculateLeft(totalTantiemes int, bills []Bill, provisions []Provision) float64 {
	var balance float64 = 0

	for _, bill := range bills {
		balance -= person.CalculateDue(totalTantiemes, bill)
	}

	for _, provision := range provisions {
		balance += person.CalculateProvision(totalTantiemes, provision)
	}

	return balance
}
