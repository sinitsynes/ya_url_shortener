package handler

import (
	"log/slog"

	ya_middleware "ya_url_shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, register ...func(chi.Router)) *chi.Mux {
	r := chi.NewRouter()
	r.Use(ya_middleware.Logger(logger))
	r.Use(ya_middleware.GzipCompressor)
	r.Use(ya_middleware.GzipDecompressor)
	r.Use(middleware.Recoverer)

	for _, fn := range register {
		fn(r)
	}
	return r
}
