package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaugeHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}

	testCases := []struct {
		name string
		request string
		want want
	}{
		{
			name: "positive test, ok",
			request: "http://localhost:8080/update/gauge/somemetric/547",
			want: want {
				code: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "negative test, not found",
			request: "http://localhost:8080/update/gauge/547",
			want: want {
				code: http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.request, nil)

			w := httptest.NewRecorder()
			GaugeHandler(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestCounterHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}

	testCases := []struct {
		name string
		request string
		want want
	}{
		{
			name: "positive test, ok",
			request: "http://localhost:8080/update/counter/somemetric/547",
			want: want {
				code: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "negative test, not found",
			request: "http://localhost:8080/update/counter/547",
			want: want {
				code: http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.request, nil)

			w := httptest.NewRecorder()
			CounterHandler(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestDefaultHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}

	testCases := []struct {
		name string
		request string
		want want
	}{
		{
			name: "positive test, ok",
			request: "http://localhost:8080/update/counter/somemetric/547",
			want: want {
				code: http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.request, nil)

			w := httptest.NewRecorder()
			DefaultHandler(w, r)

			res := w.Result()
			defer res.Body.Close()
			
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
