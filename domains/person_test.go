package domains

import (
	"testing"
	"time"
)

func TestCalculateDue(t *testing.T) {
	person := Person{
		Name:     "John Doe",
		Tantieme: 2,
	}
	bill := Bill{
		Label:       "Electricity",
		Amount:      1000,
		BillingDate: time.Now(),
	}
	totalTantiemes := 5

	expectedDue := 400.0 // 2/5 * 1000
	calculatedDue := person.calculateDue(totalTantiemes, bill)

	if calculatedDue != expectedDue {
		t.Errorf("Expected due %.2f, but got %.2f", expectedDue, calculatedDue)
	}
}
