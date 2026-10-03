package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: len(cafeList["tula"])},
	}

	server := httptest.NewServer(http.HandlerFunc(mainHandle))
	defer server.Close()

	for _, v := range requests {
		resp, err := server.Client().Get(server.URL + "/cafe?city=tula&count=" + strconv.Itoa(v.count))

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		require.NoError(t, err)

		actual := strings.TrimSpace(string(body))

		var cafes []string

		if actual != "" {
			cafes = strings.Split(actual, ",")
		}

		assert.Equal(t, v.want, len(cafes))
	}

}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	server := httptest.NewServer(http.HandlerFunc(mainHandle))
	defer server.Close()

	for _, v := range requests {
		resp, err := server.Client().Get(server.URL + "/cafe?city=moscow&search=" + v.search)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		require.NoError(t, err)

		actual := strings.TrimSpace(string(body))

		var cafes []string

		if actual != "" {
			cafes = strings.Split(actual, ",")
		}

		assert.Equal(t, v.wantCount, len(cafes))

		for _, cafe := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search)))
		}
	}
}
