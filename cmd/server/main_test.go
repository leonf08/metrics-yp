package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	gaugeStorage map[string]storage.GaugeMetric
	counterStorage map[string]storage.CounterMetric
}

func (mem *MockStorage) WriteGaugeMetric(name string, val float64) {
	mem.gaugeStorage[name] = storage.GaugeMetric(val)
}

func (mem *MockStorage) WriteCounterMetric(name string, val int64) {
	mem.counterStorage[name] = storage.CounterMetric(val)
}

func (mem MockStorage) GetGaugeMetrics() map[string]storage.GaugeMetric {
	return mem.gaugeStorage
}

func (mem MockStorage) GetCounterMetrics() map[string]storage.CounterMetric {
	return mem.counterStorage
}

func (mem MockStorage) GetGaugeMetricVal(name string) (storage.GaugeMetric, bool) {
	v, ok := mem.gaugeStorage[name]
	return v, ok
}

func (mem MockStorage) GetCounterMetricVal(name string) (storage.CounterMetric, bool) {
	v, ok := mem.counterStorage[name]
	return v, ok
}

func (mem *MockStorage) UpdateGaugeMetrics() {
	
}

func (mem *MockStorage) UpdateCounterMetrics() {
	
}

func TestMetricsRouter(t *testing.T) {
	mockStorage := &MockStorage{
		gaugeStorage: map[string]storage.GaugeMetric{
			"someGaugeMetric": 3,
		},
		counterStorage: map[string]storage.CounterMetric{
			"someCounterMetric": 1,
		},
	}

	ts := httptest.NewServer(MetricsRouter(mockStorage))

	type want struct {
		status int
		contentType string
	}

	testCases := []struct {
		name string
		method string
		url string
		want want
	}{
		{
			name: "test 1, POST method, /",
			method: http.MethodPost,
			url: "/",
			want: want {
				status: http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 2, GET method, /",
			method: http.MethodGet,
			url: "/",
			want: want {
				status: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 3, GET method, get gauge metric value, metric present",
			method: http.MethodGet,
			url: "/value/gauge/someGaugeMetric",
			want: want {
				status: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 4, GET method, get gauge metric value, metric not present",
			method: http.MethodGet,
			url: "/value/gauge/someGaugeMetric2",
			want: want {
				status: http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 5, GET method, get counter metric value, metric present",
			method: http.MethodGet,
			url: "/value/counter/someCounterMetric",
			want: want {
				status: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 6, GET method, get counter metric value, metric not present",
			method: http.MethodGet,
			url: "/value/counter/someCounterMetric2",
			want: want {
				status: http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 7, POST method, update gauge metric",
			method: http.MethodPost,
			url: "/update/gauge/alloc/123",
			want: want {
				status: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 8, POST method, update counter metric",
			method: http.MethodPost,
			url: "/update/counter/value/123",
			want: want {
				status: http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 9, POST method, no value",
			method: http.MethodPost,
			url: "/update/counter/value",
			want: want {
				status: http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test 10, POST method, incorrect type",
			method: http.MethodPost,
			url: "/update/someType/value/43",
			want: want {
				status: http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testRequest(t, ts, tc.method, tc.url)
			defer resp.Body.Close()
			assert.Equal(t, tc.want.status, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, url string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+url, nil)
    require.NoError(t, err)

    resp, err := ts.Client().Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    return resp
}