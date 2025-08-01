package agent

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/proto"
)

type GRPCSender struct {
	client   proto.MetricsClient
	conn     *grpc.ClientConn
	log      *logger.ZeroLogger
	secret   string
	compress bool
}

func NewGRPCSender(log *logger.ZeroLogger, host, secret string, compress bool) (*GRPCSender, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errMsg := "failed to init gRPC connection"
		log.Fatal().Err(err).Msg(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}
	client := proto.NewMetricsClient(conn)

	return &GRPCSender{
		conn:     conn,
		client:   client,
		log:      log,
		secret:   secret,
		compress: compress,
	}, nil
}

func (s *GRPCSender) Send(
	jobs <-chan chan model.Metric, m *sync.Mutex, wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		m.Lock()
		jobCount := len(jobs)
		if jobCount == 0 {
			m.Unlock()
			return
		}

		j, ok := <-jobs
		if !ok {
			m.Unlock()
			return
		}
		m.Unlock()

		batch := toProtoMetrics(j)

		ctx, cancel := context.WithTimeout(
			context.Background(), constants.TimeoutAgentRequest)
		_, err := s.client.Batch(ctx, &proto.BatchRequest{
			Metrics: batch,
		})
		cancel()
		if err != nil {
			if e, ok := status.FromError(err); ok {
				s.log.Error().Err(e.Err()).Msg(e.Message())
			} else {
				s.log.Error().Err(err).Msg("failed to parse error")
			}
		}
	}
}

func toProtoMetrics(ch <-chan model.Metric) []*proto.Metric {
	batch := make([]*proto.Metric, len(ch))
	var i int
	for m := range ch {
		protoM := &proto.Metric{}
		protoM.Name = m.Name
		switch m.Type {
		case model.MetricTypeCounter:
			if m.Delta == nil {
				continue
			}
			protoM.Type = proto.Metric_Counter
			protoM.Delta = *m.Delta
		case model.MetricTypeGauge:
			if m.Value == nil {
				continue
			}
			protoM.Type = proto.Metric_Gauge
			protoM.Value = *m.Value
		default:
			protoM.Type = proto.Metric_Unspecified
		}

		batch[i] = protoM
		i++
	}

	return batch[:i]
}

func (s *GRPCSender) Close() error {
	err := s.conn.Close()
	if err != nil {
		errMsg := "failed to close gRPC connection"
		s.log.Error().Err(err).Msg("failed to close gRPC connection")
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	return nil
}
