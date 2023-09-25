package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/models"
)

type Repository interface {
	ReadAll(context.Context) (map[string]any, error)
	Update(context.Context, any) error
	SetVal(context.Context, string, any) error
	GetVal(context.Context, string) (any, error)
}

type Pinger interface {
	Ping() error
}

type Server struct {
	sv          *http.Server
	storage     Repository
	config      *serverconf.Config
	logger      logger.Logger
	fileStorage *storage.FileStorage
}

type ServerOption func(*Server) error

func NewServer(st Repository, cfg *serverconf.Config,
	logger logger.Logger) (*Server, error) {
	server := &Server{
		sv: &http.Server{
			Addr: cfg.Addr,
		},
		storage: st,
		config:  cfg,
		logger:  logger,
	}

	return server, nil
}

func (s *Server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	if s.config.IsFileStorage() {
		if s.config.Restore {
			s.logger.Infoln("Load metrics from file")
			m, err := s.fileStorage.LoadFromFile()
			if err != nil {
				return err
			}

			for k, v := range m.Storage {
				if err := s.storage.SetVal(ctx, k, v); err != nil {
					return err
				}
			}
		}

		if s.config.StoreInt > 0 {
			timer := time.NewTicker(time.Duration(s.config.StoreInt) * time.Second)
			defer timer.Stop()

			go func() {
				for {
					<-timer.C
					s.logger.Infoln("Save current metrics")
					if err := s.fileStorage.SaveInFile(s.storage); err != nil {
						s.logger.Errorln(err)
					}
				}
			}()
		}
	}

	g.Go(func() error {
		<-gCtx.Done()
		if s.config.IsFileStorage() {
			s.logger.Infoln("Save current metrics")
			if err := s.fileStorage.SaveInFile(s.storage); err != nil {
				s.logger.Errorln(err)
			}

			s.fileStorage.CloseFileStorage()
		}

		return s.sv.Shutdown(context.Background())
	})

	g.Go(func() error {
		s.logger.Infoln("Running server", "address", s.config.Addr)
		return s.sv.ListenAndServe()
	})

	return g.Wait()
}

func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	var vStr string
	name := chi.URLParam(r, "name")

	val, err := s.storage.GetVal(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
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

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	val := chi.URLParam(r, "val")

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = s.storage.SetVal(r.Context(), name, v); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = s.storage.SetVal(r.Context(), name, v); err != nil {
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

func (s *Server) Default(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	metrics, err := s.storage.ReadAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	str := "Metric name - value\r\n"
	for n, v := range metrics {
		str += fmt.Sprintf("%s - %v\r\n", n, v)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, str)
}

func (s *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metrics := new(models.MetricJSON)
	if err := json.NewDecoder(r.Body).Decode(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metrics.ID == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	v, err := s.storage.GetVal(r.Context(), metrics.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch metrics.MType {
	case "gauge":
		m, ok := v.(storage.Metric)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		val, ok := m.Val.(float64)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		metrics.Value = new(float64)
		*metrics.Value = val
	case "counter":
		m, ok := v.(storage.Metric)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		v, ok := m.Val.(int64)
		if !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

		metrics.Delta = new(int64)
		*metrics.Delta = v
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

func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric := new(models.MetricJSON)
	if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
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
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := s.storage.SetVal(r.Context(), metric.ID, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.config.IsFileStorage() && s.config.StoreInt == 0 {
		s.logger.Infoln("Save current metrics")
		if err := s.fileStorage.SaveInFile(s.storage); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	metrics := make([]models.MetricJSON, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricsDB := make([]storage.MetricsDB, len(metrics))
	for i, v := range metrics {
		metricsDB[i].Name = v.ID
		metricsDB[i].Type = v.MType
		switch v.MType {
		case "gauge":
			metricsDB[i].Val = *v.Value
		case "counter":
			metricsDB[i].Val = *v.Delta
		}
	}

	if err := s.storage.Update(r.Context(), metricsDB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) PingDB(w http.ResponseWriter, r *http.Request) {
	p, ok := s.storage.(Pinger)
	if !ok {
		http.Error(w, "not implemented", http.StatusInternalServerError)
		return
	}

	if err := p.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &logger.LoggingResponse{
			ResponseWriter: w,
			ResponseData:   &logger.ResponseData{},
		}

		start := time.Now()
		next.ServeHTTP(lw, r)
		duration := time.Since(start)

		s.logger.Infoln("Incoming HTTP request",
			"uri", r.URL.String(),
			"method", r.Method,
			"duration", duration)

		s.logger.Infoln("Response",
			"status", lw.ResponseData.Status,
			"size", lw.ResponseData.Size)
	})
}

func (s *Server) CompressMiddleware(next http.Handler) http.Handler {
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

func (s *Server) RegisterHandler(h http.Handler) {
	s.sv.Handler = h
}

func (s *Server) WithFileStorage() error {
	f, err := storage.NewFileStorage(s.config.FileStoragePath)
	if err != nil {
		return err
	}

	s.fileStorage = f
	return nil
}
