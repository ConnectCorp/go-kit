package service

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCommonConfig(t *testing.T) {
	os.Setenv("TEST_JWT_KEY_ID", "test")
	os.Setenv("TEST_JWT_KEY_PUBLIC", "test")
	os.Setenv("TEST_JWT_ISSUER", "test")
	os.Setenv("TEST_JWT_AUDIENCE", "test")
	os.Setenv("TEST_TEST_PROXY", "test")
	os.Setenv("TEST_CHILD_VALUE", "test")

	type childConfig struct {
		CommonConfig
		ChildValue string `envconfig:"CHILD_VALUE" required:"true"`
	}

	config := &childConfig{}
	envconfig.MustProcess("TEST", config)
	assert.Equal(t, "test", config.CommonConfig.JWTKeyID)
	assert.Equal(t, "test", config.ChildValue)
}
