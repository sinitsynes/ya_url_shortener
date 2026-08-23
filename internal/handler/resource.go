package handler

import (
	"io"
	"net/http"
	"strconv"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

type Handler struct {
	serverAddress string
	storage       repository.MemoryStorage
}

func NewHandler() *Handler {
	return &Handler{
		serverAddress: "http://localhost:8080",
		storage:       repository.NewStorage(),
	}
}

func (h *Handler) CreateUrl(w http.ResponseWriter, r *http.Request) {
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
	shortened := service.CreateResource(h.storage, bodyString)
	resp := h.serverAddress + "/" + shortened
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp)) //nolint: errcheck
}

func (h *Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
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
	resource, err := service.GetResource(h.storage, int32(identifier))
	if err != nil {
		http.Error(w, "Ошибка получения ресурса", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", resource)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
