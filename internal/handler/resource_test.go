package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantCreateUrl struct {
	code        int
	input       []byte
	response    string
	contentType string
}

type wantGetUrl struct {
	code     int
	input    int32
	response string
}

func TestCreateUrl(t *testing.T) {
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{Host: "localhost", Port: "8080"},
		BaseUrl:    "http://localhost:8080",
	}
	controller := service.NewResourceController()
	handler := NewResourceHandler(cfg, controller)
	tests := []struct {
		name string
		want wantCreateUrl
	}{
		{
			name: "positive test #1",
			want: wantCreateUrl{
				code:        201,
				input:       []byte("https://practicum.yandex.ru/"),
				response:    cfg.BaseUrl + "/1",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(test.want.input))
			writer := httptest.NewRecorder()
			handler.CreateUrl(writer, request)
			res := writer.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer request.Body.Close() //nolint:errcheck
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetUrl(t *testing.T) {
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{Host: "localhost", Port: "8080"},
		BaseUrl:    "http://localhost:8080",
	}
	controller := service.NewResourceController()
	handler := NewResourceHandler(cfg, controller)

	tests := []struct {
		name string
		want wantGetUrl
	}{
		{
			name: "positive test #2",
			want: wantGetUrl{
				code:     307,
				input:    1,
				response: "https://practicum.yandex.ru/",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newResource := controller.CreateResource(test.want.response)
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("id", strconv.Itoa(int(newResource.Identifier)))
			writer := httptest.NewRecorder()
			handler.GetUrl(writer, request)
			res := writer.Result()
			defer res.Body.Close() //nolint:errcheck

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.response, res.Header.Get("Location"))
		})
	}
}
