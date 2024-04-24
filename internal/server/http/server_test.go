package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	type args struct {
		h       http.Handler
		address string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test 1",
			args: args{
				h:       http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
				address: "localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(tt.args.h, tt.args.address)

			assert.Equal(t, tt.args.address, s.server.Addr)
			assert.NotNil(t, s.server.Handler)
			assert.NotNil(t, s.err)
		})
	}
}

func TestServer_Err(t *testing.T) {
	s := NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), "localhost:8080")

	assert.NotNil(t, s.Err())
}

func TestServer_Shutdown(t *testing.T) {
	s := NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), "localhost:8080")

	assert.Nil(t, s.Shutdown())
}
