package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIpCheck(t *testing.T) {
	type args struct {
		ip string
	}

	tests := []struct {
		name           string
		args           args
		expectedStatus int
	}{
		{
			name: "Trusted IP",
			args: args{
				ip: "192.168.1.4",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Untrusted IP",
			args: args{
				ip: "192.168.0.0",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Invalid IP",
			args: args{
				ip: "192.168.1",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	ipCheck, err := services.NewIPChecker("192.168.1.0/24")
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Use(IpCheck(ipCheck))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(r)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)

			req.Header.Set("X-Real-IP", tt.args.ip)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
