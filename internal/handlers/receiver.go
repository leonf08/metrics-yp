package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

//var serverStorage storage.MemStorage

func GetMetric(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v string
		name := chi.URLParam(r, "name")

		switch typeMetric := chi.URLParam(r, "type"); typeMetric {
		case "gauge":
			if val, ok := st.GetGaugeMetricVal(name); ok {
				v = strconv.FormatFloat(float64(val), 'f', -1, 64)
			} else {
				http.Error(w, fmt.Sprintf("Metric %s not found", name), http.StatusNotFound)
				return
			}
		case "counter":
			if val, ok := st.GetCounterMetricVal(name); ok {
				v = strconv.FormatInt(int64(val), 10)
			} else {
				http.Error(w, fmt.Sprintf("Metric %s not found", name), http.StatusNotFound)
				return
			}
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, v)
	}
}

func UpdateMetric(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		val := chi.URLParam(r, "val")

		switch typeMetric := chi.URLParam(r, "type"); typeMetric {
		case "gauge":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			st.WriteGaugeMetric(name, v)
		case "counter":
			v, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			st.WriteCounterMetric(name, v)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func DefaultHandler(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		s := "Metric name - value\r\n"
		for n, v := range st.GetGaugeMetrics() {
			s += fmt.Sprintf("%s - %v\r\n", n, v)
		}

		for n, v := range st.GetCounterMetrics() {
			s += fmt.Sprintf("%s - %v\r\n", n, v)
		}

		io.WriteString(w, s)
	}
}