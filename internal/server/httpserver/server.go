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

type (
	Repository interface {
		ReadAll(context.Context) (map[string]any, error)
		Update(context.Context, any) error
		SetVal(context.Context, string, any) error
		GetVal(context.Context, string) (any, error)
	}

	Pinger interface {
		Ping() error
	}

	Server struct {
		sv      *http.Server
		storage Repository
		config  *serverconf.Config
		logger  logger.Logger
	}
)

func NewServer(st Repository, cfg *serverconf.Config,
	logger logger.Logger) *Server {
	return &Server{
		sv: &http.Server{
			Addr: cfg.Addr,
		},
		storage: st,
		config:  cfg,
		logger:  logger,
	}
}

func (s *Server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	if s.config.IsFileStorage() {
		g.Go(func() error {
			m, ok := s.storage.(*storage.MemStorage)
			if !ok {
				err := fmt.Errorf("invalid type assertion for in-memory storage")
				s.logger.Errorln(err)
				return err
			}

			if s.config.Restore {
				s.logger.Infoln("Load metrics from file")
				if err := m.LoadFromFile(); err != nil {
					s.logger.Errorln(err)
					return err
				}
			}

			if s.config.StoreInt > 0 {
				timer := time.NewTicker(time.Duration(s.config.StoreInt) * time.Second)
				defer timer.Stop()

				for {
					select {
					case <-gCtx.Done():
						return gCtx.Err()
					case <-timer.C:
						s.logger.Infoln("Save current metrics")
						if err := m.SaveInFile(); err != nil {
							s.logger.Errorln(err)
						}
					}
				}
			}

			return nil
		})
	}

	g.Go(func() error {
		<-gCtx.Done()
		switch st := s.storage.(type) {
		case *storage.MemStorage:
			if s.config.IsFileStorage() {
				s.logger.Infoln("Save current metrics")
				if err := st.SaveInFile(); err != nil {
					s.logger.Errorln(err)
				}
				st.CloseFileStorage()
			}
		case *storage.PGStorage:
			st.Close()
		}

		s.logger.Infoln("Server shutdown")
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
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch typeMetric := chi.URLParam(r, "type"); typeMetric {
	case "gauge":
		m, ok := val.(storage.Metric)
		if !ok {
			s.logger.Errorln("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		v, ok := m.Val.(float64)
		if !ok {
			s.logger.Errorln("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatFloat(v, 'f', -1, 64)
	case "counter":
		m, ok := val.(storage.Metric)
		if !ok {
			s.logger.Errorln("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		v, ok := m.Val.(int64)
		if !ok {
			s.logger.Errorln("invalid type assertion")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		vStr = strconv.FormatInt(v, 10)
	default:
		s.logger.Errorln("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
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
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = s.storage.SetVal(r.Context(), name, v); err != nil {
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "counter":
		v, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = s.storage.SetVal(r.Context(), name, v); err != nil {
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	default:
		s.logger.Errorln("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Default(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.logger.Errorln("bad request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	metrics, err := s.storage.ReadAll(r.Context())
	if err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	metric := new(models.MetricJSON)
	if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		s.logger.Errorln("bad request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	v, err := s.storage.GetVal(r.Context(), metric.ID)
	if err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	m, ok := v.(storage.Metric)
	if !ok {
		s.logger.Errorln("invalid type assertion")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	switch metric.MType {
	case "gauge":
		val, ok := m.Val.(float64)
		if !ok {
			s.logger.Errorln("invalid type assertion")
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
				s.logger.Errorln("invalid type assertion")
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			val = int64(valFloat)
		}

		metric.Delta = new(int64)
		*metric.Delta = val
	default:
		s.logger.Errorln("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric := new(models.MetricJSON)
	if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
		s.logger.Errorln(err)
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
		s.logger.Errorln("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if err := s.storage.SetVal(r.Context(), metric.ID, v); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.config.IsFileStorage() && s.config.StoreInt == 0 {
		s.logger.Infoln("Save current metrics")
		m, ok := s.storage.(*storage.MemStorage)
		if !ok {
			err := fmt.Errorf("invalid type assertion for in-memory storage")
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := m.SaveInFile(); err != nil {
			s.logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	metrics := make([]models.MetricJSON, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricsDB := make([]storage.MetricDB, len(metrics))
	for i, v := range metrics {
		metricsDB[i].Name = v.ID
		metricsDB[i].Type = v.MType
		switch v.MType {
		case "gauge":
			metricsDB[i].Val = *v.Value
		case "counter":
			metricsDB[i].Val = *v.Delta
		default:
			s.logger.Errorln("invalid metric type")
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
	}

	if err := s.storage.Update(r.Context(), metricsDB); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&metrics[0]); err != nil {
		s.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) PingDB(w http.ResponseWriter, _ *http.Request) {
	p, ok := s.storage.(Pinger)
	if !ok {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	if err := p.Ping(); err != nil {
		s.logger.Errorln(err)
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
					s.logger.Errorln(err)
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

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aw := w
		//if s.config.IsAuthKeyExists() {
		//	body, err := io.ReadAll(r.Body)
		//	if err != nil {
		//		s.logger.Errorln(err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//
		//	calcHash, err := auth.CalcHash(body, []byte(s.config.Key))
		//	if err != nil {
		//		s.logger.Errorln(err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//
		//	getHash, err := hex.DecodeString(r.Header.Get("HashSHA256"))
		//	if err != nil {
		//		s.logger.Errorln(err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//
		//	if !auth.CheckHash(calcHash, getHash) {
		//		s.logger.Errorln("hash check failed")
		//		w.WriteHeader(http.StatusBadRequest)
		//		return
		//	}
		//
		//	aw = auth.NewAuthentificator(w, []byte(s.config.Key))
		//	r.Body = io.NopCloser(bytes.NewReader(body))
		//}

		next.ServeHTTP(aw, r)
	})
}

func (s *Server) RegisterHandler(h http.Handler) {
	s.sv.Handler = h
}
