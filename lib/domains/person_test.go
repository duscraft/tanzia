package domains

import (
	"testing"
)

func TestCalculateDue(t *testing.T) {
	person := Person{
		Name:     "John Doe",
		Tantieme: 2,
	}
	bill := Bill{
		Label:  "Electricity",
		Amount: 1000,
	}
	totalTantiemes := 5

	expectedDue := 400.0 // 2/5 * 1000
	calculatedDue := person.CalculateDue(totalTantiemes, bill)

	if calculatedDue != expectedDue {
		t.Errorf("Expected due %.2f, but got %.2f", expectedDue, calculatedDue)
	}
}

func TestCalculateProvision(t *testing.T) {
	person := Person{
		Name:     "John Doe",
		Tantieme: 2,
	}
	provision := Provision{
		Label:  "Trimestre 1 2025",
		Amount: 1000,
	}
	totalTantiemes := 5

	expectedProvision := 400.0 // 2/5 * 1000
	calculatedProvision := person.CalculateProvision(totalTantiemes, provision)

	if calculatedProvision != expectedProvision {
		t.Errorf("Expected provision %.2f, but got %.2f", expectedProvision, calculatedProvision)
	}
}

func TestCalculateLeft(t *testing.T) {
	person := Person{
		Name:     "John Doe",
		Tantieme: 5,
	}

	provisions := []Provision{
		{
			Label:  "Trimestre 1 2025",
			Amount: 2000,
		},
		{
			Label:  "Trimestre 2 2025",
			Amount: 2200,
		},
	}

	bills := []Bill{
		{
			Label:  "Travaux 1 2025",
			Amount: 1800,
		},
		{
			Label:  "Travaux 2 2025",
			Amount: 2600,
		},
	}

	expectedLeft := -100.0
	calculatedLeft := person.CalculateLeft(10, bills, provisions)

	if calculatedLeft != expectedLeft {
		t.Errorf("Expected balance of %.2f, but got %.2f", expectedLeft, calculatedLeft)
	}
}
