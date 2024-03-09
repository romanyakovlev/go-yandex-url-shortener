package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_saveURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		bodyIsEmpty  bool
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, bodyIsEmpty: true},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, bodyIsEmpty: true},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, bodyIsEmpty: false},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()

			saveURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if !tc.bodyIsEmpty {
				assert.NotEqual(t, w.Body.String(), "", "Тело ответа пустое")
			}
		})
	}
}

func Test_getURLByID(t *testing.T) {

	linkToSave := "https://practicum.yandex.ru/"

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(linkToSave))
	res := httptest.NewRecorder()
	saveURL(res, req)

	shortLink, _ := io.ReadAll(res.Body)

	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, string(shortLink), nil)
			w := httptest.NewRecorder()

			getURLByID(w, r)

			if tc.method == http.MethodGet {
				assert.Equal(t, w.Header().Get("Location"), linkToSave, "Location-заголовок не совпадает с ожидаемым")
			}
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}

}
