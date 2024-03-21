package Middleware

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read request body", zap.Error(err))
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

			clientIP := r.RemoteAddr

			logger.Debug("request details",
				zap.ByteString("body", reqBody),
				zap.Any("headers", r.Header),
				zap.String("clientIP", clientIP),
			)

			next.ServeHTTP(w, r)
		})
	}
}

func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("recovered from panic", zap.Any("error", err))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func TraceMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateTraceID() // Функция генерации traceID
				logger.Debug("Generated new traceID", zap.String("traceID", traceID))

				w.Header().Set("X-Trace-ID", traceID)
			} else {
				logger.Debug("Received request with traceID", zap.String("traceID", traceID))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func generateTraceID() string {

	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}

	return fmt.Sprintf("%x-%s", time.Now().UnixNano(), string(b))
}
