package server

import (
	"github.com/jmoiron/sqlx"
	"gopkg.in/redis.v3"
)

// DBHealthChecker is a HealthChecker that checks a DB connection.
type DBHealthChecker struct {
	db *sqlx.DB
}

// NewDBHealthChecker initializes a new DBHealthChecker.
func NewDBHealthChecker(db *sqlx.DB) *DBHealthChecker {
	return &DBHealthChecker{db: db}
}

// CheckHealth implements the HealthCheck interface.
func (d *DBHealthChecker) CheckHealth() error {
	return d.db.Ping()
}

// RedisHealthChecker is a HealthChecker that checks a Redis connection.
type RedisHealthChecker struct {
	redisClient *redis.Client
}

// NewRedisHealthChecker initializes a new RedisHealthChecker.
func NewRedisHealthChecker(redisClient *redis.Client) {
	&RedisHealthChecker{redisClient: redisClient}
}

// CheckHealth implements the HealthCheck interface.
func (r *RedisHealthChecker) CheckHealth() error {
	return r.redisClient.Ping().Err()
}
