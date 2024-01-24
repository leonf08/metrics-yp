package middleware

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/services"
)

func Auth(s *services.HashSigner) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aw := w
			hashReq := r.Header.Get("HashSHA256")
			if s != nil && hashReq != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				calcHash, err := s.CalcHash(body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				getHash, err := hex.DecodeString(hashReq)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if ok := s.CheckHash(calcHash, getHash); !ok {
					http.Error(w, "invalid hash", http.StatusBadRequest)
					return
				}

				s.ResponseWriter = w
				aw = s
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			next.ServeHTTP(aw, r)
		})
	}
}
