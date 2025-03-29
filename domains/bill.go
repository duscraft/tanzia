package domains

import (
	"net/http"
	"time"
)

type Bill struct {
	Label       string
	Amount      int
	BillingDate time.Time
}

func BillsHandler(w http.ResponseWriter, r *http.Request) {

}
