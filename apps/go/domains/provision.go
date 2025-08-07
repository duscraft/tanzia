package domains

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tanzia/apps/go/helpers"
)

type Provision struct {
	Label  string
	Amount float64
}

func AddProvisionHandler(w http.ResponseWriter, r *http.Request) {
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

	canUserCreateProvision, err := helpers.CanUserCreateProvision(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !canUserCreateProvision {
		http.Error(w, "Free tier does not allow adding more provision", http.StatusForbidden)
		return
	}

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	_, err = db.Exec("INSERT INTO provisions (label, amount, userId) VALUES ($1, $2, $3)", r.FormValue("label"), amount, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/dashboard#provision_added", http.StatusFound)
}

func ProvisionsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := GetAuthenticatedUserID(w, r)

	if !ok {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	// Example: return empty provisions list as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]Provision{})
}
