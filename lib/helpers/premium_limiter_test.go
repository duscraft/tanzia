package helpers

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIsUserPremium_PremiumUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "user-123"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(true))

	isPremium, err := IsUserPremium(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !isPremium {
		t.Error("Expected user to be premium")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestIsUserPremium_FreeUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "user-456"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))

	isPremium, err := IsUserPremium(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if isPremium {
		t.Error("Expected user to not be premium")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestIsUserPremium_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "nonexistent-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = IsUserPremium(db, userID)
	if err == nil {
		t.Error("Expected error for nonexistent user")
	}
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found' error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateProvision_PremiumUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "premium-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(true))

	canCreate, err := CanUserCreateProvision(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Premium user should always be able to create provisions")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateProvision_FreeUserUnderLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM provisions WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	canCreate, err := CanUserCreateProvision(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Free user under limit should be able to create provisions")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateProvision_FreeUserAtLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user-at-limit"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM provisions WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(FreeTierProvisionLimit))

	canCreate, err := CanUserCreateProvision(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if canCreate {
		t.Error("Free user at limit should not be able to create provisions")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateBill_PremiumUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "premium-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(true))

	canCreate, err := CanUserCreateBill(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Premium user should always be able to create bills")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateBill_FreeUserUnderLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bills WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	canCreate, err := CanUserCreateBill(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Free user under limit should be able to create bills")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreateBill_FreeUserAtLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user-at-limit"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bills WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(FreeTierBillLimit))

	canCreate, err := CanUserCreateBill(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if canCreate {
		t.Error("Free user at limit should not be able to create bills")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreatePerson_PremiumUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "premium-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(true))

	canCreate, err := CanUserCreatePerson(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Premium user should always be able to create persons")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreatePerson_FreeUserUnderLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM persons WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	canCreate, err := CanUserCreatePerson(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("Free user under limit should be able to create persons")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCanUserCreatePerson_FreeUserAtLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	userID := "free-user-at-limit"
	mock.ExpectQuery("SELECT is_premium FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"is_premium"}).AddRow(false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM persons WHERE userId = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(FreeTierPersonLimit))

	canCreate, err := CanUserCreatePerson(db, userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if canCreate {
		t.Error("Free user at limit should not be able to create persons")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestFreeTierLimitConstants(t *testing.T) {
	if FreeTierPersonLimit != 5 {
		t.Errorf("Expected FreeTierPersonLimit to be 5, got %d", FreeTierPersonLimit)
	}
	if FreeTierBillLimit != 5 {
		t.Errorf("Expected FreeTierBillLimit to be 5, got %d", FreeTierBillLimit)
	}
	if FreeTierProvisionLimit != 10 {
		t.Errorf("Expected FreeTierProvisionLimit to be 10, got %d", FreeTierProvisionLimit)
	}
}
