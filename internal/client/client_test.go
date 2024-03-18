package client

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	type args struct {
		cl     *resty.Client
		a      services.Agent
		s      *services.HashSigner
		cr     services.Crypto
		l      zerolog.Logger
		config agentconf.Config
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "Test 1",
			args: args{
				cl:     &resty.Client{},
				a:      &services.AgentService{},
				s:      &services.HashSigner{},
				cr:     &services.CryptoService{},
				l:      zerolog.Logger{},
				config: agentconf.Config{},
			},
			want: &Client{
				client: &resty.Client{},
				agent:  &services.AgentService{},
				signer: &services.HashSigner{},
				crypto: &services.CryptoService{},
				log:    zerolog.Logger{},
				config: agentconf.Config{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewClient(tt.args.cl, tt.args.a, tt.args.s, tt.args.cr, tt.args.l, tt.args.config), "NewClient(%v, %v, %v, %v, %v, %v)", tt.args.cl, tt.args.a, tt.args.s, tt.args.cr, tt.args.l, tt.args.config)
		})
	}
}
