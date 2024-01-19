package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type want struct {
	code        int
	contentType string
	body        string
}

type MockStorage struct {
	storage map[string]models.Metric
	counter int64
}

func (m *MockStorage) Update(_ context.Context, _ any) error {
	return nil
}

func (m *MockStorage) ReadAll(_ context.Context) (map[string]models.Metric, error) {
	return m.storage, nil
}

func (m *MockStorage) GetVal(_ context.Context, k string) (models.Metric, error) {
	v, ok := m.storage[k]
	if !ok {
		return models.Metric{}, fmt.Errorf("metric %s not found", k)
	}

	return v, nil
}

func (m *MockStorage) SetVal(_ context.Context, k string, metric models.Metric) error {
	m.storage[k] = metric
	return nil
}

type mockFileStore struct{}

func (m mockFileStore) Save(_ services.Repository) error {
	return nil
}

func (m mockFileStore) Load(_ services.Repository) error {
	return nil
}

func TestGetMetric(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{
			"Metric1": {
				Type: "gauge",
				Val:  2.5,
			},
			"Metric2": {
				Type: "counter",
				Val:  int64(3),
			},
		},
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "test 1, get Metric1",
			request: "/value/gauge/Metric1",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        "2.5",
			},
		},
		{
			name:    "test 2, get Metric2",
			request: "/value/counter/Metric2",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        "3",
			},
		},
		{
			name:    "test 3, get unknown Metric3",
			request: "/value/gauge/Metric3",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "metric Metric3 not found\n",
			},
		},
		{
			name:    "test 4, get unknown Metric4",
			request: "/value/counter/Metric4",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "metric Metric4 not found\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   mockFileStore{},
				log:  zerolog.Logger{},
			}

			route.Get("/value/{type}/{name}", h.GetMetric)
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(http.MethodGet, s.URL+tt.request, nil)
			require.NoError(t, err)
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{},
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "test 1, update Metric1",
			request: "/update/gauge/Metric1/234.324",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 2, update Metric2",
			request: "/update/counter/Metric2/213",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 3, bad request",
			request: "/update/someType/Metric3/34",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   mockFileStore{},
				log:  zerolog.Logger{},
			}

			route.Post("/update/{type}/{name}/{val}", h.UpdateMetric)
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(http.MethodPost, s.URL+tt.request, nil)
			require.NoError(t, err)
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestDefaultHandler(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{},
	}

	tests := []struct {
		name    string
		method  string
		request string
		want    want
	}{
		{
			name:    "test 1, get all metrics",
			method:  http.MethodGet,
			request: "/",
			want: want{
				code:        http.StatusOK,
				contentType: "text/html",
			},
		},
		{
			name:    "test 2, bad request",
			method:  http.MethodPost,
			request: "/",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 3, page not found",
			method:  http.MethodPost,
			request: "/update",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   mockFileStore{},
				log:  zerolog.Logger{},
			}

			route.Get("/", h.Default)
			route.Post("/", h.Default)
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(tt.method, s.URL+tt.request, nil)
			require.NoError(t, err)
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetricJSON(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{
			"Metric1": {
				Type: "gauge",
				Val:  2.5,
			},
			"Metric2": {
				Type: "counter",
				Val:  int64(3),
			},
		},
	}

	tests := []struct {
		name    string
		method  string
		request string
		body    string
		want    want
	}{
		{
			name:    "test 1, get Metric1",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric1", "type": "gauge"}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric1", "type": "gauge", "value": 2.5}`,
			},
		},
		{
			name:    "test 2, get Metric2",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric2", "type": "counter"}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric2", "type": "counter", "delta": 3}`,
			},
		},
		{
			name:    "test 3, get unknown Metric3",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric3", "type": "counter"}`,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   mockFileStore{},
				log:  zerolog.Logger{},
			}

			route.Route("/value", func(r chi.Router) {
				r.Post("/", h.GetMetricJSON)
			})
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(tt.method, s.URL+tt.request, bytes.NewReader([]byte(tt.body)))
			require.NoError(t, err)
			r.Header.Set("Content-Type", "application/json")
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			var buf bytes.Buffer
			buf.ReadFrom(resp.Body)

			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, buf.String())
			}
		})
	}
}

func TestUpdateMetricJSON(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{},
	}

	tests := []struct {
		name    string
		method  string
		request string
		body    string
		want    want
	}{
		{
			name:    "test 1, add Metric1",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric1", "type": "gauge", "value": 2.5}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric1", "type": "gauge", "value": 2.5}`,
			},
		},
		{
			name:    "test 2, update Metric1",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric1", "type": "gauge", "value": 3.5}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric1", "type": "gauge", "value": 3.5}`,
			},
		},
		{
			name:    "test 3, update unknown Metric2",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric2", "type": "gauge", "value": 3.5}`,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   nil,
				log:  zerolog.Logger{},
			}

			route.Route("/update", func(r chi.Router) {
				r.Post("/", h.UpdateMetricJSON)
			})
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(tt.method, s.URL+tt.request, bytes.NewReader([]byte(tt.body)))
			require.NoError(t, err)
			r.Header.Set("Content-Type", "application/json")
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			var buf bytes.Buffer
			buf.ReadFrom(resp.Body)

			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, buf.String())
			}
		})
	}
}

func TestUpdateMetricsBatch(t *testing.T) {
	storage := &MockStorage{
		storage: map[string]models.Metric{},
	}

	tests := []struct {
		name    string
		method  string
		request string
		body    string
		want    want
	}{
		{
			name:    "test 1, update Metrics by batch",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric1", "type": "gauge", "value": 2.5},
			{"id": "Metric2", "type": "gauge", "value": 3.5}]`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric1", "type": "gauge", "value": 2.5}`,
			},
		},
		{
			name:    "test 2, update one Metric",
			method:  http.MethodPost,
			request: "/updates/",
			body:    `{"id": "Metric1", "type": "gauge", "value": 2.5}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 3, update Metrics by batch, invalid type",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric1", "type": "invalid", "value": 2.5},
			{"id": "Metric2", "type": "gauge", "value": 3.5}]`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: storage,
				fs:   mockFileStore{},
				log:  zerolog.Logger{},
			}

			route.Post("/updates/", h.UpdateMetricsBatch)
			s := httptest.NewServer(route)
			defer s.Close()

			r, err := http.NewRequest(tt.method, s.URL+tt.request, bytes.NewReader([]byte(tt.body)))
			require.NoError(t, err)
			r.Header.Set("Content-Type", "application/json")
			resp, err := s.Client().Do(r)
			require.NoError(t, err)

			var buf bytes.Buffer
			buf.ReadFrom(resp.Body)

			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, buf.String())
			}
		})
	}
}
