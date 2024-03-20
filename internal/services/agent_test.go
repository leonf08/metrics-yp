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

type key struct{}

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
			if ctx.Value(key{}) == "error" {
				return errors.New("error")
			}

			return nil
		})

	r.On("SetVal", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, k string, m models.Metric) error {
			if ctx.Value(key{}) == "error" {
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
				ctx: context.WithValue(context.Background(), key{}, "no error"),
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
				ctx: context.WithValue(context.Background(), key{}, "error"),
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
				ctx: context.WithValue(context.Background(), key{}, "error"),
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

	type want struct {
		metricsLen int
		err        bool
	}

	r := mocks.NewRepository(t)

	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "test 1, report metrics, error",
			fields: fields{
				mode: "invalid",
				repo: nil,
			},
			want: want{
				metricsLen: 0,
				err:        true,
			},
		},
		{
			name: "test 2, report metrics, json",
			fields: fields{
				mode: "json",
				repo: r,
			},
			want: want{
				metricsLen: 2,
				err:        false,
			},
		},
		{
			name: "test 3, report metrics, query",
			fields: fields{
				mode: "query",
				repo: r,
			},
			want: want{
				metricsLen: 2,
				err:        false,
			},
		},
		{
			name: "test 4, report metrics, batch",
			fields: fields{
				mode: "batch",
				repo: r,
			},
			want: want{
				metricsLen: 1,
				err:        false,
			},
		},
	}

	r.On("ReadAll", mock.Anything).
		Return(func(ctx context.Context) (map[string]models.Metric, error) {
			m := make(map[string]models.Metric)
			m["test"] = models.Metric{
				Type: "gauge",
				Val:  3.2,
			}
			m["test2"] = models.Metric{
				Type: "counter",
				Val:  int64(3),
			}

			return m, nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentService{
				mode: tt.fields.mode,
				repo: tt.fields.repo,
			}
			m, err := a.ReportMetrics(context.Background())
			if (err != nil) != tt.want.err {
				t.Errorf("ReportMetrics() error = %v, wantErr %v", err, tt.want.err)
				return
			}

			assert.Equal(t, tt.want.metricsLen, len(m))
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

			switch ctx.Value(key{}) {
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
				ctx: context.WithValue(context.Background(), key{}, "no error"),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "test 2, batch metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, batch metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, batch metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, batch metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "unknownType"),
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

			switch ctx.Value(key{}) {
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
				ctx: context.WithValue(context.Background(), key{}, "no error"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "test 2, json metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, json metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, json metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, json metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "unknownType"),
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

			switch ctx.Value(key{}) {
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
				ctx: context.WithValue(context.Background(), key{}, "no error"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "test 2, json metrics, error",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 3, json metrics, invalid gauge value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "gaugeError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 4, json metrics, invalid counter value",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "counterError"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "test 5, json metrics, unknown metric type",
			args: args{
				ctx: context.WithValue(context.Background(), key{}, "unknownType"),
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

func TestNewAgentService(t *testing.T) {
	type args struct {
		mode string
		r    repo.Repository
	}

	tests := []struct {
		name string
		args args
		want *AgentService
	}{
		{
			name: "test 1, new agent service",
			args: args{
				mode: "json",
				r:    nil,
			},
			want: &AgentService{
				mode: "json",
				repo: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewAgentService(tt.args.mode, tt.args.r), "NewAgentService(%v, %v)", tt.args.mode, tt.args.r)
		})
	}
}
