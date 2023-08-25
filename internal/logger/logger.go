package logger

import (
	"net/http"
	"time"
	"log/slog"
)

type (
	responseData struct {
		status int
		size int
		body string
	}

	loggingResponse struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (l *loggingResponse) Write(b []byte) (int, error) {
    size, err := l.ResponseWriter.Write(b)
	l.responseData.body = string(b)
    l.responseData.size += size
    return size, err
}

func (l *loggingResponse) WriteHeader(statusCode int) {
    l.ResponseWriter.WriteHeader(statusCode) 
    l.responseData.status = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponse{
			responseData: &responseData{},
		}
		lw.ResponseWriter = w

		start := time.Now()
		next.ServeHTTP(lw, r)
		duration := time.Since(start)

		slog.Info("Incoming HTTP request:",
				"uri", r.URL.String(),
				"method", r.Method,
				"duration", duration)	

		slog.Info("Response",
				"status", lw.responseData.status,
				"size", lw.responseData.size,
				"body", lw.responseData.body)
	})
}