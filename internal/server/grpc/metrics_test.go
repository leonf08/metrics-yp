package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/proto"
	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestMetricsServer_UpdateMetric(t *testing.T) {
	r := mocks.NewRepository(t)
	fs := mocks.NewFileStore(t)

	type args struct {
		ctx context.Context
		in  *proto.UpdateMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "UpdateMetric with valid input",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: &proto.Metric{
						Id:    "test",
						Type:  "counter",
						Value: 1.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UpdateMetric with nil metric",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "UpdateMetric with empty id",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: &proto.Metric{
						Id:    "",
						Type:  "counter",
						Value: 1.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "UpdateMetric with invalid type",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: &proto.Metric{
						Id:    "test",
						Type:  "invalid",
						Value: 1.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "UpdateMetric with error from repo",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: &proto.Metric{
						Id:    "test1",
						Type:  "counter",
						Value: 1.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "UpdateMetric with error from filestore",
			args: args{
				ctx: context.Background(),
				in: &proto.UpdateMetricRequest{
					Metric: &proto.Metric{
						Id:    "test2",
						Type:  "counter",
						Value: 1.0,
					},
				},
			},
			wantErr: true,
		},
	}

	r.On("SetVal", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, id string, m models.Metric) error {
			if id == "test1" {
				return errors.New("error")
			}

			r.TestData().Set(id, m)

			return nil
		})
	fs.On("Save", mock.Anything).Return(func(repository repo.Repository) error {
		rp := repository.(*mocks.Repository)
		if rp.TestData().Has("test2") {
			return errors.New("error")
		}

		return nil
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMetricsServer(r, fs, zerolog.Nop())
			_, err := s.UpdateMetric(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsServer_GetMetric(t *testing.T) {
	r := mocks.NewRepository(t)

	type args struct {
		ctx context.Context
		in  *proto.GetMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "GetMetric with valid id",
			args: args{
				ctx: context.Background(),
				in: &proto.GetMetricRequest{
					Id: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "GetMetric with empty id",
			args: args{
				ctx: context.Background(),
				in: &proto.GetMetricRequest{
					Id: "",
				},
			},
			wantErr: true,
		},
		{
			name: "GetMetric with error from repo",
			args: args{
				ctx: context.Background(),
				in: &proto.GetMetricRequest{
					Id: "test1",
				},
			},
			wantErr: true,
		},
	}

	r.On("GetVal", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, id string) (models.Metric, error) {
			if id == "test1" {
				return models.Metric{}, errors.New("error")
			}

			return models.Metric{Type: "counter", Val: 1.0}, nil
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMetricsServer(r, nil, zerolog.Nop())
			_, err := s.GetMetric(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
