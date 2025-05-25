package domains

import (
	"html/template"
	"net/http"
	"strconv"
	"tanzia/helpers"
)

type Person struct {
	Name     string
	Tanzia int
}

func PersonHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserId(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("templates/edit-persons.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AddPersonHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetAuthenticatedUserId(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tanzia, _ := strconv.Atoi(r.FormValue("tanzia"))
	_, err = db.Exec("INSERT INTO persons (name, tanzia, userId) VALUES ($1, $2, $3)", r.FormValue("name"), tanzia, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#person_added", http.StatusFound)
}

func (person *Person) CalculateDue(totalTanzias int, bill Bill) float64 {
	return float64(person.Tanzia) / float64(totalTanzias) * bill.Amount
}

func (person *Person) CalculateProvision(totalTanzias int, provision Provision) float64 {
	return float64(person.Tanzia) / float64(totalTanzias) * provision.Amount
}

func (person *Person) CalculateLeft(totalTanzias int, bills []Bill, provisions []Provision) float64 {
	var Balance float64 = 0

	for _, bill := range bills {
		Balance -= person.CalculateDue(totalTanzias, bill)
	}

	for _, provision := range provisions {
		Balance += person.CalculateProvision(totalTanzias, provision)
	}

	return Balance
}
