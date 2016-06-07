package utils

import (
	"encoding/base64"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/url"
)

const (
	// ErrorInvalidEnv is returned when an environment variable cannot be parsed as binary or URL.
	ErrorInvalidEnv = "invalid env variable"
)

// EnvBinary describes a base-64 encoded binary environment variable value.
type EnvBinary []byte

// Decode implements the envconfig.Decoder interface.
func (b *EnvBinary) Decode(s string) error {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return xerror.Wrap(err, ErrorInvalidEnv)
	}
	*b = EnvBinary(decoded)
	return nil
}

// EnvURL describes a URL environment variable value.
type EnvURL struct {
	URL *url.URL
}

// Decode implements the envconfig.Decoder interface.
func (u *EnvURL) Decode(s string) error {
	decoded, err := url.Parse(s)
	if err != nil {
		return xerror.Wrap(err, ErrorInvalidEnv)
	}
	u.URL = decoded
	return nil
}
