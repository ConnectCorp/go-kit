package utils

import (
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"time"
)

const (
	// ErrorTooManyFailedAttempts is returned when an operation fails retryCount times.
	ErrorTooManyFailedAttempts = "too many failed attempts"
)

const (
	defaultBaseDelay  = 50 * time.Millisecond
	defaultRetryCount = 10
)

// Backoff calls the given function at most retryCount times, increasing the baseDelay exponentially between calls.
func Backoff(baseDelay time.Duration, retryCount int, do func() error) error {
	var err error
	for i := 0; i < retryCount; i++ {
		err = do()
		if err == nil {
			return nil
		}
		time.Sleep(baseDelay)
		baseDelay *= 2
	}
	return xerror.Wrap(err, ErrorTooManyFailedAttempts)
}

// MustBackoff is like Backoff but panics on final error.
func MustBackoff(baseDelay time.Duration, retryCount int, do func() error) {
	if err := Backoff(baseDelay, retryCount, do); err != nil {
		panic(err)
	}
}
