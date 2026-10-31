package healthcheck

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthcheckHandler interface { //nolint: revive
	CheckDatabase(w http.ResponseWriter, r *http.Request)
}

func RegisterRoutes(r chi.Router, handler HealthcheckHandler) {
	r.Get("/ping", handler.CheckDatabase)
}
