package RateLimiter

import (
	"context"
	"sync"
	"time"

	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
)

// RateLimit provides rate limiting for the GitHub API
type RateLimit struct {
	mu           sync.Mutex
	requestCount int
	lastReset    time.Time
	maxRequests  int
	resetPeriod  time.Duration
	logger       *logger.Logger
}

// NewRateLimit creates a new RateLimit instance with default settings
// GitHub API allows 5000 requests per hour for authenticated users
func NewRateLimit(logger *logger.Logger) *RateLimit {
	return &RateLimit{
		requestCount: 0,
		lastReset:    time.Now(),
		maxRequests:  5000,      // 5000 requests
		resetPeriod:  time.Hour, // per hour
		logger:       logger,
	}
}

// Wait waits if necessary to comply with API rate limits
func (r *RateLimit) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the reset period has passed
	elapsed := time.Since(r.lastReset)
	if elapsed >= r.resetPeriod {
		r.logger.Debug("Rate limit period reset")
		r.requestCount = 0
		r.lastReset = time.Now()
		return nil
	}

	// If we've reached the request limit
	if r.requestCount >= r.maxRequests {
		// Calculate time until reset
		waitTime := r.resetPeriod - elapsed
		r.logger.Warn("Rate limit reached, waiting for %v", waitTime)

		// Create a timer for waiting
		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		// Wait either for the rate limit to reset or the context to be canceled
		select {
		case <-timer.C:
			// Reset counter after waiting
			r.requestCount = 0
			r.lastReset = time.Now()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Increment request counter
	r.requestCount++
	r.logger.Debug("API request count: %d/%d", r.requestCount, r.maxRequests)

	return nil
}

// GetRemainingRequests returns the number of remaining requests
func (r *RateLimit) GetRemainingRequests() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the reset period has passed
	if time.Since(r.lastReset) >= r.resetPeriod {
		r.requestCount = 0
		r.lastReset = time.Now()
		return r.maxRequests
	}

	return r.maxRequests - r.requestCount
}

// GetResetTime returns the time until the rate limit resets
func (r *RateLimit) GetResetTime() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastReset)
	if elapsed >= r.resetPeriod {
		return 0
	}
	return r.resetPeriod - elapsed
}

// UpdateLimits updates the rate limits based on GitHub API response
func (r *RateLimit) UpdateLimits(remaining int, resetTime time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestCount = r.maxRequests - remaining
	r.lastReset = resetTime.Add(-r.resetPeriod)

	r.logger.Debug("Updated rate limits: remaining=%d, reset=%v",
		remaining, resetTime.Format(time.RFC3339))
}
