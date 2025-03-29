package domains

import (
	"html/template"
	"net/http"
	"tantieme/helpers"
)

type DashboardData struct {
	Persons        []Person
	Bills          []Bill
	TotalTantiemes int
}

func getDashboardData(username string) DashboardData {
	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)

	personRows, _ := db.Query("SELECT * FROM persons")
	billRows, _ := db.Query("SELECT * FROM bills")

	var Persons []Person
	var Bills []Bill
	var TotalTantiemes int = 0

	for personRows.Next() {
		var person Person
		err := personRows.Scan(&person.Name, &person.Tantieme)
		if err != nil {
			panic(err)
		}
		Persons = append(Persons, person)
		TotalTantiemes += person.Tantieme
	}

	for billRows.Next() {
		var bill Bill
		err := billRows.Scan(&bill.Label, &bill.Amount)
		if err != nil {
			panic(err)
		}
		Bills = append(Bills, bill)
	}

	return DashboardData{
		Persons,
		Bills,
		TotalTantiemes,
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := GetAuthenticatedUsername(w, r)

	if ok != true {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	data := getDashboardData(username)

	t, _ := template.ParseFiles("templates/dashboard.html")
	err := t.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
