package domains

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/duscraft/tanzia/lib/helpers"
	"github.com/go-pdf/fpdf"
)

func ExportPDFHandler(w http.ResponseWriter, r *http.Request) {
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

	isPremium, err := helpers.IsUserPremium(db, userID)
	if err != nil {
		log.Printf("Error checking premium status: %v", err)
		http.Error(w, "Could not verify subscription status", http.StatusInternalServerError)
		return
	}

	if !isPremium {
		http.Error(w, "Premium subscription required for PDF export", http.StatusForbidden)
		return
	}

	data, err := getDashboardData(userID)
	if err != nil {
		log.Printf("Error getting dashboard data: %v", err)
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 12, "Tanzia - Rapport de Copropriete")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Date: %s", time.Now().Format("02/01/2006")))
	pdf.Ln(12)

	if len(data.Provisions) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Provisions de Charges")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		colWidths := []float64{80, 50, 50}
		pdf.CellFormat(colWidths[0], 7, "Libelle", "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 7, "Montant", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[2], 7, "Total Tantiemes", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		for i, provision := range data.Provisions {
			fill := i%2 == 0
			if fill {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(colWidths[0], 6, provision.Label, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 6, fmt.Sprintf("%.2f EUR", provision.Amount), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(colWidths[2], 6, fmt.Sprintf("%d", data.TotalTantiemes), "1", 1, "R", fill, 0, "")
		}

		pdf.Ln(10)
	}

	if len(data.Bills) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Travaux et Charges")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		colWidths := []float64{80, 50, 50}
		pdf.CellFormat(colWidths[0], 7, "Libelle", "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 7, "Montant", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[2], 7, "Total Tantiemes", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		for i, bill := range data.Bills {
			fill := i%2 == 0
			if fill {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(colWidths[0], 6, bill.Label, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 6, fmt.Sprintf("%.2f EUR", bill.Amount), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(colWidths[2], 6, fmt.Sprintf("%d", data.TotalTantiemes), "1", 1, "R", fill, 0, "")
		}

		pdf.Ln(10)
	}

	if len(data.Persons) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Coproprietaires")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		colWidths := []float64{60, 40, 40, 40}
		pdf.CellFormat(colWidths[0], 7, "Nom", "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 7, "Tantiemes", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[2], 7, "Part (%)", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[3], 7, "Solde", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		for i, person := range data.Persons {
			fill := i%2 == 0
			if fill {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			percentage := 0.0
			if data.TotalTantiemes > 0 {
				percentage = float64(person.Tantieme) / float64(data.TotalTantiemes) * 100
			}
			balance := person.CalculateLeft(data.TotalTantiemes, data.Bills, data.Provisions)

			pdf.CellFormat(colWidths[0], 6, person.Name, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 6, fmt.Sprintf("%d", person.Tantieme), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(colWidths[2], 6, fmt.Sprintf("%.2f%%", percentage), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(colWidths[3], 6, fmt.Sprintf("%.2f EUR", balance), "1", 1, "R", fill, 0, "")
		}

		pdf.Ln(10)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(100, 8, "Solde Global", "1", 0, "L", true, 0, "")
	pdf.CellFormat(80, 8, fmt.Sprintf("%.2f EUR", data.Balance), "1", 1, "R", true, 0, "")

	pdf.Ln(20)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Document genere automatiquement par Tanzia le %s", time.Now().Format("02/01/2006 15:04")))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=tanzia-rapport-%s.pdf", time.Now().Format("2006-01-02")))

	if err := pdf.Output(w); err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
	}
}
