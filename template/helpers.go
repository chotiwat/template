package template

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"
)

// Helpers is a namespace for helper functions.
type Helpers struct{}

// UTCNow returns the current time in utc.
func (h Helpers) UTCNow() time.Time {
	return time.Now().UTC()
}

// CreateKey creates an encryption key (base64 encoded).
func (h Helpers) CreateKey(keySize int) string {
	key := make([]byte, keySize)
	io.ReadFull(rand.Reader, key)
	return base64.StdEncoding.EncodeToString(key)
}

// UUID returns a uuidv4 as a string.
func (h Helpers) UUID() string {
	return UUIDv4().String()
}
