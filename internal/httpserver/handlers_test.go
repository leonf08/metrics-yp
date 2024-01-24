package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type want struct {
	code        int
	contentType string
	body        string
}

func TestGetMetric(t *testing.T) {
	repo := mocks.NewRepository(t)

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
			},
		},
		{
			name:    "test 4, get unknown Metric4",
			request: "/value/counter/Metric4",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 5, invalid value",
			request: "/value/counter/Metric5",
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 6, invalid value",
			request: "/value/gauge/Metric6",
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 7, invalid type",
			request: "/value/invalid/Metric7",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	repo.On("GetVal", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string) (models.Metric, error) {
			switch k {
			case "Metric1":
				return models.Metric{
					Type: "gauge",
					Val:  2.5,
				}, nil
			case "Metric2":
				return models.Metric{
					Type: "counter",
					Val:  int64(3),
				}, nil
			case "Metric5":
				return models.Metric{
					Type: "counter",
					Val:  3.5,
				}, nil
			case "Metric6":
				return models.Metric{
					Type: "gauge",
					Val:  int64(3),
				}, nil
			case "Metric7":
				return models.Metric{}, nil
			default:
				return models.Metric{}, fmt.Errorf("metric %s not found", k)
			}
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   nil,
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
			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	repo := mocks.NewRepository(t)

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
			name:    "test 3, bad request, invalid type",
			request: "/update/someType/Metric3/34",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 4, bad request, invalid gauge value",
			request: "/update/gauge/Metric4/ff",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 5, counter metric not found",
			request: "/update/counter/Metric5/34",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 6, gauge metric not found",
			request: "/update/gauge/Metric6/3.4",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test 7, bad request, invalid counter value",
			request: "/update/counter/Metric7/ff",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	repo.On("SetVal", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string, metric models.Metric) error {
			switch k {
			case "Metric5":
				return fmt.Errorf("metric %s not found", k)
			case "Metric6":
				return fmt.Errorf("metric %s not found", k)
			}

			return nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   nil,
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
	repo := mocks.NewRepository(t)

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

	repo.On("ReadAll", mock.Anything).
		Return(func(ctx context.Context) (map[string]models.Metric, error) {
			return map[string]models.Metric{
				"Metric1": {},
			}, nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   nil,
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
	repo := mocks.NewRepository(t)

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
		{
			name:    "test 4, invalid request, incorrect json",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric4", "type": "counter"`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 5, invalid request, no id",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "", "type": "counter"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 6, invalid gauge value",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric5", "type": "gauge"}`,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 7, invalid counter value",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric6", "type": "counter"}`,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 8, invalid type",
			method:  http.MethodPost,
			request: "/value/",
			body:    `{"id": "Metric7", "type": "invalid"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}

	repo.On("GetVal", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string) (models.Metric, error) {
			switch k {
			case "Metric1":
				return models.Metric{
					Type: "gauge",
					Val:  2.5,
				}, nil
			case "Metric2":
				return models.Metric{
					Type: "counter",
					Val:  int64(3),
				}, nil
			case "Metric5":
				return models.Metric{
					Type: "gauge",
					Val:  "3.5",
				}, nil
			case "Metric6":
				return models.Metric{
					Type: "counter",
					Val:  "3.5",
				}, nil
			case "Metric7":
				return models.Metric{}, nil
			default:
				return models.Metric{}, fmt.Errorf("metric %s not found", k)
			}
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   nil,
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
	repo := mocks.NewRepository(t)
	fs := mocks.NewFileStore(t)

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
			name:    "test 2, add Metric2",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric2", "type": "counter", "delta": 2}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id": "Metric2", "type": "counter", "delta": 2}`,
			},
		},
		{
			name:    "test 3, update Metric1",
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
			name:    "test 4, update error",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric3", "type": "gauge", "value": 3.5}`,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 5, invalid request, incorrect json",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric4", "type": "counter"`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 6, invalid request, nil gauge value",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric4", "type": "gauge"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 7, invalid request, nil counter value",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric4", "type": "counter"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 8, invalid request, invalid type",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric4", "type": "type"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 9, save metrics error",
			method:  http.MethodPost,
			request: "/update/",
			body:    `{"id": "Metric5", "type": "gauge", "value": 3.5}`,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}

	repo.On("SetVal", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string, metric models.Metric) error {
			switch k {
			case "Metric3":
				return fmt.Errorf("error")
			}

			repo.TestData()[k] = metric

			return nil
		})

	fs.On("Save", mock.Anything).
		Return(func(r services.Repository) error {
			m := r.(*mocks.Repository)
			if _, ok := m.TestData()["Metric5"]; ok {
				return fmt.Errorf("error")
			}

			return nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   fs,
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
	repo := mocks.NewRepository(t)

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
			{"id": "Metric2", "type": "counter", "delta": 3}]`,
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
			name:    "test 3, update Metrics by batch, nil counter value",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric1", "type": "counter", "value": 2.5},
			{"id": "Metric2", "type": "gauge", "value": 3.5}]`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 4, update Metrics by batch, nil gauge value",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric1", "type": "gauge", "value": 2.5},
			{"id": "Metric2", "type": "gauge", "delta": 3}]`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 5, update Metrics by batch, invalid type",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric1", "type": "gauge", "value": 2.5},
			{"id": "Metric2", "type": "type", "delta": 3}]`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
		{
			name:    "test 6, update Metrics by batch, update error",
			method:  http.MethodPost,
			request: "/updates/",
			body: `[{"id": "Metric3", "type": "gauge", "value": 2.5},
			{"id": "Metric3", "type": "counter", "delta": 3}]`,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "",
			},
		},
	}

	repo.On("Update", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, any any) error {
			m := any.([]models.MetricDB)
			if m[0].Name == "Metric3" {
				return fmt.Errorf("error")
			}

			return nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := chi.NewRouter()

			h := &handler{
				repo: repo,
				fs:   nil,
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
