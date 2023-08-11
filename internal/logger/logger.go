package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size int
	}

	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggerResponseWriter) Write(b []byte) (int, error) {
    size, err := r.ResponseWriter.Write(b) 
    r.responseData.size += size
    return size, err
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
    r.ResponseWriter.WriteHeader(statusCode) 
    r.responseData.status = statusCode
}

var Log *zap.Logger = zap.NewNop()

func InitLogger() error {
	lvl, err := zap.ParseAtomicLevel("info")
    if err != nil {
        return err
    }

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request)  {
		lw := &loggerResponseWriter{
			ResponseWriter: w,
			responseData: &responseData{},
		}

		start := time.Now()

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		Log.Info("Incoming HTTP request", 
			zap.String("url", r.URL.RawPath), 
			zap.String("method", r.Method),
			zap.String("duration", duration.String()))

		Log.Info("Response",
			zap.Int("status", lw.responseData.status),
			zap.Int("size", lw.responseData.size))
	})
}
