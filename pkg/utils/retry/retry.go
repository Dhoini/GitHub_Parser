package retry

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"github.com/google/go-github/v39/github"
	"go.mongodb.org/mongo-driver/mongo"
)

// WithRetry executes the operation with retries
func WithRetry(ctx context.Context, maxRetries int, initialBackoff time.Duration, logger *logger.Logger, operation func() error) error {
	var err error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Check if the error is retryable
		if IsRetryable(err) {
			logger.Warn("Operation failed with retryable error (attempt %d/%d): %v. Retrying in %v...",
				attempt+1, maxRetries, err, backoff)

			// Create a timer for waiting before the next attempt
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
				// Increase waiting time with exponential backoff
				backoff = backoff * 2
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		} else {
			// If the error is not retryable, return immediately
			return err
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, err)
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	// Check error type
	if err == nil {
		return false
	}

	// Temporary network errors
	if netErr, ok := err.(net.Error); ok && (netErr.Temporary() || netErr.Timeout()) {
		return true
	}

	// GitHub API errors that make sense to retry (e.g., 500, 502, 503)
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		switch ghErr.Response.StatusCode {
		case 500, 502, 503, 504:
			return true
		}
	}

	// Temporary MongoDB errors
	if mongoErr, ok := err.(mongo.ServerError); ok {
		code := mongoErr.Code()
		return code == 11600 || // Interrupted
			code == 11601 || // Interrupted At Shutdown
			code == 13435 || // Not Master
			code == 10107 || // NotMaster
			code == 13436 // Not Master No Slave Ok
	}

	return false
}
