package httpserver

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"io"
	"net/http"
	"strconv"
)

type handler struct {
	repo services.Repository
	fs   services.FileStore
	log  services.Logger
}

func newHandler(r *chi.Mux, repo services.Repository, fs services.FileStore, l services.Logger) {
	h := handler{
		repo: repo,
		fs:   fs,
		log:  l,
	}

	r.Get("/", h.Default)
	r.Post("/", h.Default)
	r.Get("/ping", h.PingDB)
	r.Post("/updates", h.UpdateMetricsBatch)
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", h.GetMetric)
		r.Post("/", h.GetMetricJSON)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateMetricJSON)
		r.Post("/{type}/{name}/{val}", h.UpdateMetric)
	})
}

func (h handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	var vStr string
	name := chi.URLParam(r, "name")

	metric, err := h.repo.GetVal(r.Context(), name)
	if err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, ok := metric.Val.(float64)
		if !ok {
			h.log.Error("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatFloat(v, 'f', -1, 64)
	case "counter":
		v, ok := metric.Val.(int64)
		if !ok {
			h.log.Error("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatInt(v, 10)
	default:
		h.log.Error("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, vStr)
}

func (h handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	val := chi.URLParam(r, "val")

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			h.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = h.repo.SetVal(r.Context(), name, models.Metric{Type: "gauge", Val: v}); err != nil {
			h.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			h.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = h.repo.SetVal(r.Context(), name, models.Metric{Type: "counter", Val: v}); err != nil {
			h.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	default:
		h.log.Error("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h handler) Default(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.log.Error("bad request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	metrics, err := h.repo.ReadAll(r.Context())
	if err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Metric name - value\r\n"
	for n, v := range metrics {
		str += fmt.Sprintf("%s - %v\r\n", n, v)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(str))
	if err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric := models.MetricJSON{}
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		h.log.Error("bad request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	m, err := h.repo.GetVal(r.Context(), metric.ID)
	if err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch metric.MType {
	case "gauge":
		val, ok := m.Val.(float64)
		if !ok {
			h.log.Error("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		metric.Value = new(float64)
		*metric.Value = val
	case "counter":
		var val int64
		varInt, ok := m.Val.(int64)
		if ok {
			val = varInt
		} else {
			valFloat, ok := m.Val.(float64)
			if !ok {
				h.log.Error("invalid type assertion")
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			val = int64(valFloat)
		}

		metric.Delta = new(int64)
		*metric.Delta = val
	default:
		h.log.Error("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric := models.MetricJSON{}
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var v any
	switch metric.MType {
	case "gauge":
		v = *(metric.Value)
	case "counter":
		v = *(metric.Delta)
	default:
		h.log.Error("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if err := h.repo.SetVal(r.Context(), metric.ID, models.Metric{Type: metric.MType, Val: v}); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.fs != nil {
		h.log.Info("Save current metrics")
		if err := h.fs.Save(h.repo); err != nil {
			h.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	metrics := make([]models.MetricJSON, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricsDB := make([]models.MetricDB, len(metrics))
	for i, v := range metrics {
		metricsDB[i].Name = v.ID
		metricsDB[i].Type = v.MType
		switch v.MType {
		case "gauge":
			metricsDB[i].Val = *v.Value
		case "counter":
			metricsDB[i].Val = *v.Delta
		default:
			h.log.Error("invalid metric type")
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
	}

	if err := h.repo.Update(r.Context(), metricsDB); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics[0]); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handler) PingDB(w http.ResponseWriter, _ *http.Request) {
	p, ok := h.repo.(services.Pinger)
	if !ok {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	if err := p.Ping(); err != nil {
		h.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
