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

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}
			h.ServeHTTP(&lw, r)

			logger.InfoContext(r.Context(),
				"request",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", time.Since(startTime),
				"status", lw.responseData.status,
				"size", lw.responseData.size)
		})
	}
}
