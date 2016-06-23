package service

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
)

// CommonConfig contains common configuration keys for Connect microservices.
type CommonConfig struct {
	JWTKeyID     string          `envconfig:"JWT_KEY_ID" required:"true"`
	JWTKeyPublic utils.EnvBinary `envconfig:"JWT_KEY_PUBLIC" required:"true"`
	JWTIssuer    string          `envconfig:"JWT_ISSUER" required:"true"`
	JWTAudience  string          `envconfig:"JWT_AUDIENCE" required:"true"`
	TestProxy    utils.EnvURL    `envconfig:"TEST_PROXY"`
}
