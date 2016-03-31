package utils

import (
	"github.com/ibrt/go-xerror/xerror"
	"time"
)

const (
	defaultBaseDelay  = 50 * time.Millisecond
	defaultRetryCount = 10
)

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
	return xerror.Wrap(err)
}

func MustBackoff(baseDelay time.Duration, retryCount int, do func() error) {
	if err := Backoff(baseDelay, retryCount, do); err != nil {
		panic(err)
	}
}
