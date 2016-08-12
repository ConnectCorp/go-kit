package utils

import (
	"testing"
	"regexp"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestSaltAndHash(t *testing.T) {
	r := regexp.MustCompile("^[0-9a-f]+$")
	for i := 0; i < 1000; i++ {
		v := SaltAndHash("subject", fmt.Sprintf("%v", i))
		assert.Equal(t, 64, len(v))
		assert.True(t, r.MatchString(v))
	}
}

func TestMD5(t *testing.T) {
	assert.Equal(t, "098f6bcd4621d373cade4e832627b4f6", MD5("test"))
}