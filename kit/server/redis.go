package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"gopkg.in/redis.v3"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	errorInvalidRedisSpec  = "invalid Redis spec"
	redisMaxRetries        = 3
	redisReadTimeout       = 10 * time.Second
	redisWriteTimeout      = 10 * time.Millisecond
	redisConnPoolSize      = 25
	baseRedisInitDelay     = 25 * time.Millisecond
	maxRedisInitRetryCount = 10
)

// MustInitRedis initializes a Redis client, or panics.
func MustInitRedis(spec *url.URL) *redis.Client {
	redisClient := redis.NewClient(mustParseRedisSpec(spec))

	utils.MustBackoff(baseRedisInitDelay, maxRedisInitRetryCount, func() error {
		return redisClient.Ping().Err()
	})

	return redisClient
}

func mustParseRedisSpec(spec *url.URL) *redis.Options {
	if spec.Scheme != "redis" {
		panic(xerror.New(errorInvalidRedisSpec))
	}

	options := &redis.Options{
		Addr:         spec.Host,
		MaxRetries:   redisMaxRetries,
		ReadTimeout:  redisReadTimeout,
		WriteTimeout: redisWriteTimeout,
		PoolSize:     redisConnPoolSize,
	}

	if spec.User != nil {
		if pwd, ok := spec.User.Password(); ok {
			options.Password = pwd
		}
	}

	if db := strings.TrimLeft(spec.Path, "/"); len(db) > 0 {
		dbn, err := strconv.ParseInt(db, 10, 64)
		if err != nil {
			panic(xerror.Wrap(err, errorInvalidRedisSpec))
		}
		options.DB = dbn
	}

	return options
}
