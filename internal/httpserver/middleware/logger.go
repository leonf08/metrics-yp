package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// Logging logs requests and responses.
func Logging(l zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := l.With().Str("component", "middleware/logging").Logger()
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With().Str("method", r.Method).Str("url", r.URL.Path).Logger()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()
			defer func() {
				entry.Info().
					Int("status", ww.Status()).
					Int("size", ww.BytesWritten()).
					Str("duration", time.Since(t).String()).
					Msg("response")
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
