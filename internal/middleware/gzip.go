package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	w.Header().Set("Content-Encoding", "gzip")
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
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			h.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
		},
	)
}
