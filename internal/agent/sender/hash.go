package sender

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HashCount - возвращает хэш сумму
func HashCount(body []byte, key string) string {
	if key == "" {
		return ""
	}
	secretKey := []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(body))
	signature := h.Sum(nil)
	signatureHex := hex.EncodeToString(signature)
	return signatureHex
}
