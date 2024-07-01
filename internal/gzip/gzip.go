// Package gzip содержит логику работы с gzip-сжатием.
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

// compressWriter обертывает http.ResponseWriter, добавляя gzip-сжатие к ответу.
type compressWriter struct {
	w  http.ResponseWriter // Исходный ResponseWriter
	zw *gzip.Writer        // Writer для gzip-сжатия
}

// NewCompressWriter создает новый экземпляр compressWriter.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки HTTP-ответа.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в gzip.Writer, сжимая их перед отправкой клиенту.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader отправляет HTTP-статус код. Если код статуса позволяет сжатие,
// добавляет заголовок "Content-Encoding: gzip".
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 || statusCode == 409 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и освобождает ресурсы.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader обертывает io.ReadCloser, добавляя возможность чтения gzip-сжатых данных.
type compressReader struct {
	r  io.ReadCloser // Исходный ReadCloser
	zr *gzip.Reader  // Reader для распаковки gzip-сжатия
}

// NewCompressReader создает новый экземпляр compressReader для чтения gzip-сжатых данных.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и распаковывает gzip-сжатые данные.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader и исходный ReadCloser, освобождая ресурсы.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
