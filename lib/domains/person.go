package domains

import (
	"github.com/duscraft/tanzia/lib/helpers"
	"html/template"
	"net/http"
	"strconv"
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

	t, _ := template.ParseFiles("lib/templates/edit-persons.html")
	err := t.Execute(w, nil)
	if err != nil {
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
		http.Error(w, "Free tier does not allow adding more persons", http.StatusForbidden)
		return
	}

	tantieme, _ := strconv.Atoi(r.FormValue("tantieme"))
	_, err = db.Exec("INSERT INTO persons (name, tantieme, userId) VALUES ($1, $2, $3)", r.FormValue("name"), tantieme, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	var Balance float64 = 0

	for _, bill := range bills {
		Balance -= person.CalculateDue(totalTantiemes, bill)
	}

	for _, provision := range provisions {
		Balance += person.CalculateProvision(totalTantiemes, provision)
	}

	return Balance
}
