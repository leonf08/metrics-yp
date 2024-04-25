package middleware

import (
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

func TestCompress(t *testing.T) {
	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		isRequestBody   bool
		expectedStatus  int
	}{
		{
			name:            "no accept-encoding, no content-encoding",
			acceptEncoding:  "",
			contentEncoding: "",
			isRequestBody:   true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "accept-encoding: gzip, content-encoding: gzip",
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			isRequestBody:   true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "accept-encoding: gzip, no content-encoding",
			acceptEncoding:  "gzip",
			contentEncoding: "",
			isRequestBody:   true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "no accept-encoding, content-encoding: gzip",
			acceptEncoding:  "",
			contentEncoding: "gzip",
			isRequestBody:   true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "accept-encoding: deflate, content-encoding: deflate",
			acceptEncoding:  "deflate",
			contentEncoding: "deflate",
			isRequestBody:   true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "accept-encoding: gzip, content-encoding: gzip, no body",
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			isRequestBody:   false,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	r := chi.NewRouter()
	r.Use(Compress)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		s, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		b, _ := hex.DecodeString("1f8b08005299ca6502ff2b492d2e01000c7e7fd804000000")
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R().SetHeaders(map[string]string{
				"Accept-Encoding":  tt.acceptEncoding,
				"Content-Encoding": tt.contentEncoding,
				"Accept":           "application/json",
			}).
				SetHeader("Content-Encoding", tt.contentEncoding)

			if tt.isRequestBody {
				req = req.SetBody(b)
			}

			resp, err := req.Post(ts.URL)

			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode())
			}
		})
	}
}
