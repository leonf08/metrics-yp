package logger

import (
	"net/http"
	"time"
)

type Logger interface {
	Infoln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

type (
	responseData struct {
		status int
		size int
	}

	loggingResponse struct {
		http.ResponseWriter
		responseData *responseData
	}

	Log struct {
		Logger
	}
)

func (l *loggingResponse) Write(b []byte) (int, error) {
    size, err := l.ResponseWriter.Write(b)
    l.responseData.size += size
    return size, err
}

func (l *loggingResponse) WriteHeader(statusCode int) {
    l.ResponseWriter.WriteHeader(statusCode) 
    l.responseData.status = statusCode
}

func NewLogger(l Logger) *Log {
	return &Log{Logger: l}
}

func LoggingMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			lw := &loggingResponse{
				responseData: &responseData{},
			}
			lw.ResponseWriter = w

			start := time.Now()
			next.ServeHTTP(lw, r)
			duration := time.Since(start)

			log.Infoln("Incoming HTTP request",
					"uri", r.URL.String(),
					"method", r.Method,
					"duration", duration)	

			log.Infoln("Response",
					"status", lw.responseData.status,
					"size", lw.responseData.size)
		}

		return http.HandlerFunc(fn)
	}
}