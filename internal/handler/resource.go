package handler

import (
	"io"
	"net/http"
	"strconv"
	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/service"
)

type Controller interface {
	CreateResource(url string) model.Resource
	GetResource(id int32) (model.Resource, error)
}

type ResourceHandler struct {
	serverAddress string
	controller    Controller
}

func NewResourceHandler(controller Controller) *ResourceHandler {
	if controller == nil {
		controller = service.NewResourceController()
	}
	return &ResourceHandler{
		serverAddress: "http://localhost:8080",
		controller:    controller,
	}
}

func (h *ResourceHandler) CreateUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод недоступен", http.StatusBadRequest)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	bodyString := string(bodyBytes)
	resource := h.controller.CreateResource(bodyString)
	resp := h.serverAddress + "/" + resource.Shortened
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp)) //nolint: errcheck
}

func (h *ResourceHandler) GetUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод недоступен", http.StatusBadRequest)
		return
	}
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
