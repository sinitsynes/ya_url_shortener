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

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler/resource"
	"ya_url_shortener/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	wantCreateURL struct {
		responseCode    int
		input           []byte
		baseURL         string
		responsePattern []byte
		contentType     string
	}
	wantGetURL struct {
		responseCode int
		input        int32
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
		input        []model.ResourceBatchInput
		inputHeader  string
		response     []model.ResourceBatchOutput
		contentType  string
	}
	stubController struct {
		createFn      func(context.Context, string) (model.Resource, error)
		getFn         func(context.Context, string) (model.Resource, error)
		createBatchFn func(context.Context, []model.ResourceBatchInput) ([]model.ResourceBatchOutput, error)
	}
)

func (c stubController) CreateResource(ctx context.Context, url string) (model.Resource, error) {
	return c.createFn(ctx, url)
}
func (c stubController) GetResource(ctx context.Context, short string) (model.Resource, error) {
	return c.getFn(ctx, short)
}

func (c stubController) CreateBatch(
	ctx context.Context,
	input []model.ResourceBatchInput,
) ([]model.ResourceBatchOutput, error) {
	return c.createBatchFn(ctx, input)
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

	ctrl := stubController{
		createFn: func(_ context.Context, url string) (model.Resource, error) {
			return model.Resource{ID: 1, OriginalURL: url, ShortURL: "6bdb5b0"}, nil
		},
		getFn: func(_ context.Context, short string) (model.Resource, error) {
			return model.Resource{ID: 1, OriginalURL: "https://practicum.yandex.ru/", ShortURL: short}, nil
		},
		createBatchFn: func(_ context.Context, input []model.ResourceBatchInput) ([]model.ResourceBatchOutput, error) {
			out := make([]model.ResourceBatchOutput, 0, len(input))
			for _, item := range input {
				out = append(out, model.ResourceBatchOutput{
					CorrelationID: item.CorrelationID,
					ShortURL:      "6bdb5b0",
				})
			}
			return out, nil
		},
	}
	return ctrl
}

func setupHandler(t *testing.T, baseURL string, controller resource.Controller) resource.ResourceHandler {
	t.Helper()

	logger := slog.Default()
	return resource.NewResourceHandler(baseURL, controller, logger)
}

func TestCreateURL(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	tests := []struct {
		name string
		want wantCreateURL
	}{
		{
			name: "happy CreateURL #1",
			want: wantCreateURL{
				responseCode:    http.StatusCreated,
				input:           []byte("https://practicum.yandex.ru/"),
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
			h := setupHandler(t, cfg.BaseURL, controller)
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
	cfg := testAppConfig()
	tests := []struct {
		name string
		want wantGetURL
	}{
		{
			name: "happy GetURL #1",
			want: wantGetURL{
				responseCode: http.StatusTemporaryRedirect,
				input:        1,
				response:     "https://practicum.yandex.ru/",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, cfg.BaseURL, controller)

			created, createdErr := controller.CreateResource(t.Context(), test.want.response)
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
			h := setupHandler(t, cfg.BaseURL, controller)
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
	tests := []struct {
		name string
		want wantCreateBatch
	}{
		{
			name: "happy CreateBatch #1",
			want: wantCreateBatch{
				responseCode: http.StatusCreated,
				input: []model.ResourceBatchInput{
					{CorrelationID: uuid.MustParse("0cc3dd37-e05f-49ab-a716-e783556a7980"), OriginalURL: "https://ya.ru"},
					{CorrelationID: uuid.MustParse("241af749-05ea-42d9-8ada-3786c6d6018b"), OriginalURL: "https://yandex.ru"},
				},
				inputHeader: resource.ContentTypeJSON,
				response: []model.ResourceBatchOutput{
					{CorrelationID: uuid.MustParse("0cc3dd37-e05f-49ab-a716-e783556a7980"), ShortURL: "6bdb5b0"},
					{CorrelationID: uuid.MustParse("241af749-05ea-42d9-8ada-3786c6d6018b"), ShortURL: "6bdb5b0"},
				},
				contentType: resource.ContentTypeJSON,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := setupController(t)
			h := setupHandler(t, cfg.BaseURL, controller)
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
