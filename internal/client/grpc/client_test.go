package grpc

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClient(t *testing.T) {
	type args struct {
		a      services.Agent
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
				a:      &services.AgentService{},
				l:      zerolog.Logger{},
				config: agentconf.Config{},
			},
			want: &Client{
				agent:  &services.AgentService{},
				log:    zerolog.Logger{},
				config: agentconf.Config{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.a, tt.args.l, tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Start(t *testing.T) {
	mockAgent := mocks.NewAgent(t)
	mockAgent.On("GatherMetrics", mock.Anything).Return(nil)
	mockAgent.On("GetMetrics", mock.Anything).Return(make(map[string]models.Metric), nil)

	client := NewClient(mockAgent, zerolog.Logger{}, agentconf.Config{
		GRPCAddr:  "localhost:50051",
		PollInt:   1,
		ReportInt: 2,
		RateLim:   10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Start(ctx)
	assert.NotNil(t, err)

	mockAgent.AssertExpectations(t)
}
