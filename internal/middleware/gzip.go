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
	Writer *gzip.Writer
	status int
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.Writer.Write(b)
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

func GzipCompressor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				h.ServeHTTP(w, r)
				return
			}

			buffer := &bytes.Buffer{}
			gz := gzip.NewWriter(buffer)
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
				status:         http.StatusOK,
			}

			h.ServeHTTP(gzw, r)
			_ = gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
			w.WriteHeader(gzw.status)
			_, _ = w.Write(buffer.Bytes())
		},
	)
}
