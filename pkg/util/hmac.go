package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HMac(payload, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}