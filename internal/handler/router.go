package handler

import (
	"log/slog"
	"time"

	"ya_url_shortener/internal/config"
	ya_middleware "ya_url_shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	maxRequestSize int64         = 20 * (1 << 10) //nolint: mnd // 20KB
	requestTimeout time.Duration = 3 * time.Second
)

func NewRouter(logger *slog.Logger, secretKey config.AuthSecretKey, register ...func(chi.Router)) *chi.Mux {
	r := chi.NewRouter()
	r.Use(ya_middleware.Logger(logger))
	r.Use(middleware.Recoverer)
	r.Use(ya_middleware.Timeout(requestTimeout))
	r.Use(ya_middleware.GzipDecompressor)
	r.Use(ya_middleware.GzipCompressor)
	r.Use(ya_middleware.LimitMaxRequestSize(maxRequestSize))
	r.Use(ya_middleware.CookieAuth(config.UserIDKey, secretKey, logger))

	for _, fn := range register {
		fn(r)
	}
	return r
}
