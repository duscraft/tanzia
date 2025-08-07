package domains

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tanzia/apps/go/helpers"
)

type Bill struct {
	Label  string
	Amount float64
}

func AddBillHandler(w http.ResponseWriter, r *http.Request) {
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

	canUserCreateBill, err := helpers.CanUserCreateBill(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !canUserCreateBill {
		http.Error(w, "Free tier does not allow adding more bills", http.StatusForbidden)
		return
	}

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	_, err = db.Exec("INSERT INTO bills (label, amount, userId) VALUES ($1, $2, $3)", r.FormValue("label"), amount, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#bill_added", http.StatusFound)
}

func BillsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserID(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	// Example: return empty bills list as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]Bill{})
}
