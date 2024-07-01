package middlewares

import (
	"net/http"
	"strings"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/gzip"
)

// GzipMiddleware является промежуточным обработчиком (middleware), который добавляет поддержку gzip-сжатия.
// Он проверяет заголовки запроса на наличие поддержки gzip и, если поддерживается, сжимает ответ.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		ContentType := r.Header.Get("Content-Type")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		supportsContentType := strings.Contains(ContentType, "application/json") || strings.Contains(ContentType, "text/html")
		if supportsGzip && supportsContentType {
			cw := gzip.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	})
}
