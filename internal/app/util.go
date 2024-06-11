package app

import (
	"crypto/sha1"
	"encoding/hex"
)

func encode(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	id := hex.EncodeToString(h.Sum(nil))
	return id
}
