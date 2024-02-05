package services

import (
	"context"
	"errors"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAgentService_GatherMetrics(t *testing.T) {
	r := mocks.NewRepository(t)

	type fields struct {
		mode string
		r    repo.Repository
	}

	type args struct {
		ctx context.Context
	}

	r.On("Update", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, m any) error {
			if ctx.Value("update") == "error" {
				return errors.New("error")
			}

			return nil
		})

	r.On("SetVal", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string, m models.Metric) error {
			if ctx.Value("setval") == "error" {
				return errors.New("error")
			}

			return nil
		})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test 1, gather metrics, no error",
			fields: fields{
				mode: "json",
				r:    r,
			},
			args: args{
				ctx: context.WithValue(context.Background(), "update", "no error"),
			},
			wantErr: false,
		},
		{
			name: "test 2, gather metrics, error in update",
			fields: fields{
				mode: "json",
				r:    r,
			},
			args: args{
				ctx: context.WithValue(context.Background(), "update", "error"),
			},
			wantErr: true,
		},
		{
			name: "test 3, gather metrics, error in setval",
			fields: fields{
				mode: "json",
				r:    r,
			},
			args: args{
				ctx: context.WithValue(context.Background(), "setval", "error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: tt.fields.mode,
				repo: tt.fields.r,
			}
			if err := a.GatherMetrics(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("GatherMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentService_ReportMetrics(t *testing.T) {
	type fields struct {
		mode string
		repo repo.Repository
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test 1, report metrics, error",
			fields: fields{
				mode: "invalid",
				repo: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: tt.fields.mode,
				repo: tt.fields.repo,
			}
			_, err := a.ReportMetrics(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAgentService_batchMetrics(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	r := mocks.NewRepository(t)
	r.On("ReadAll", mock.Anything).
		Return(func(ctx context.Context) (map[string]models.Metric, error) {
			m := make(map[string]models.Metric, 1)

			switch ctx.Value("readall") {
			case "error":
				return nil, errors.New("error")
			case "gaugeError":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  "3.2",
				}
			case "counterError":
				m["test"] = models.Metric{
					Type: "counter",
					Val:  "3",
				}
			case "unknownType":
				m["test"] = models.Metric{
					Type: "unknown",
					Val:  "3",
				}
			case "no error":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  3.2,
				}
				m["test2"] = models.Metric{
					Type: "counter",
					Val:  int64(3),
				}
			}

			return m, nil
		})

	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "test 1, batch metrics, no error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "no error"),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "test 2, batch metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, batch metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, batch metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, batch metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "unknownType"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: "batch",
				repo: r,
			}
			got, err := a.batchMetrics(tt.args.ctx)

			assert.Equal(t, tt.wantLen, len(got))
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAgentService_jsonMetrics(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	r := mocks.NewRepository(t)
	r.On("ReadAll", mock.Anything).
		Return(func(ctx context.Context) (map[string]models.Metric, error) {
			m := make(map[string]models.Metric, 1)

			switch ctx.Value("readall") {
			case "error":
				return nil, errors.New("error")
			case "gaugeError":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  "3.2",
				}
			case "counterError":
				m["test"] = models.Metric{
					Type: "counter",
					Val:  "3",
				}
			case "unknownType":
				m["test"] = models.Metric{
					Type: "unknown",
					Val:  "3",
				}
			case "no error":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  3.2,
				}
				m["test2"] = models.Metric{
					Type: "counter",
					Val:  int64(3),
				}
			}

			return m, nil
		})

	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "test 1, json metrics, no error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "no error"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "test 2, json metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, json metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, json metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, json metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "unknownType"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: "json",
				repo: r,
			}
			got, err := a.jsonMetrics(tt.args.ctx)

			assert.Equal(t, tt.wantLen, len(got))
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAgentService_queryMetrics(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	r := mocks.NewRepository(t)
	r.On("ReadAll", mock.Anything).
		Return(func(ctx context.Context) (map[string]models.Metric, error) {
			m := make(map[string]models.Metric, 1)

			switch ctx.Value("readall") {
			case "error":
				return nil, errors.New("error")
			case "gaugeError":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  "3.2",
				}
			case "counterError":
				m["test"] = models.Metric{
					Type: "counter",
					Val:  "3",
				}
			case "unknownType":
				m["test"] = models.Metric{
					Type: "unknown",
					Val:  "3",
				}
			case "no error":
				m["test"] = models.Metric{
					Type: "gauge",
					Val:  3.2,
				}
				m["test2"] = models.Metric{
					Type: "counter",
					Val:  int64(3),
				}
			}

			return m, nil
		})

	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "test 1, json metrics, no error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "no error"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "test 2, json metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, json metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, json metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, json metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), "readall", "unknownType"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: "query",
				repo: r,
			}
			got, err := a.queryMetrics(tt.args.ctx)

			assert.Equal(t, tt.wantLen, len(got))
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
