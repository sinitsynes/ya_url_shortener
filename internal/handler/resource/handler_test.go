package resource_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler/resource"
	"ya_url_shortener/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	wantCreateURL struct {
		responseCode    int
		input           []byte
		userID          *uuid.UUID
		baseURL         string
		responsePattern []byte
		contentType     string
	}
	wantGetURL struct {
		responseCode int
		input        int32
		userID       *uuid.UUID
		response     string
	}
	wantShortenURL struct {
		responseCode int
		input        model.ResourceInput
		inputHeader  string
		response     model.ResourceResult
		contentType  string
	}
	wantCreateBatch struct {
		responseCode int
		userID       *uuid.UUID
		input        []model.ResourceBatchInput
		inputHeader  string
		response     []model.ResourceBatchOutput
		contentType  string
	}
	wantURLbyUser struct {
		responseCode int
		userID       *uuid.UUID
		input        []model.Resource
		response     []model.ResourcesByUser
		contentType  string
	}
	stubController struct {
		createFn             func(context.Context, string, *uuid.UUID) (model.Resource, error)
		getFn                func(context.Context, string) (model.Resource, error)
		createBatchFn        func(context.Context, []model.ResourceBatchInput, *uuid.UUID) ([]model.ResourceBatchOutput, error)
		getResourcesByUserFn func(context.Context, *uuid.UUID) ([]model.ResourcesByUser, error)
	}
)

func (c stubController) CreateResource(ctx context.Context, url string, userID *uuid.UUID) (model.Resource, error) {
	return c.createFn(ctx, url, userID)
}
func (c stubController) GetResource(ctx context.Context, short string) (model.Resource, error) {
	return c.getFn(ctx, short)
}

func (c stubController) CreateBatch(
	ctx context.Context,
	input []model.ResourceBatchInput,
	userID *uuid.UUID,
) ([]model.ResourceBatchOutput, error) {
	return c.createBatchFn(ctx, input, userID)
}

func (c stubController) GetResourcesByUser(
	ctx context.Context,
	userID *uuid.UUID,
) ([]model.ResourcesByUser, error) {
	return c.getResourcesByUserFn(ctx, userID)
}

func testAppConfig() *config.Config {
	return &config.Config{
		ServerAddress:   "localhost:8000",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "storage.txt",
	}
}

func setupController(t *testing.T) resource.Controller {
	t.Helper()
	cfg := testAppConfig()

	ctrl := stubController{
		createFn: func(_ context.Context, url string, userID *uuid.UUID) (model.Resource, error) {
			return model.Resource{ID: 1, OriginalURL: url, ShortURL: cfg.BaseURL + "/6bdb5b0", UserID: userID}, nil
		},
		getFn: func(_ context.Context, short string) (model.Resource, error) {
			return model.Resource{ID: 1, OriginalURL: "https://practicum.yandex.ru/", ShortURL: short}, nil
		},
		createBatchFn: func(_ context.Context, input []model.ResourceBatchInput, _ *uuid.UUID) ([]model.ResourceBatchOutput, error) {
			out := make([]model.ResourceBatchOutput, 0, len(input))
			for _, item := range input {
				out = append(out, model.ResourceBatchOutput{
					CorrelationID: item.CorrelationID,
					ShortURL:      cfg.BaseURL + "/6bdb5b0",
				})
			}
			return out, nil
		},
		getResourcesByUserFn: func(_ context.Context, _ *uuid.UUID) ([]model.ResourcesByUser, error) {
			return []model.ResourcesByUser{
				{
					OriginalURL: "https://practicum.yandex.ru/",
					ShortURL:    cfg.BaseURL + "/6bdb5b0"},
			}, nil
		},
	}
	return ctrl
}

func setupHandler(t *testing.T, controller resource.Controller) resource.ResourceHandler {
	t.Helper()

	logger := slog.Default()
	return resource.NewResourceHandler(controller, logger)
}

func TestCreateURL(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	goodUser := uuid.New()

	tests := []struct {
		name string
		want wantCreateURL
	}{
		{
			name: "happy CreateURL #1",
			want: wantCreateURL{
				responseCode:    http.StatusCreated,
				input:           []byte("https://practicum.yandex.ru/"),
				userID:          &goodUser,
				baseURL:         cfg.BaseURL,
				responsePattern: []byte(cfg.BaseURL + "/6bdb5b0"),
				contentType:     resource.ContentTypePlainText,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, controller)
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(test.want.input))
			writer := httptest.NewRecorder()
			h.CreateURL(writer, request)
			res := writer.Result()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			defer request.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Regexp(t,
				test.want.responsePattern,
				string(resBody),
			)
			assert.Equal(t, test.want.contentType, res.Header.Get(resource.ContentTypeHeader))
		})
	}
}

func TestGetURL(t *testing.T) {
	t.Parallel()
	goodUser := uuid.New()
	tests := []struct {
		name string
		want wantGetURL
	}{
		{
			name: "happy GetURL #1",
			want: wantGetURL{
				responseCode: http.StatusTemporaryRedirect,
				input:        1,
				userID:       &goodUser,
				response:     "https://practicum.yandex.ru/",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, controller)

			created, createdErr := controller.CreateResource(t.Context(), test.want.response, &goodUser)
			require.NoError(t, createdErr)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("url", created.ShortURL)
			writer := httptest.NewRecorder()
			h.GetURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.response, res.Header.Get("Location"))
		})
	}
}

func TestShortenURL(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	tests := []struct {
		name string
		want wantShortenURL
	}{
		{
			name: "happy ShortenURL #1",
			want: wantShortenURL{
				responseCode: http.StatusCreated,
				input:        model.ResourceInput{URL: "https://practicum.yandex.ru/"},
				inputHeader:  resource.ContentTypeJSON,
				response:     model.ResourceResult{Result: cfg.BaseURL + "/6bdb5b0"},
				contentType:  resource.ContentTypeJSON,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, controller)
			input, err := json.Marshal(test.want.input)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(input))
			request.Header.Set(resource.ContentTypeHeader, test.want.inputHeader)
			writer := httptest.NewRecorder()
			h.ShortenURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(resource.ContentTypeHeader))
			assert.NotEmpty(t, res.Header.Get(resource.ContentLengthHeader))
		})
	}
}

func TestCreateBatch(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	goodUser := uuid.New()
	tests := []struct {
		name string
		want wantCreateBatch
	}{
		{
			name: "happy CreateBatch #1",
			want: wantCreateBatch{
				responseCode: http.StatusCreated,
				input: []model.ResourceBatchInput{
					{
						CorrelationID: "example1",
						OriginalURL:   "https://ya.ru",
					},
					{
						CorrelationID: "example2",
						OriginalURL:   "https://yandex.ru",
					},
				},
				userID:      &goodUser,
				inputHeader: resource.ContentTypeJSON,
				response: []model.ResourceBatchOutput{
					{
						CorrelationID: "example1",
						ShortURL:      cfg.BaseURL + "/6bdb5b0",
					},
					{
						CorrelationID: "example2",
						ShortURL:      cfg.BaseURL + "/6bdb5b0",
					},
				},
				contentType: resource.ContentTypeJSON,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, controller)
			input, err := json.Marshal(test.want.input)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(input))
			request.Header.Set(resource.ContentTypeHeader, test.want.inputHeader)
			writer := httptest.NewRecorder()
			h.CreateBatch(writer, request)
			res := writer.Result()
			items := []model.ResourceBatchOutput{}
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			got := json.Unmarshal(body, &items)
			require.NoError(t, got)
			defer res.Body.Close()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(resource.ContentTypeHeader))
			assert.ElementsMatch(t, test.want.response, items)
		})
	}
}

func TestGetResourcesByUser(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	goodUser := uuid.New()
	tests := []struct {
		name string
		want wantURLbyUser
	}{
		{
			name: "happy GetResourcesByUser #1",
			want: wantURLbyUser{
				responseCode: http.StatusOK,
				userID:       &goodUser,
				input: []model.Resource{
					{
						ID:          1,
						OriginalURL: "https://practicum.yandex.ru/",
						ShortURL:    cfg.BaseURL + "/6bdb5b0",
						UserID:      &goodUser,
					},
				},
				response: []model.ResourcesByUser{
					{
						OriginalURL: "https://practicum.yandex.ru/",
						ShortURL:    cfg.BaseURL + "/6bdb5b0",
					},
				},
				contentType: resource.ContentTypeJSON,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, controller)
			request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			request = request.WithContext(context.WithValue(request.Context(), config.UserIDKey, *test.want.userID))
			writer := httptest.NewRecorder()
			h.GetResourcesByUser(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			items := []model.ResourcesByUser{}
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(body, &items))

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(resource.ContentTypeHeader))
			assert.ElementsMatch(t, test.want.response, items)
		})
	}
}
