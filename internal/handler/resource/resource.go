package resource

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

const (
	ContentTypeJSON      string        = "application/json"
	ContentTypePlainText string        = "text/plain; charset=utf-8"
	ContentTypeHeader    string        = "Content-Type"
	ContentLengthHeader  string        = "Content-Length"
	MaxRequestSize       int64         = 20 * (1 << 10) //nolint: mnd // 20KB
	RequestTimeout       time.Duration = 3 * time.Second
)

type Controller interface {
	CreateResource(ctx context.Context, url string) (model.Resource, error)
	GetResource(ctx context.Context, shortenedURL string) (model.Resource, error)
	CreateBatch(ctx context.Context, resources []model.ResourceBatchInput) ([]model.ResourceBatchOutput, error)
}

type Handler struct {
	controller     Controller
	logger         *slog.Logger
	requestTimeout time.Duration
}

func NewResourceHandler(controller Controller, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		controller:     controller,
		logger:         logger,
		requestTimeout: RequestTimeout,
	}
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "Превышен максимальный размер запроса", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	bodyString := string(bodyBytes)
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()
	resource, err := h.controller.CreateResource(ctx, bodyString)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMaxRetriesExceeded):
			h.logger.ErrorContext(ctx, "create url error", "error", err)
			http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
			return
		case errors.Is(err, repository.ErrConflict):
			http.Error(w, "Запись с URL уже существует", http.StatusConflict)
			return
		default:
			h.logger.ErrorContext(ctx, "create url error", "error", err)
			http.Error(w, "Ошибка создания ресурса", http.StatusInternalServerError)
			return
		}
	}
	resp := resource.ShortURL
	w.Header().Set(ContentTypeHeader, ContentTypePlainText)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp)) //nolint: errcheck,gosec
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("url")
	if identifier == "" {
		http.Error(w, "Ошибка чтения идентификатора", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()
	resource, err := h.controller.GetResource(ctx, identifier)
	if err != nil {
		http.Error(w, "Ошибка получения ресурса", http.StatusNotFound)
		return
	}
	w.Header().Add("Location", resource.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get(ContentTypeHeader)
	if contentType != ContentTypeJSON {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)

	input := model.ResourceInput{}
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "Превышен максимальный размер запроса", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyBytes, &input)
	if err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()
	resource, err := h.controller.CreateResource(ctx, input.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMaxRetriesExceeded):
			h.logger.ErrorContext(r.Context(), "shorten url error", "error", err)
			http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
			return
		case errors.Is(err, repository.ErrConflict):
			http.Error(w, "Запись с этим URL уже существует", http.StatusConflict)
			return
		default:
			h.logger.ErrorContext(ctx, "shorten url error", "error", err)
			http.Error(w, "Ошибка создания ресурса", http.StatusInternalServerError)
			return
		}
	}
	result := model.ResourceResult{Result: resource.ShortURL}
	resp, err := json.Marshal(result)
	if err != nil {
		h.logger.ErrorContext(ctx, "shorten url error", "error", err)
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
		return
	}
	w.Header().Set(ContentLengthHeader, strconv.Itoa(len(resp)))
	w.WriteHeader(http.StatusCreated)
	w.Write(resp) //nolint: errcheck,gosec
}

func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get(ContentTypeHeader)
	if contentType != ContentTypeJSON {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "Превышен максимальный размер запроса", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	input := []model.ResourceBatchInput{}
	err = json.Unmarshal(bodyBytes, &input)
	if err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()
	created, err := h.controller.CreateBatch(ctx, input)
	if err != nil {
		http.Error(w, "Ошибка создания ресурсов", http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(created)
	if err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(res) //nolint: errcheck,gosec
}
