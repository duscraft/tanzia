package domains

import (
	"html/template"
	"log"
	"net/http"

	"github.com/duscraft/tanzia/lib/helpers"
)

type DashboardData struct {
	Persons        []Person
	Bills          []Bill
	Provisions     []Provision
	TotalTantiemes int
	Balance        float64
	IsPremium      bool
}

func getDashboardData(userID string) (DashboardData, error) {
	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		return DashboardData{}, err
	}

	personRows, err := db.Query("SELECT name, tantieme FROM persons WHERE userId = $1", userID)
	if err != nil {
		return DashboardData{}, err
	}
	defer func() { _ = personRows.Close() }()

	billRows, err := db.Query("SELECT label, amount FROM bills WHERE userId = $1", userID)
	if err != nil {
		return DashboardData{}, err
	}
	defer func() { _ = billRows.Close() }()

	provisionRows, err := db.Query("SELECT label, amount FROM provisions WHERE userId = $1", userID)
	if err != nil {
		return DashboardData{}, err
	}
	defer func() { _ = provisionRows.Close() }()

	var persons []Person
	var bills []Bill
	var provisions []Provision
	var balance float64
	totalTantiemes := 0

	for personRows.Next() {
		var person Person
		if err := personRows.Scan(&person.Name, &person.Tantieme); err != nil {
			return DashboardData{}, err
		}
		persons = append(persons, person)
		totalTantiemes += person.Tantieme
	}
	if err := personRows.Err(); err != nil {
		return DashboardData{}, err
	}

	for billRows.Next() {
		var bill Bill
		if err := billRows.Scan(&bill.Label, &bill.Amount); err != nil {
			return DashboardData{}, err
		}
		balance -= bill.Amount
		bills = append(bills, bill)
	}
	if err := billRows.Err(); err != nil {
		return DashboardData{}, err
	}

	for provisionRows.Next() {
		var provision Provision
		if err := provisionRows.Scan(&provision.Label, &provision.Amount); err != nil {
			return DashboardData{}, err
		}
		balance += provision.Amount
		provisions = append(provisions, provision)
	}
	if err := provisionRows.Err(); err != nil {
		return DashboardData{}, err
	}

	isPremium, err := helpers.IsUserPremium(db, userID)
	if err != nil {
		log.Printf("Warning: could not check premium status for user %s: %v", userID, err)
		isPremium = false
	}

	return DashboardData{
		Persons:        persons,
		Bills:          bills,
		Provisions:     provisions,
		TotalTantiemes: totalTantiemes,
		Balance:        balance,
		IsPremium:      isPremium,
	}, nil
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	data, err := getDashboardData(userID)
	if err != nil {
		log.Printf("Error getting dashboard data: %v", err)
		http.Error(w, "Failed to load dashboard data", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("lib/templates/dashboard.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
