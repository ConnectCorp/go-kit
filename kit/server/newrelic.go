package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/newrelic/go-agent"
)

// NewNewrelicApplication tries to instantiate a newrelic application, otherwise returns error.
// On error, it returns a mock application in case the client just decides to log.
func NewNewrelicApplication(newRelicConfig *NewRelicConfig) (newrelic.Application, error) {
	config := newrelic.NewConfig(newRelicConfig.NewRelicAppName, newRelicConfig.NewRelicLicenseKey)
	config.BetaToken = newRelicConfig.NewRelicBetaToken
	app, err := newrelic.NewApplication(config)
	if err != nil {
		app = &utils.NoopNewrelicApplication{}
	}
	return app, err
}
