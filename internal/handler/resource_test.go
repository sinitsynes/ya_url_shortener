package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/model"

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
	stubController struct {
		createFn func(string) (model.Resource, error)
		getFn    func(string) (model.Resource, error)
	}
)

func (s stubController) CreateResource(url string) (model.Resource, error) {
	return s.createFn(url)
}
func (s stubController) GetResource(short string) (model.Resource, error) {
	return s.getFn(short)
}

func testAppConfig() *config.Config {
	return &config.Config{
		ServerAddress:   "localhost:8000",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "storage.txt",
	}
}

func setupController(t *testing.T) handler.Controller {
	t.Helper()

	ctrl := stubController{
		createFn: func(url string) (model.Resource, error) {
			return model.Resource{ID: 1, Address: url, Shortened: "6bdb5b0"}, nil
		},
		getFn: func(short string) (model.Resource, error) {
			return model.Resource{ID: 1, Address: "https://practicum.yandex.ru/", Shortened: short}, nil
		},
	}
	return ctrl
}

func setupHandler(t *testing.T, baseURL string, controller handler.Controller) handler.Handler {
	t.Helper()

	logger := slog.Default()
	return handler.NewResourceHandler(baseURL, controller, logger)
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
				contentType:     handler.ContentTypePlainText,
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
			assert.Equal(t, test.want.contentType, res.Header.Get(handler.ContentTypeHeader))
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

			created, createdErr := controller.CreateResource(test.want.response)
			require.NoError(t, createdErr)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("url", created.Shortened)
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
				inputHeader:  handler.ContentTypeJSON,
				response:     model.ResourceResult{Result: cfg.BaseURL + "/6bdb5b0"},
				contentType:  handler.ContentTypeJSON,
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
			request.Header.Set(handler.ContentTypeHeader, test.want.inputHeader)
			writer := httptest.NewRecorder()
			h.ShortenURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(handler.ContentTypeHeader))
			assert.NotEmpty(t, res.Header.Get(handler.ContentLengthHeader))
		})
	}
}
