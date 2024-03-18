package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

type URLShortenerController struct {
	Shortener service.URLShortenerService
}

func (c URLShortenerController) SaveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	shortURL := c.Shortener.AddURL(urlStr)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", shortURL)
}

func (c URLShortenerController) GetURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	value, ok := c.Shortener.GetURL(shortURL)
	if ok {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
