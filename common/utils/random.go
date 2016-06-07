package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"io"
	"math/big"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GenRandomString generates a random alphabetical string of the given length.
func GenRandomString(length int) string {
	b := make([]byte, length)
	for i, cache, remain := length-1, GenRandomInt(63), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = GenRandomInt(63), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// GenRandomInt generates a random integer using at most the given number of bits.
func GenRandomInt(bits uint) int64 {
	r, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(bits)), nil))
	if err != nil {
		panic(err)
	}
	return r.Int64()
}

// GenRandomIntRange generates a random int in the given range.
func GenRandomIntRange(min, max int64) int64 {
	if min < 0 || max < 0 || min >= max {
		panic(xerror.New("invalid min and/or max values (%v, %v)", min, max))
	}
	r, err := rand.Int(rand.Reader, big.NewInt(max+1-min)) // Add one because Int returns [0, max).
	if err != nil {
		panic(err)
	}
	return r.Int64() + min
}

// SaltAndHash returns a SHA-256 hash of the salted subject, formatted as hex string.
func SaltAndHash(subject, salt string) string {
	h := sha256.New()
	io.WriteString(h, subject+salt)
	return fmt.Sprintf("%x", h.Sum(nil))
}
