package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level string) (*zap.Logger, error) {
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)
	return config.Build()
}

// ErrorLoggerMiddleware логирует HTTP ошибки (4xx/5xx)
func ErrorLoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := ww.Status()
				// Пропускаем успешные запросы и 404 для статики (опционально)
				if status < 400 {
					return
				}

				// Определяем уровень логирования
				logLevel := zapcore.ErrorLevel
				if status >= 400 {
					logLevel = zapcore.WarnLevel
				}

				// Логируем ошибку
				logger.Check(logLevel, "http error").Write(
					zap.String("method", r.Method),
					zap.String("uri", r.RequestURI), // r.URL.Path
					zap.Int("status", status),
					zap.Duration("duration", time.Since(start)),
					zap.Int("size", ww.BytesWritten()),
					zap.String("referer", r.Referer()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// RequestLoggerMiddleware — middleware-логер для входящих HTTP-запросов
func RequestLoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("http request",
					zap.String("method", r.Method),
					zap.String("uri", r.RequestURI),
					zap.Int("status", ww.Status()),
					zap.Duration("duration", time.Since(start)),
					zap.Int("size", ww.BytesWritten()),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
