package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urlMap = map[string]string{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func saveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	randomPath := RandStringBytes(8)
	urlMap[randomPath] = urlStr
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v/%v", flagBAddr, randomPath)
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	value, ok := urlMap[shortURL]
	if ok {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func URLShortenerRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", saveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
	return r
}

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	err := http.ListenAndServe(flagAAddr, URLShortenerRouter())
	if err != nil {
		panic(err)
	}
	return nil
}
