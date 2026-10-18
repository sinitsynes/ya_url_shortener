package resource

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"uuid"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

const (
	ContentTypeJSON      string = "application/json"
	ContentTypePlainText string = "text/plain; charset=utf-8"
	ContentTypeHeader    string = "Content-Type"
	ContentLengthHeader  string = "Content-Length"
)

type Controller interface {
	CreateResource(ctx context.Context, url string, userID *uuid.UUID) (model.Resource, error)
	GetResource(ctx context.Context, shortenedURL string) (model.Resource, error)
	CreateBatch(
		ctx context.Context,
		resources []model.ResourceBatchInput,
		userID *uuid.UUID,
	) ([]model.ResourceBatchOutput, error)
	GetResourcesByUser(ctx context.Context, userID *uuid.UUID) ([]model.ResourcesByUser, error)
}

type Handler struct {
	controller Controller
	logger     *slog.Logger
}

func NewResourceHandler(controller Controller, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		controller: controller,
		logger:     logger,
	}
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	httpStatus := http.StatusCreated
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
	userID, _ := ctx.Value(config.UserIDKey).(uuid.UUID)
	resource, err := h.controller.CreateResource(ctx, bodyString, &userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMaxRetriesExceeded):
			h.logger.ErrorContext(ctx, "create url error", "error", err)
			http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
			return
		case errors.Is(err, repository.ErrOriginalURLConflict):
			httpStatus = http.StatusConflict
		default:
			h.logger.ErrorContext(ctx, "create url error", "error", err)
			http.Error(w, "Ошибка создания ресурса", http.StatusInternalServerError)
			return
		}
	}
	resp := resource.ShortURL
	w.Header().Set(ContentTypeHeader, ContentTypePlainText)
	authCookie := http.Cookie{ //nolint:gosec
		Name:     string(config.UserIDKey),
		Value:    userID.String(),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authCookie)
	w.WriteHeader(httpStatus)
	w.Write([]byte(resp)) //nolint: errcheck,gosec
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("url")
	if identifier == "" {
		http.Error(w, "Ошибка чтения идентификатора", http.StatusBadRequest)
		return
	}
	resource, err := h.controller.GetResource(r.Context(), identifier)
	if err != nil {
		http.Error(w, "Ошибка получения ресурса", http.StatusNotFound)
		return
	}
	w.Header().Add("Location", resource.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusCreated
	contentType := r.Header.Get(ContentTypeHeader)
	if contentType != ContentTypeJSON {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "Превышен максимальный размер запроса", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	input := new(model.ResourceInput)
	if marshalErr := json.Unmarshal(bodyBytes, input); marshalErr != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	userID, _ := ctx.Value(config.UserIDKey).(uuid.UUID)
	resource, err := h.controller.CreateResource(ctx, input.URL, &userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMaxRetriesExceeded):
			h.logger.ErrorContext(r.Context(), "shorten url error", "error", err)
			http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
			return
		case errors.Is(err, repository.ErrOriginalURLConflict):
			httpStatus = http.StatusConflict
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
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	w.Header().Set(ContentLengthHeader, strconv.Itoa(len(resp)))
	authCookie := http.Cookie{ //nolint:gosec
		Name:     string(config.UserIDKey),
		Value:    userID.String(),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authCookie)
	w.WriteHeader(httpStatus)
	w.Write(resp) //nolint: errcheck,gosec
}

func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get(ContentTypeHeader)
	if contentType != ContentTypeJSON {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}
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
	ctx := r.Context()
	userID, _ := ctx.Value(config.UserIDKey).(uuid.UUID)
	created, err := h.controller.CreateBatch(ctx, input, &userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMaxRetriesExceeded):
			h.logger.ErrorContext(ctx, "create batch", "err", err)
			http.Error(w, "Превышено количество попыток создания ресурса", http.StatusInternalServerError)
			return
		case errors.Is(err, repository.ErrOriginalURLConflict):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		default:
			h.logger.ErrorContext(ctx, "create batch", "err", err)
			http.Error(w, "Ошибка создания ресурсов", http.StatusInternalServerError)
			return
		}
	}
	res, err := json.Marshal(created)
	if err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	authCookie := http.Cookie{ //nolint:gosec
		Name:     string(config.UserIDKey),
		Value:    userID.String(),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authCookie)
	w.WriteHeader(http.StatusCreated)
	w.Write(res) //nolint: errcheck,gosec
}

func (h *Handler) GetResourcesByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseStatus := http.StatusOK
	userID, _ := ctx.Value(config.UserIDKey).(uuid.UUID)
	resources, err := h.controller.GetResourcesByUser(ctx, &userID)
	if err != nil {
		http.Error(w, "Ошибка получения ресурсов", http.StatusInternalServerError)
		return
	}
	if len(resources) == 0 {
		responseStatus = http.StatusNoContent
	}
	res, err := json.Marshal(resources)
	if err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	authCookie := http.Cookie{ //nolint:gosec
		Name:     string(config.UserIDKey),
		Value:    userID.String(),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authCookie)
	w.WriteHeader(responseStatus)
	w.Write(res) //nolint: errcheck,gosec
}
