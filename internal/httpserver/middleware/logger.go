package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

func Logging(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := l.With(slog.String("component", "middleware/logger"))
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("url", r.URL.Path),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()
			defer func() {
				entry.Info(
					"response",
					slog.Int("status", ww.Status()),
					slog.Int("size", ww.BytesWritten()),
					slog.String("duration", time.Since(t).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
