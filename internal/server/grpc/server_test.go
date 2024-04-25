package grpc

import (
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	type args struct {
		repo          repo.Repository
		fs            services.FileStore
		log           zerolog.Logger
		address       string
		trustedSubnet string
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "Test 1",
			args: args{
				repo:          &repo.MemStorage{},
				fs:            &services.FileStorage{},
				log:           zerolog.Nop(),
				address:       "localhost:8080",
				trustedSubnet: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(tt.args.repo, tt.args.fs, tt.args.log, tt.args.address, tt.args.trustedSubnet)

			assert.NotNil(t, s.server)
			assert.NotNil(t, s.repo)
			assert.NotNil(t, s.fs)
		})
	}
}

func TestServer_Err(t *testing.T) {
	s := NewServer(&repo.MemStorage{}, &services.FileStorage{}, zerolog.Nop(), "localhost:8080", "")

	assert.NotNil(t, s.Err())
}
