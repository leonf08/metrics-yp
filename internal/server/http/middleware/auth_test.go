package middleware

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "test 1, valid hash",
			body:           "test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test 2, invalid hash",
			body:           "test2",
			expectedStatus: http.StatusBadRequest,
		},
	}

	s := services.NewHashSigner("test")
	hash, err := s.CalcHash([]byte("test"))
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Use(Auth(s))

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	req := resty.New().R()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := req.SetHeader("HashSHA256", hex.EncodeToString(hash)).
				SetBody(tt.body).
				Post(ts.URL)

			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode())
		})
	}
}
