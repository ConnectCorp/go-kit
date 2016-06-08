package utils

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testConfig struct {
	Binary EnvBinary `envconfig:"BINARY" required:"true"`
	URL    EnvURL    `envconfig:"URL" required:"true"`
}

func TestEnv(t *testing.T) {
	os.Setenv("TEST_BINARY", "dGVzdA==")
	os.Setenv("TEST_URL", "http://localhost")

	config := &testConfig{}
	envconfig.MustProcess("TEST", config)

	assert.Equal(t, []byte("test"), []byte(config.Binary))
	assert.Equal(t, "http://localhost", config.URL.URL.String())
}
