package logger

import (
	"net/http"
)

type Logger interface {
	Infoln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

type (
	ResponseData struct {
		Status int
		Size   int
	}

	LoggingResponse struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}

	Log struct {
		Logger
	}
)

func (l *LoggingResponse) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.ResponseData.Size += size
	return size, err
}

func (l *LoggingResponse) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.ResponseData.Status = statusCode
}

func NewLogger(l Logger) *Log {
	return &Log{Logger: l}
}
