package helpers

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	FREE_TIER_PROVISION_LIMIT = 10
	FREE_TIER_BILL_LIMIT      = 5
	FREE_TIER_PERSON_LIMIT    = 5
)

func IsUserPremium(db *sql.DB, userID string) (bool, error) {
	var isPremium bool
	query := "SELECT is_premium FROM users WHERE id = $1"
	err := db.QueryRow(query, userID).Scan(&isPremium)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user not found")
		}
		return false, fmt.Errorf("error checking premium status: %w", err)
	}
	return isPremium, nil
}

func CanUserCreateProvision(db *sql.DB, userID string) (bool, error) {
	isPremium, _ := IsUserPremium(db, userID)
	if isPremium {
		return true, nil
	}
	var count int
	query := "SELECT COUNT(*) FROM provisions WHERE userId = $1"
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking provision count: %w", err)
	}

	return count < FREE_TIER_PROVISION_LIMIT, nil
}

func CanUserCreateBill(db *sql.DB, userID string) (bool, error) {
	isPremium, _ := IsUserPremium(db, userID)
	if isPremium {
		return true, nil
	}
	var count int
	query := "SELECT COUNT(*) FROM bills WHERE userId = $1"
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking bill count: %w", err)
	}

	return count < FREE_TIER_BILL_LIMIT, nil
}

func CanUserCreatePerson(db *sql.DB, userID string) (bool, error) {
	isPremium, _ := IsUserPremium(db, userID)
	if isPremium {
		return true, nil
	}
	var count int
	query := "SELECT COUNT(*) FROM persons WHERE userId = $1"
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking person count: %w", err)
	}

	return count < FREE_TIER_PERSON_LIMIT, nil
}
