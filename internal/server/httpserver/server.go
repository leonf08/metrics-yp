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

type Server struct {
	s       *http.Server
	storage Repository
	config  *serverconf.Config
	saver   *saver
	loader  *loader
	logger  logger.Logger
}

type ServerOption func(*Server) error

func NewServer(st Repository, cfg *serverconf.Config,
	logger logger.Logger, opts ...ServerOption) (*Server, error) {
	server := &Server{
		s: &http.Server{
			Addr: cfg.Addr,
		},
		storage: st,
		config:  cfg,
		logger:  logger,
	}

	for _, opt := range opts {
		if err := opt(server); err != nil {
			return nil, err
		}
	}

	return server, nil
}

func (server *Server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	if server.config.FileStoragePath != "" {
		if server.config.Restore {
			server.logger.Infoln("Load metrics from file")
			m, err := server.loader.loadMetrics()
			if err != nil {
				return err
			}

			for k, v := range m.Storage {
				if err := server.storage.SetVal(ctx, k, v); err != nil {
					return err
				}
			}
		}

		if server.config.StoreInt > 0 {
			timer := time.NewTicker(time.Duration(server.config.StoreInt) * time.Second)
			defer timer.Stop()

			go func() {
				for {
					<-timer.C
					server.logger.Infoln("Save current metrics")
					if err := server.saver.saveMetrics(); err != nil {
						server.logger.Errorln(err)
					}
				}
			}()
		}
	}

	g.Go(func() error {
		<-gCtx.Done()
		if server.config.FileStoragePath != "" {
			server.logger.Infoln("Save current metrics")
			if err := server.saver.saveMetrics(); err != nil {
				server.logger.Errorln(err)
			}
		}

		return server.s.Shutdown(context.Background())
	})

	g.Go(func() error {
		server.logger.Infoln("Running server", "address", server.config.Addr)
		return server.s.ListenAndServe()
	})

	return g.Wait()
}

func (server *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	var vStr string
	name := chi.URLParam(r, "name")

	val, err := server.storage.GetVal(r.Context(), name)
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

		if err = server.storage.SetVal(r.Context(), name, v); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = server.storage.SetVal(r.Context(), name, v); err != nil {
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

	metrics, err := server.storage.ReadAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	s := "Metric name - value\r\n"
	for n, v := range metrics {
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

	v, err := server.storage.GetVal(r.Context(), metrics.ID)
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

func (server *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metrics := new(models.Metrics)
	if err := json.NewDecoder(r.Body).Decode(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var v any
	switch metrics.MType {
	case "gauge":
		v = *(metrics.Value)
	case "counter":
		v = *(metrics.Delta)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := server.storage.SetVal(r.Context(), metrics.ID, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if server.config.StoreInt == 0 && server.saver != nil {
		server.logger.Infoln("Save current metrics")
		if err := server.saver.saveMetrics(); err != nil {
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

func (server *Server) CloseFile() {
	server.saver.close()
	server.loader.close()
}

func (server *Server) PingDB(w http.ResponseWriter, r *http.Request) {
	p, ok := server.storage.(storage.Pinger)
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

func (server *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &logger.LoggingResponse{
			ResponseWriter: w,
			ResponseData:   &logger.ResponseData{},
		}

		start := time.Now()
		next.ServeHTTP(lw, r)
		duration := time.Since(start)

		server.logger.Infoln("Incoming HTTP request",
			"uri", r.URL.String(),
			"method", r.Method,
			"duration", duration)

		server.logger.Infoln("Response",
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

func (server *Server) RegisterHandler(h http.Handler) {
	server.s.Handler = h
}

func WithSaverOpt() ServerOption {
	return func(s *Server) error {
		sv, err := newSaver(s.config.FileStoragePath, s.storage)
		if err != nil {
			return err
		}

		s.saver = sv
		return nil
	}
}

func WithLoaderOpt() ServerOption {
	return func(s *Server) error {
		ld, err := newLoader(s.config.FileStoragePath)
		if err != nil {
			return err
		}

		s.loader = ld
		return nil
	}
}
