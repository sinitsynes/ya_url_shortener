package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter

	buffer *bytes.Buffer
	status int
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func GzipDecompressor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzReader, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "invalid gzip body", http.StatusBadRequest)
					return
				}
				defer gzReader.Close()
				r.Body = gzReader
			}
			h.ServeHTTP(w, r)
		})
}

func shouldCompress(header http.Header) bool {
	c := header.Get("Content-Type")
	contentType := strings.TrimSpace(strings.Split(c, ";")[0])
	switch contentType {
	case "application/json":
		return true
	case "text/html":
		return true
	default:
		return false
	}
}

func GzipCompressor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				h.ServeHTTP(w, r)
				return
			}

			buffer := &bytes.Buffer{}
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				buffer:         buffer}
			h.ServeHTTP(gzw, r)

			body := gzw.buffer.Bytes()
			if shouldCompress(w.Header()) {
				var compressed bytes.Buffer
				gz := gzip.NewWriter(&compressed)
				_, _ = gz.Write(body)
				_ = gz.Close()
				w.Header().Set("Content-Encoding", "gzip")
				body = compressed.Bytes()
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			if gzw.status == 0 {
				gzw.status = http.StatusOK
			}
			w.WriteHeader(gzw.status)
			_, _ = w.Write(body)
		},
	)
}
