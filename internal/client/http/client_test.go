package http

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
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

func TestClient_Start(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mockAgent := mocks.NewAgent(t)
	mockAgent.On("GatherMetrics", ctx).Return(nil)
	mockAgent.On("ReportMetrics", ctx).Return([]string{"metric1", "metric2"}, nil)

	config := agentconf.Config{
		PollInt:   1,
		ReportInt: 2,
		RateLim:   10,
	}

	client := NewClient(resty.New(), mockAgent, &services.HashSigner{}, &services.CryptoService{}, zerolog.Logger{}, config)

	err := client.Start(ctx)
	assert.NotNil(t, err)
}

func TestGetIP(t *testing.T) {
	ip, err := GetIP()
	assert.Nil(t, err)
	assert.NotNil(t, ip)
	assert.IsType(t, net.IP{}, ip)
}
