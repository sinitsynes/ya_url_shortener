package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}
			h.ServeHTTP(&lw, r)
			duration := time.Since(startTime)

			slog.Info(
				"request",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", duration,
				"status", lw.responseData.status,
				"size", lw.responseData.size)
		},
	)
}
