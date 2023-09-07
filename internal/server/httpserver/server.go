package httpserver

import (
	"net/http"

	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"golang.org/x/exp/slices"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/models"
)

type Server struct {
	Storage storage.Repository
	Config  *serverconf.Config
	Saver   *Saver
	Loader  *Loader
	Logger  logger.Logger
}

func NewServer(st storage.Repository, cfg *serverconf.Config,
	sv *Saver, ld *Loader, logger logger.Logger) *Server {
	return &Server{
		Storage: st,
		Config:  cfg,
		Saver:   sv,
		Loader:  ld,
		Logger:  logger,
	}
}

func (server Server) Run(h http.Handler) error {
	server.Logger.Infoln("Running server", "address", server.Config.Addr)
	return http.ListenAndServe(server.Config.Addr, h)
}

func (server *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	var vStr string
	name := chi.URLParam(r, "name")

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		val, err := server.Storage.GetVal(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		m, ok := val.(storage.Metric)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		v, ok := m.Val.(float64)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		vStr = strconv.FormatFloat(v, 'f', -1, 64)
	case "counter":
		val, err := server.Storage.GetVal(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		m, ok := val.(storage.Metric)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		v, ok := m.Val.(int64)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		vStr = strconv.FormatInt(v, 10)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, vStr)
}

func (server *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	val := chi.URLParam(r, "val")

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = server.Storage.SetVal(name, v); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = server.Storage.SetVal(name, v); err != nil {
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

func (server *Server) Default(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	s := "Metric name - value\r\n"
	for n, v := range server.Storage.ReadAll() {
		s += fmt.Sprintf("%s - %v\r\n", n, v)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (server *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metrics := new(models.Metrics)
	if err := json.NewDecoder(r.Body).Decode(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metrics.ID == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	v, err := server.Storage.GetVal(metrics.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch metrics.MType {
	case "gauge":
		val, ok := v.(float64)
		if !ok {
			http.Error(w, "Type assertion error", http.StatusInternalServerError)
			return
		}
		metrics.Value = new(float64)
		*metrics.Value = float64(val)
	case "counter":
		val, ok := v.(int64)
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

func (server *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
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

	if err := server.Storage.SetVal(metrics.ID, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if server.Config.StoreInt == 0 {
		server.Logger.Infoln("Save current metrics")
		if err := server.Saver.SaveMetrics(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (server *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &logger.LoggingResponse{
			ResponseWriter: w,
			ResponseData:   &logger.ResponseData{},
		}

		start := time.Now()
		next.ServeHTTP(lw, r)
		duration := time.Since(start)

		server.Logger.Infoln("Incoming HTTP request",
			"uri", r.URL.String(),
			"method", r.Method,
			"duration", duration)

		server.Logger.Infoln("Response",
			"status", lw.ResponseData.Status,
			"size", lw.ResponseData.Size)
	})
}

func (server *Server) CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if slices.Contains(contentTypes, r.Header.Get("Accept")) {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					ow.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
		}

		next.ServeHTTP(ow, r)
	})
}
