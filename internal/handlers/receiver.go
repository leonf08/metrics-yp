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
		switch r.Header.Get("Content-Type") {
		case "text/plain":
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
			io.WriteString(w, vStr)
		case "application/json":
			var metrics models.Metrics
			if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if metrics.ID == "" {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			v, err := st.GetVal(metrics.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch metrics.MType {
			case "gauge":
				*metrics.Value = v.(float64)
			case "counter":
				*metrics.Delta = v.(int64)
			default:
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			if err = json.NewEncoder(w).Encode(&metrics); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(st storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Content-Type") {
		case "text/plain":
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
		case "application/json":
			var metrics models.Metrics
			if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
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
			if err := json.NewEncoder(w).Encode(&metrics); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
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
		for n, v := range st.ReadAll() {
			s += fmt.Sprintf("%s - %v\r\n", n, v)
		}

		io.WriteString(w, s)
	}
}