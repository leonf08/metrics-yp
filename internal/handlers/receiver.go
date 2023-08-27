package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)


func GetMetric(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var vStr string
		name := chi.URLParam(r, "name")

		switch typeMetric := chi.URLParam(r, "type"); typeMetric {
		case "gauge":
			val, err := st.GetVal(name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			v := val.(storage.GaugeMetric)
			vStr = strconv.FormatFloat(float64(v), 'f', -1, 64)
		case "counter":
			val, err := st.GetVal(name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			v := val.(storage.CounterMetric)
			vStr = strconv.FormatInt(int64(v), 10)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, vStr)
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
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err = st.SetVal(name, v); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
		case "counter":
			v, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			if err = st.SetVal(name, v); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
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

		s := "Metric name - value\r\n"
		for n, v := range st.ReadAll() {
			s += fmt.Sprintf("%s - %v\r\n", n, v)
		}

		if r.Header.Get("Accept-Encoding") == "gzip" {
			w.Header().Set("Content-Type", "text/html")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, s)
	}
}

func GetMetricJSON(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := new(models.Metrics)
		if err := json.NewDecoder(r.Body).Decode(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if metrics.ID == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		v, err := st.GetVal(metrics.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		switch metrics.MType {
		case "gauge":
			val, ok := v.(storage.GaugeMetric)
			if !ok {
				http.Error(w, "Type assertion error", http.StatusInternalServerError)
				return
			}
			metrics.Value = new(float64)
			*metrics.Value = float64(val)
		case "counter":
			val, ok := v.(storage.CounterMetric)
			if !ok {
				http.Error(w, "Type assertion error", http.StatusInternalServerError)
				return
			}
			metrics.Delta = new(int64)
			*metrics.Delta = int64(val)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func UpdateMetricJSON(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := new(models.Metrics)
		if err := json.NewDecoder(r.Body).Decode(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		var v interface{}
		switch metrics.MType {
		case "gauge":
			v = *(metrics.Value)
		case "counter":
			v = *(metrics.Delta)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		st.SetVal(metrics.ID, v)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}