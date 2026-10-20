package resource

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ResourceHandler interface { //nolint: revive
	CreateURL(w http.ResponseWriter, r *http.Request)
	GetURL(w http.ResponseWriter, r *http.Request)
	ShortenURL(w http.ResponseWriter, r *http.Request)
	CreateBatch(w http.ResponseWriter, r *http.Request)
	GetResourcesByUser(w http.ResponseWriter, r *http.Request)
	DeleteResources(w http.ResponseWriter, r *http.Request)
}

func RegisterRoutes(r chi.Router, handler ResourceHandler) {
	r.Post("/", handler.CreateURL)
	r.Get("/{url}", handler.GetURL)
	r.Post("/api/shorten", handler.ShortenURL)
	r.Post("/api/shorten/batch", handler.CreateBatch)
	r.Get("/api/user/urls", handler.GetResourcesByUser)
	r.Delete("/api/user/urls", handler.DeleteResources)
}
