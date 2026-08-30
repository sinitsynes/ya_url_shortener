package handler

import (
	"io"
	"net/http"
	"strconv"
	"ya_url_shortener/internal/model"
)

type Controller interface {
	CreateResource(url string) (model.Resource, error)
	GetResource(id int32) (model.Resource, error)
}

type ResourceHandler struct {
	baseUrl    string
	controller Controller
}

func NewResourceHandler(baseUrl string, controller Controller) *ResourceHandler {
	return &ResourceHandler{
		baseUrl:    baseUrl,
		controller: controller,
	}
}

func (h *ResourceHandler) CreateUrl(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	bodyString := string(bodyBytes)
	resource, err := h.controller.CreateResource(bodyString)
	if err != nil {
		http.Error(w, "Ошибка создания ресурса", http.StatusInternalServerError)
		return
	}
	resp := h.baseUrl + "/" + resource.Shortened
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp)) //nolint: errcheck
}

func (h *ResourceHandler) GetUrl(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Ошибка чтения идентификатора", http.StatusBadRequest)
		return
	}
	identifier, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Ошибка преобразования идентификатора", http.StatusBadRequest)
		return
	}
	resource, err := h.controller.GetResource(int32(identifier))
	if err != nil {
		http.Error(w, "Ошибка получения ресурса", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", resource.Address)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
