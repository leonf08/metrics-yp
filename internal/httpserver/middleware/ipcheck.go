package middleware

import (
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/services"
)

func IPCheck(ip services.IPChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip != nil {
				addr := r.Header.Get("X-Real-IP")
				trusted, err := ip.IsTrusted(addr)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if !trusted {
					http.Error(w, "untrusted IP", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
