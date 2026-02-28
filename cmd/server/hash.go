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

func CheckHash(logger *zap.Logger, key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			signature := r.Header.Get("HashSHA256")

			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			h := hmac.New(sha256.New, []byte(secretKey))
			h.Write(body)
			expectedSignature := hex.EncodeToString(h.Sum(nil))

			if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
				logger.Warn("Invalid signature",
					zap.Error(fmt.Errorf("wrong hash")),
					zap.String(signature, expectedSignature),
				)
				http.Error(w, "Invalid signature", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
