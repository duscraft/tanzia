package domains

import (
	"html/template"
	"net/http"
	"strconv"
	"tantieme/helpers"
)

type Person struct {
	Name     string
	Tantieme int
}

func PersonHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUsername(w, r)

	if ok != true {
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
	username, ok := GetAuthenticatedUsername(w, r)

	if ok != true {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)
	tantieme, _ := strconv.Atoi(r.FormValue("tantieme"))
	_, err := db.Exec("INSERT INTO persons (name, tantieme) VALUES (?, ?)", r.FormValue("name"), tantieme)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#person_added", http.StatusFound)
}

func (person *Person) CalculateDue(totalTantiemes int, bill Bill) float64 {
	return float64(person.Tantieme) / float64(totalTantiemes) * float64(bill.Amount)
}
