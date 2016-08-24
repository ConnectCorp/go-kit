package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
)

// CommonConfig contains common configuration keys for Connect microservices.
type CommonConfig struct {
	TestProxy utils.EnvURL `envconfig:"TEST_PROXY"`
	TestMock utils.EnvURL `envconfig:"TEST_MOCK"`
}

// TokenVerifierConfig contains configuration keys for services that verify tokens.
type TokenVerifierConfig struct {
	JWTKeyID     string          `envconfig:"JWT_KEY_ID" required:"true"`
	JWTKeyPublic utils.EnvBinary `envconfig:"JWT_KEY_PUBLIC" required:"true"`
	JWTIssuer    string          `envconfig:"JWT_ISSUER" required:"true"`
	JWTAudience  string          `envconfig:"JWT_AUDIENCE" required:"true"`
}

// TokenIssuerConfig contains configuration keys for services that issue tokens.
type TokenIssuerConfig struct {
	TokenVerifierConfig
	JWTKeyPrivate utils.EnvBinary `envconfig:"JWT_KEY_PRIVATE" required:"true"`
}

// PusherConfig contains configuration keys for services that use Pusher.
type PusherConfig struct {
	PusherSpec string `envconfig:"PUSHER_SPEC" required:"true"`
}

// NexmoConfig contains configuration keys for services that use Nexmo.
type NexmoConfig struct {
	NexmoAPIKey    string `envconfig:"NEXMO_API_KEY" required:"true"`
	NexmoAPISecret string `envconfig:"NEXMO_API_SECRET" required:"true"`
}

// DogstatsdConfig contains configuration keys for services that use Dogstatsd.
type DogstatsdConfig struct {
	DogstatsdSpec string `envconfig:"DOGSTATSD_SPEC" required:"true"`
}

// NewRelicConfig contains optional configuration keys for newrelic
type NewRelicConfig struct {
	NewRelicAppName    string `envconfig:"NEW_RELIC_APP_NAME"`
	NewRelicLicenseKey string `envconfig:"NEW_RELIC_LICENSE_KEY"`
	NewRelicBetaToken  string `envconfig:"NEW_RELIC_BETA_TOKEN"`
}
