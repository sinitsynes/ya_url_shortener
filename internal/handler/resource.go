package handler

import (
	"errors"
	"io"
	"net/http"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/service"
)

type Controller interface {
	CreateResource(url string) (model.Resource, error)
	GetResource(shortenedURL string) (model.Resource, error)
}

type ResourceHandler struct {
	baseURL    string
	controller Controller
}

func NewResourceHandler(baseURL string, controller Controller) *ResourceHandler {
	return &ResourceHandler{
		baseURL:    baseURL,
		controller: controller,
	}
}

func (h *ResourceHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	bodyString := string(bodyBytes)
	resource, err := h.controller.CreateResource(bodyString)
	if errors.Is(err, service.ErrMaxRetriesExceeded) {
		http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка создания ресурса", http.StatusInternalServerError)
		return
	}
	resp := h.baseURL + "/" + resource.Shortened
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp)) //nolint: errcheck,gosec
}

func (h *ResourceHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("url")
	if identifier == "" {
		http.Error(w, "Ошибка чтения идентификатора", http.StatusBadRequest)
		return
	}
	resource, err := h.controller.GetResource(identifier)
	if err != nil {
		http.Error(w, "Ошибка получения ресурса", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", resource.Address)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
