package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
)

type handler struct {
	repo repo.Repository
	fs   services.FileStore
	log  zerolog.Logger
}

func newHandler(r *chi.Mux, repo repo.Repository, fs services.FileStore, l zerolog.Logger) {
	h := handler{
		repo: repo,
		fs:   fs,
		log:  l,
	}

	r.Get("/", h.defaultHandler)
	r.Post("/", h.defaultHandler)
	r.Get("/ping", h.pingDB)
	r.Post("/updates/", h.updateMetricsBatch)
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", h.getMetric)
		r.Post("/", h.getMetricJSON)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.updateMetricJSON)
		r.Post("/{type}/{name}/{val}", h.updateMetric)
	})
}

// getMetric handles GET requests to /value/{type}/{name} endpoint to get metric value.
// Type and name of metric are passed as URL parameters.
// Response contains metric value in plain text format.
func (h handler) getMetric(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/getMetric").Logger()

	var vStr string
	name := chi.URLParam(r, "name")

	metric, err := h.repo.GetVal(r.Context(), name)
	if err != nil {
		logEntry.Error().Err(err).Msg("GetVal")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, ok := metric.Val.(float64)
		if !ok {
			logEntry.Error().Msg("invalid value type")
			http.Error(w, "invalid value type", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatFloat(v, 'f', -1, 64)
	case "counter":
		v, ok := metric.Val.(int64)
		if !ok {
			logEntry.Error().Msg("invalid value type")
			http.Error(w, "invalid value type", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatInt(v, 10)
	default:
		logEntry.Error().Msg("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, vStr)
}

// updateMetric handles POST requests to /update/{type}/{name}/{val} endpoint to update metric value.
// Type, name and value of metric are passed as URL parameters.
func (h handler) updateMetric(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/updateMetric").Logger()

	name := chi.URLParam(r, "name")
	val := chi.URLParam(r, "val")

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			logEntry.Error().Err(err).Msg("ParseFloat")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = h.repo.SetVal(r.Context(), name, models.Metric{Type: "gauge", Val: v}); err != nil {
			logEntry.Error().Err(err).Msg("SetVal")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			logEntry.Error().Err(err).Msg("ParseInt")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = h.repo.SetVal(r.Context(), name, models.Metric{Type: "counter", Val: v}); err != nil {
			logEntry.Error().Err(err).Msg("SetVal")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	default:
		logEntry.Error().Msg("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// defaultHandler handles GET requests to / endpoint to get all metrics.
// Response contains list of all measured metrics in plain text format.
func (h handler) defaultHandler(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/defaultHandler").Logger()

	if r.Method == http.MethodPost {
		logEntry.Error().Msg("method not allowed")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	metrics, err := h.repo.ReadAll(r.Context())
	if err != nil {
		logEntry.Error().Err(err).Msg("ReadAll")
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
		logEntry.Error().Err(err).Msg("Write")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getMetricJSON handles POST requests to /value endpoint to get metric value.
// Response contains metric object in JSON format
func (h handler) getMetricJSON(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/getMetricJSON").Logger()

	metric := models.MetricJSON{}
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logEntry.Error().Err(err).Msg("Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		logEntry.Error().Msg("bad request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	m, err := h.repo.GetVal(r.Context(), metric.ID)
	if err != nil {
		logEntry.Error().Err(err).Msg("GetVal")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch metric.MType {
	case "gauge":
		val, ok := m.Val.(float64)
		if !ok {
			logEntry.Error().Msg("invalid type assertion")
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
				logEntry.Error().Msg("invalid type assertion")
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			val = int64(valFloat)
		}

		metric.Delta = new(int64)
		*metric.Delta = val
	default:
		logEntry.Error().Msg("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		logEntry.Error().Err(err).Msg("Encode")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// updateMetricJSON handles POST requests to /update endpoint to update metric value.
// Updates metric object in JSON format received in request body.
// Response contains updated metric object in JSON format
func (h handler) updateMetricJSON(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/updateMetricJSON").Logger()

	metric := models.MetricJSON{}
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logEntry.Error().Err(err).Msg("Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var v any
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			logEntry.Error().Msg("invalid metric value")
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
		v = *(metric.Value)
	case "counter":
		if metric.Delta == nil {
			logEntry.Error().Msg("invalid metric value")
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
		v = *(metric.Delta)
	default:
		logEntry.Error().Msg("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if err := h.repo.SetVal(r.Context(), metric.ID, models.Metric{Type: metric.MType, Val: v}); err != nil {
		logEntry.Error().Err(err).Msg("SetVal")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.fs != nil {
		logEntry.Info().Msg("Save metrics to file")
		if err := h.fs.Save(h.repo); err != nil {
			logEntry.Error().Err(err).Msg("Save")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		logEntry.Error().Err(err).Msg("Encode")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// updateMetricsBatch handles POST requests to /updates endpoint to update batch of metrics.
// Updates metrics according to received JSON array of metric objects.
// Response contains first updated metric object in JSON format
func (h handler) updateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	logEntry := h.log.With().Str("component", "handler/updateMetricsBatch").Logger()

	metrics := make([]models.MetricJSON, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		logEntry.Error().Err(err).Msg("Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricsDB := make([]models.MetricDB, len(metrics))
	for i, v := range metrics {
		metricsDB[i].Name = v.ID
		metricsDB[i].Type = v.MType
		switch v.MType {
		case "gauge":
			if v.Value == nil {
				logEntry.Error().Msg("invalid metric value")
				http.Error(w, "invalid metric type", http.StatusBadRequest)
				return
			}
			metricsDB[i].Val = *v.Value
		case "counter":
			if v.Delta == nil {
				logEntry.Error().Msg("invalid metric value")
				http.Error(w, "invalid metric type", http.StatusBadRequest)
				return
			}
			metricsDB[i].Val = *v.Delta
		default:
			logEntry.Error().Msg("invalid metric type")
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
	}

	if err := h.repo.Update(r.Context(), metricsDB); err != nil {
		logEntry.Error().Err(err).Msg("Update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics[0]); err != nil {
		logEntry.Error().Err(err).Msg("Encode")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// pingDB handles GET requests to /ping endpoint to check DB connection.
func (h handler) pingDB(w http.ResponseWriter, _ *http.Request) {
	logEntry := h.log.With().Str("component", "handler/pingDB").Logger()

	p, ok := h.repo.(services.Pinger)
	if !ok {
		logEntry.Error().Msg("not implemented")
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	if err := p.Ping(); err != nil {
		logEntry.Error().Err(err).Msg("Ping")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
