package domains

import (
	"html/template"
	"net/http"
	"tantieme/helpers"
)

type DashboardData struct {
	persons []Person
	bills   []Bill
}

func getDashboardData(username string) DashboardData {
	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)

	personRows, _ := db.Query("SELECT * FROM persons")
	billRows, _ := db.Query("SELECT * FROM bills")

	var persons []Person
	var bills []Bill

	for personRows.Next() {
		var person Person
		err := personRows.Scan(&person.Name, &person.Tantieme)
		if err != nil {
			panic(err)
		}
		persons = append(persons, person)
	}

	for billRows.Next() {
		var bill Bill
		err := billRows.Scan(&bill.Label, &bill.Amount, &bill.BillingDate)
		if err != nil {
			panic(err)
		}
		bills = append(bills, bill)
	}

	return DashboardData{
		persons,
		bills,
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
