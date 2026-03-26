package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func checkSignature(r *http.Request, key string) error {
	signature := r.Header.Get("HashSHA256")

	if signature == "" || key == "" {
		return nil
	}

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(body)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("wrong hash")
	}
	return nil
}

func CheckHash(logger *zap.Logger, key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := checkSignature(r, key); err != nil {
				logger.Warn("Invalid signature",
					zap.Error(err),
				)
				http.Error(w, "Invalid signature", http.StatusBadRequest)
			}
			next.ServeHTTP(w, r)
		})
	}
}
