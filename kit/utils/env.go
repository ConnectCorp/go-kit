package utils

import (
	"encoding/base64"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/url"
)

const (
	ErrorInvalidEnv = "invalid env variable"
)

type EnvBinary []byte

func (b *EnvBinary) Decode(s string) error {
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		*b = EnvBinary(decoded)
		return nil
	} else {
		return xerror.Wrap(err, ErrorInvalidEnv)
	}
}

type EnvURL struct {
	URL *url.URL
}

func (u *EnvURL) Decode(s string) error {
	if decoded, err := url.Parse(s); err == nil {
		u.URL = decoded
		return nil
	} else {
		return xerror.Wrap(err, ErrorInvalidEnv)
	}
}
