package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/services"
)

func Crypto(cr services.Crypto) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cr != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				decBody, err := cr.Decrypt(body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(decBody))
			}

			next.ServeHTTP(w, r)
		})
	}
}
