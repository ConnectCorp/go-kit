package server

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCommonConfig(t *testing.T) {
	os.Setenv("TEST_TEST_PROXY", "test")
	os.Setenv("TEST_CHILD_VALUE", "test")

	type childConfig struct {
		CommonConfig
		ChildValue string `envconfig:"CHILD_VALUE" required:"true"`
	}

	config := &childConfig{}
	envconfig.MustProcess("TEST", config)
	assert.Equal(t, "test", config.CommonConfig.TestProxy.URL.String())
	assert.Equal(t, "test", config.ChildValue)
}
