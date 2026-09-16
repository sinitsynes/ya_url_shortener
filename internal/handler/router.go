package handler

import (
	"net/http"

	ya_middleware "ya_url_shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler interface {
	CreateURL(w http.ResponseWriter, r *http.Request)
	GetURL(w http.ResponseWriter, r *http.Request)
}

func NewRouter(handler Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(ya_middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", handler.CreateURL)
	r.Get("/{url}", handler.GetURL)

	return r
}
