package domains

import (
	"html/template"
	"net/http"
	"strconv"
	"tanzia/helpers"
)

type Bill struct {
	Label  string
	Amount float64
}

func AddBillHandler(w http.ResponseWriter, r *http.Request) {
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

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	_, err = db.Exec("INSERT INTO bills (label, amount, userId) VALUES ($1, $2, $3)", r.FormValue("label"), amount, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#bill_added", http.StatusFound)
}

func BillsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserId(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("templates/edit-bills.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
