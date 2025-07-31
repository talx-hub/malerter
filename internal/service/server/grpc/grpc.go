package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/proto"
)

type Storage interface {
	Batch(context.Context, []model.Metric) error
}

type Server struct {
	proto.UnimplementedMetricsServer
	storage Storage
	log     *logger.ZeroLogger
}

func New(storage Storage, log *logger.ZeroLogger) *Server {
	return &Server{
		storage: storage,
		log:     log,
	}
}

func (s Server) Batch(ctx context.Context, r *proto.BatchRequest,
) (*proto.BatchResponse, error) {
	metrics := make([]model.Metric, len(r.Metrics))
	var j = 0
	for _, protoMetric := range r.Metrics {
		m, err := fromGRPC(protoMetric)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to parse metric")
			continue
		}
		metrics[j] = m
		j++
	}
	metrics = metrics[:j]

	ctxTO, cancel := context.WithTimeout(ctx, constants.TimeoutStorage)
	defer cancel()
	wrappedBatch := func(args ...any) (any, error) {
		return nil, s.storage.Batch(ctxTO, metrics)
	}
	if _, err := db.WithConnectionCheck(wrappedBatch); err != nil {
		s.log.Error().Err(err).Msg("failed to batch metrics in repo")
		return nil, status.Errorf(codes.Internal, "failed to batch metrics in repo: %v", err)
	}
	return &proto.BatchResponse{}, nil
}

func fromGRPC(pbMetric *proto.Metric) (model.Metric, error) {
	switch pbMetric.GetType() {
	case proto.Metric_Gauge:
		return model.Metric{
			Delta: nil,
			Value: &pbMetric.Value,
			Type:  model.MetricTypeGauge,
			Name:  pbMetric.GetName(),
		}, nil
	case proto.Metric_Counter:
		return model.Metric{
			Delta: &pbMetric.Delta,
			Value: nil,
			Type:  model.MetricTypeCounter,
			Name:  pbMetric.GetName(),
		}, nil
	default:
		return model.Metric{}, errors.New(
			"metric has unspecified type")
	}
}
