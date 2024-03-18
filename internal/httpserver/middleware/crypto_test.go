package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	crypto := mocks.NewCrypto(t)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "Crypto middleware, no error",
			err:            nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Crypto middleware, decrypt error",
			err:            assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	r := chi.NewRouter()
	r.Use(Crypto(crypto))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crypto.On("Decrypt", []byte("test")).
				Return(func(src []byte) ([]byte, error) {
					if tt.err != nil {
						return nil, assert.AnError
					}

					return src, nil
				})

			resp, err := ts.Client().Post(ts.URL, "application/json", bytes.NewReader([]byte("test")))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
