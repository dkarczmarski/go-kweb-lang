package internal

import (
	"crypto/sha1"
	"encoding/hex"
)

func KeyFile(key string) string {
	return key + ".json"
}

func KeyHash(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))

	return hex.EncodeToString(hash.Sum(nil))
}
