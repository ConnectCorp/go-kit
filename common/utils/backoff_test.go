package utils

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	count := 0

	err := Backoff(time.Nanosecond, 10, func() error {
		count += 1
		if count < 5 {
			return xerror.New("some-error")
		} else {
			return nil
		}
	})

	assert.Nil(t, err)
	assert.Equal(t, 5, count)

	count = 0

	err = Backoff(time.Nanosecond, 10, func() error {
		count += 1
		return xerror.New("some-error %v", count)
	})

	assert.Equal(t, "too many failed attempts: some-error 10", err.Error())
	assert.Equal(t, 10, count)
}

func TestMustBackoff(t *testing.T) {
	assert.Panics(t, func() {
		MustBackoff(time.Nanosecond, 1, func() error {
			return xerror.New("some-error")
		})
	})

}
