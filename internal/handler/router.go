package handler

import (
	"log/slog"
	"net/http"

	ya_middleware "ya_url_shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler interface {
	CreateURL(w http.ResponseWriter, r *http.Request)
	GetURL(w http.ResponseWriter, r *http.Request)
	ShortenURL(w http.ResponseWriter, r *http.Request)
	Healthy(w http.ResponseWriter, r *http.Request)
}

func NewRouter(handler Handler, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(ya_middleware.Logger(logger))
	r.Use(ya_middleware.GzipCompressor)
	r.Use(ya_middleware.GzipDecompressor)
	r.Use(middleware.Recoverer)

	r.Post("/", handler.CreateURL)
	r.Get("/{url}", handler.GetURL)
	r.Post("/api/shorten", handler.ShortenURL)
	r.Get("/ping", handler.Healthy)

	return r
}
