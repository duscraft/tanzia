package helpers

import (
	"testing"
	"time"
)

func TestRateLimiterBasicFlow(t *testing.T) {
	rl := &RateLimiter{
		attempts: make(map[string]*loginAttempt),
	}

	email := "test@example.com"

	if rl.IsLocked(email) {
		t.Error("New email should not be locked")
	}

	if rl.GetRemainingAttempts(email) != maxLoginAttempts {
		t.Errorf("Expected %d attempts, got %d", maxLoginAttempts, rl.GetRemainingAttempts(email))
	}
}

func TestRateLimiterLockout(t *testing.T) {
	rl := &RateLimiter{
		attempts: make(map[string]*loginAttempt),
	}

	email := "locktest@example.com"

	for i := 0; i < maxLoginAttempts-1; i++ {
		locked := rl.RecordFailedAttempt(email)
		if locked {
			t.Errorf("Should not be locked after %d attempts", i+1)
		}
	}

	locked := rl.RecordFailedAttempt(email)
	if !locked {
		t.Error("Should be locked after max attempts")
	}

	if !rl.IsLocked(email) {
		t.Error("IsLocked should return true")
	}

	if rl.GetRemainingAttempts(email) != 0 {
		t.Error("Should have 0 remaining attempts when locked")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := &RateLimiter{
		attempts: make(map[string]*loginAttempt),
	}

	email := "reset@example.com"

	rl.RecordFailedAttempt(email)
	rl.RecordFailedAttempt(email)

	if rl.GetRemainingAttempts(email) != maxLoginAttempts-2 {
		t.Error("Should have recorded 2 failed attempts")
	}

	rl.ResetAttempts(email)

	if rl.GetRemainingAttempts(email) != maxLoginAttempts {
		t.Error("Should have reset to max attempts")
	}
}

func TestRateLimiterLockoutDuration(t *testing.T) {
	rl := &RateLimiter{
		attempts: make(map[string]*loginAttempt),
	}

	email := "duration@example.com"

	rl.attempts[email] = &loginAttempt{
		count:     maxLoginAttempts,
		firstFail: time.Now().Add(-1 * time.Hour),
		lockedAt:  time.Now(),
	}

	remaining := rl.GetLockoutRemaining(email)
	if remaining <= 0 || remaining > lockoutDuration {
		t.Errorf("Lockout remaining should be between 0 and %v, got %v", lockoutDuration, remaining)
	}
}
