package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestGenRandomString(t *testing.T) {
	r := regexp.MustCompile("^[a-zA-Z]+$")
	for i := 0; i < 10000; i++ {
		v := GenRandomString(10)
		assert.Equal(t, 10, len(v))
		assert.True(t, r.MatchString(v))
	}
}

func TestGenRandomInt(t *testing.T) {
	for i := 0; i < 10000; i++ {
		v := GenRandomInt(32)
		assert.EqualValues(t, v, int32(v))
	}
}

func TestGenRandomIntRange(t *testing.T) {
	for i := 0; i < 10000; i++ {
		v := GenRandomIntRange(100, 1000)
		assert.True(t, v >= 100 && v <= 1000)
	}
}

func TestSaltAndHash(t *testing.T) {
	r := regexp.MustCompile("^[0-9a-f]+$")
	for i := 0; i < 1000; i++ {
		v := SaltAndHash("subject", fmt.Sprintf("%v", i))
		assert.Equal(t, 64, len(v))
		assert.True(t, r.MatchString(v))
	}
}
