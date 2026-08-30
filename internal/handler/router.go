package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler interface {
	CreateUrl(w http.ResponseWriter, r *http.Request)
	GetUrl(w http.ResponseWriter, r *http.Request)
}

func NewRouter(handler Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", handler.CreateUrl)
	r.Get("/{url}", handler.GetUrl)

	return r
}
