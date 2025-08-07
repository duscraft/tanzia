package domains

import (
	"encoding/json"
	"net/http"
	"tanzia/apps/go/helpers"
)

type DashboardData struct {
	Persons        []Person
	Bills          []Bill
	Provisions     []Provision
	TotalTantiemes int
	Balance        float64
}

func getDashboardData(userID string) DashboardData {
	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		panic(err)
	}

	personRows, _ := db.Query("SELECT name, tantieme FROM persons WHERE userId = $1", userID)
	billRows, _ := db.Query("SELECT label, amount FROM bills WHERE userId = $1", userID)
	provisionRows, _ := db.Query("SELECT label, amount FROM provisions WHERE userId = $1", userID)

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
	userID, ok := GetAuthenticatedUserID(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	data := getDashboardData(userID)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
