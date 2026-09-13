package handler_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantCreateURL struct {
	code        int
	input       []byte
	baseURL     string
	contentType string
}

type wantGetURL struct {
	code     int
	input    int32
	response string
}

func testAppConfig() *config.Config {
	return &config.Config{
		ServerAddress: "localhost:8000",
		BaseURL:       "http://localhost:8080",
	}
}

func TestCreateUrl(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	tests := []struct {
		name string
		want wantCreateURL
	}{
		{
			name: "positive test #1",
			want: wantCreateURL{
				code:        201,
				input:       []byte("https://practicum.yandex.ru/"),
				baseURL:     cfg.BaseURL,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := repository.NewStore()
			controller := service.NewResourceController(repo)
			h := handler.NewResourceHandler(cfg.BaseURL, controller)
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(test.want.input))
			writer := httptest.NewRecorder()
			h.CreateURL(writer, request)
			res := writer.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer request.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Regexp(t,
				`^`+regexp.QuoteMeta(test.want.baseURL)+`/[0-9A-Za-z]+$`,
				string(resBody),
			)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetUrl(t *testing.T) {
	t.Parallel()
	cfg := testAppConfig()
	tests := []struct {
		name string
		want wantGetURL
	}{
		{
			name: "positive test #2",
			want: wantGetURL{
				code:     307,
				input:    1,
				response: "https://practicum.yandex.ru/",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := repository.NewStore()
			controller := service.NewResourceController(repo)
			h := handler.NewResourceHandler(cfg.BaseURL, controller)

			created, createdErr := controller.CreateResource(test.want.response)
			require.NoError(t, createdErr)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("url", created.Shortened)
			writer := httptest.NewRecorder()
			h.GetURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.response, res.Header.Get("Location"))
		})
	}
}
