package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/jmoiron/sqlx"
	"time"
)

const (
	baseDBInitDelay     = 25 * time.Millisecond
	maxDBInitRetryCount = 10
)

// MustInitDB initializes a DB connection, or panics.
func MustInitDB(driver, spec string) *sqlx.DB {
	var db *sqlx.DB
	var err error

	utils.MustBackoff(baseDBInitDelay, maxDBInitRetryCount, func() error {
		db, err = sqlx.Connect(driver, spec)
		return err
	})

	return db
}
