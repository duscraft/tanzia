package domains

import "net/http"

type Person struct {
	Name     string
	Tantieme int
}

func PersonHandler(w http.ResponseWriter, r *http.Request) {

}

func (person *Person) calculateDue(totalTantiemes int, bill Bill) float64 {
	return float64(person.Tantieme) / float64(totalTantiemes) * float64(bill.Amount)
}
