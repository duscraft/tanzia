package domains

import (
	"html/template"
	"net/http"
	"strconv"
	"tantieme/helpers"
)

type Bill struct {
	Label  string
	Amount float64
}

func AddBillHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := GetAuthenticatedUsername(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	db := helpers.GetConnectionManager().GetConnection("sqlite3", username)

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	_, err := db.Exec("INSERT INTO bills (label, amount) VALUES (?, ?)", r.FormValue("label"), amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#bill_added", http.StatusFound)
}

func BillsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUsername(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("templates/edit-Bills.html")
	err := t.Execute(w, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
