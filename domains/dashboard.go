package domains

import (
	"html/template"
	"net/http"
	"tantieme/helpers"
)

type DashboardData struct {
	Persons        []Person
	Bills          []Bill
	Provisions     []Provision
	TotalTantiemes int
	Balance        float64
}

func getDashboardData(username string) DashboardData {
	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)

	personRows, _ := db.Query("SELECT * FROM persons")
	billRows, _ := db.Query("SELECT * FROM bills")
	provisionRows, _ := db.Query("SELECT * FROM provisions")

	var Persons []Person
	var Bills []Bill
	var Provisions []Provision
	var Balance float64
	TotalTantiemes := 0

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
		Balance -= bill.Amount
		Bills = append(Bills, bill)
	}

	for provisionRows.Next() {
		var provision Provision
		err := provisionRows.Scan(&provision.Label, &provision.Amount)
		if err != nil {
			panic(err)
		}
		Balance += provision.Amount
		Provisions = append(Provisions, provision)
	}

	return DashboardData{
		Persons,
		Bills,
		Provisions,
		TotalTantiemes,
		Balance,
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := GetAuthenticatedUsername(w, r)

	if !ok {
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
