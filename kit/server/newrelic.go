package server

import (
	"github.com/newrelic/go-agent"
)

// NewNewrelicApplication tries to instantiate a newrelic application, otherwise returns error
func NewNewrelicApplication(newRelicConfig *NewRelicConfig) (newrelic.Application, error) {
	config := newrelic.NewConfig(newRelicConfig.NewRelicAppName, newRelicConfig.NewRelicLicenseKey)
	config.BetaToken = newRelicConfig.NewRelicBetaToken
	return newrelic.NewApplication(config)
}
