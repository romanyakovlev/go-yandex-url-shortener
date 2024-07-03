package middlewares

import (
	"net/http"
	"time"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter // Встраивание оригинального http.ResponseWriter для расширения его функционала.
		responseData        *responseData
	}
)

// Write переопределяет метод Write оригинального http.ResponseWriter,
// позволяя захватить размер отправленных данных.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // Захватываем размер отправленных данных.
	return size, err
}

// WriteHeader переопределяет метод WriteHeader оригинального http.ResponseWriter,
// позволяя захватить статус ответа.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // Захватываем статус ответа.
}

// RequestLoggerMiddleware логирует информацию о каждом запросе и ответе.
func RequestLoggerMiddleware(s *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w, // Встраиваем оригинальный http.ResponseWriter.
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, r) // Внедряем нашу реализацию http.ResponseWriter.

			duration := time.Since(start)

			// Логирование информации о запросе и ответе.
			s.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status, // Получаем перехваченный код статуса ответа.
				"duration", duration,
				"size", responseData.size, // Получаем перехваченный размер ответа.
			)
		}
		return http.HandlerFunc(logFn)
	}
}
