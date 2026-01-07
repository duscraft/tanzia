package helpers

import (
	"sync"
	"time"
)

type loginAttempt struct {
	count     int
	firstFail time.Time
	lockedAt  time.Time
}

type RateLimiter struct {
	attempts map[string]*loginAttempt
	mu       sync.RWMutex
}

const (
	maxLoginAttempts   = 5
	lockoutDuration    = 15 * time.Minute
	attemptWindowReset = 15 * time.Minute
)

var rateLimiter *RateLimiter

func GetRateLimiter() *RateLimiter {
	if rateLimiter == nil {
		rateLimiter = &RateLimiter{
			attempts: make(map[string]*loginAttempt),
		}
		go rateLimiter.cleanup()
	}
	return rateLimiter
}

func (rl *RateLimiter) IsLocked(identifier string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	attempt, exists := rl.attempts[identifier]
	if !exists {
		return false
	}

	if !attempt.lockedAt.IsZero() {
		if time.Since(attempt.lockedAt) < lockoutDuration {
			return true
		}
	}

	return false
}

func (rl *RateLimiter) GetLockoutRemaining(identifier string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	attempt, exists := rl.attempts[identifier]
	if !exists {
		return 0
	}

	if attempt.lockedAt.IsZero() {
		return 0
	}

	remaining := lockoutDuration - time.Since(attempt.lockedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (rl *RateLimiter) RecordFailedAttempt(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	attempt, exists := rl.attempts[identifier]
	if !exists {
		rl.attempts[identifier] = &loginAttempt{
			count:     1,
			firstFail: time.Now(),
		}
		return false
	}

	if time.Since(attempt.firstFail) > attemptWindowReset {
		attempt.count = 1
		attempt.firstFail = time.Now()
		attempt.lockedAt = time.Time{}
		return false
	}

	attempt.count++

	if attempt.count >= maxLoginAttempts {
		attempt.lockedAt = time.Now()
		return true
	}

	return false
}

func (rl *RateLimiter) ResetAttempts(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.attempts, identifier)
}

func (rl *RateLimiter) GetRemainingAttempts(identifier string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	attempt, exists := rl.attempts[identifier]
	if !exists {
		return maxLoginAttempts
	}

	if time.Since(attempt.firstFail) > attemptWindowReset {
		return maxLoginAttempts
	}

	remaining := maxLoginAttempts - attempt.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, attempt := range rl.attempts {
			if !attempt.lockedAt.IsZero() && now.Sub(attempt.lockedAt) > lockoutDuration {
				delete(rl.attempts, key)
			} else if attempt.lockedAt.IsZero() && now.Sub(attempt.firstFail) > attemptWindowReset {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}
