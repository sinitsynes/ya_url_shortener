package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"

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
)

func testAppConfig() *config.Config {
	return &config.Config{
		ServerAddress: "localhost:8000",
		BaseURL:       "http://localhost:8080",
	}
}

func setupController(t *testing.T) handler.Controller {
	t.Helper()

	repo := repository.NewStore()
	return service.NewResourceController(repo)

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
			name: "positive test #1",
			want: wantCreateURL{
				responseCode:    http.StatusCreated,
				input:           []byte("https://practicum.yandex.ru/"),
				baseURL:         cfg.BaseURL,
				responsePattern: []byte(`^` + regexp.QuoteMeta(cfg.BaseURL+`/[0-9A-Za-z]+$`)),
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
			name: "positive test #2",
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
			name: "positive test #1",
			want: wantShortenURL{
				responseCode: http.StatusCreated,
				input:        model.ResourceInput{URL: "https://practicum.yandex.ru/"},
				inputHeader:  handler.ContentTypeJSON,
				response:     model.ResourceResult{Result: regexp.QuoteMeta(cfg.BaseURL + `/[0-9A-Za-z]+$`)},
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
			assert.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(input))
			request.Header.Set(handler.ContentTypeHeader, test.want.inputHeader)
			writer := httptest.NewRecorder()
			h.ShortenURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(handler.ContentTypeHeader))
			assert.NotEqual(t, "", res.Header.Get(handler.ContentLengthHeader))
		})
	}
}
