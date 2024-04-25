package grpc

import (
	"context"

	"github.com/leonf08/metrics-yp.git/internal/models"
	proto2 "github.com/leonf08/metrics-yp.git/internal/proto"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type metricsServer struct {
	proto2.UnimplementedMetricsServer

	repo repo.Repository
	fs   services.FileStore
	log  zerolog.Logger
}

func newMetricsServer(repo repo.Repository, fs services.FileStore, log zerolog.Logger) *metricsServer {
	return &metricsServer{
		repo: repo,
		fs:   fs,
		log:  log,
	}
}

func (s *metricsServer) UpdateMetric(ctx context.Context, in *proto2.UpdateMetricRequest) (*proto2.UpdateMetricResponse, error) {
	logEntry := s.log.With().Str("method", "UpdateMetric").Logger()

	var response proto2.UpdateMetricResponse

	if in.Metric == nil {
		logEntry.Error().Msg("metric is required")
		return nil, status.Error(codes.InvalidArgument, "metric is required")
	}

	if in.Metric.Id == "" {
		logEntry.Error().Msg("metric id is required")
		return nil, status.Error(codes.InvalidArgument, "metric id is required")
	}

	if in.Metric.Type != "counter" && in.Metric.Type != "gauge" {
		logEntry.Error().Msg("invalid metric type")
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}

	err := s.repo.SetVal(ctx, in.Metric.Id, models.Metric{Type: in.Metric.Type, Val: in.Metric.Value})
	if err != nil {
		logEntry.Error().Err(err).Msg("failed to set metric")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if s.fs != nil {
		err = s.fs.Save(s.repo)
		if err != nil {
			logEntry.Error().Err(err).Msg("failed to write metric to file")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	response.Metric = in.Metric
	return &response, nil
}

func (s *metricsServer) GetMetric(ctx context.Context, in *proto2.GetMetricRequest) (*proto2.GetMetricResponse, error) {
	logEntry := s.log.With().Str("method", "GetMetric").Logger()

	var response proto2.GetMetricResponse

	if in.Id == "" {
		logEntry.Error().Msg("metric id is required")
		return nil, status.Error(codes.InvalidArgument, "metric id is required")
	}

	metric, err := s.repo.GetVal(ctx, in.Id)
	if err != nil {
		logEntry.Error().Err(err).Msg("failed to get metric")
		return nil, status.Error(codes.NotFound, err.Error())
	}

	response.Metric = &proto2.Metric{
		Id:    in.Id,
		Type:  metric.Type,
		Value: metric.Val.(float64),
	}

	return &response, nil
}
