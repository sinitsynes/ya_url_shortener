package healthcheck

import (
	"log/slog"
	"net/http"

	"ya_url_shortener/internal/infra/db"
)

type Handler struct {
	pinger db.Pinger
	logger *slog.Logger
}

func NewHandler(pinger db.Pinger, logger *slog.Logger) *Handler {
	return &Handler{pinger: pinger, logger: logger}
}

func (h *Handler) CheckDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.pinger.Ping(ctx); err != nil {
		h.logger.ErrorContext(r.Context(), "database is not available", "error", err)
		http.Error(w, "service not available", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
