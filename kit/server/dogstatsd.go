package server

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"time"
)

const (
	dogstatsdFlushInterval = time.Second
)

// InitDogstatsdEmitter initializes a new dogstatsd.Emitter, if configured.
func InitDogstatsdEmitter(prefix string, logger log.Logger, cfg *DogstatsdConfig) *dogstatsd.Emitter {
	if cfg.DogstatsdSpec == "" {
		return nil
	}
	return dogstatsd.NewEmitter("udp", cfg.DogstatsdSpec, prefix, dogstatsdFlushInterval, logger)
}
