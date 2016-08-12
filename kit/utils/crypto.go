package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
)

// SaltAndHash returns a SHA-256 hash of the salted subject, formatted as hex string.
func SaltAndHash(subject, salt string) string {
	h := sha256.New()
	io.WriteString(h, subject+salt)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD5 returns the MD5 hash of the given string.
func MD5(subject string) string {
	h := md5.New()
	io.WriteString(h, subject)
	return fmt.Sprintf("%x", h.Sum(nil))
}
