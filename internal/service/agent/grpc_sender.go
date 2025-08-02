package agent

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/pkg/crypto"
	pb "github.com/talx-hub/malerter/proto"
)

type GRPCSender struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
	log    *logger.ZeroLogger
}

func NewGRPCSender(log *logger.ZeroLogger, encrypter *crypto.Encrypter, host, secret string,
) (*GRPCSender, error) {
	//nolint:staticcheck //i'm tired boss
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			NewSigningInterceptor(secret, log),
			NewEncryptingInterceptor(encrypter, log),
		))
	if err != nil {
		errMsg := "failed to init gRPC connection"
		log.Fatal().Err(err).Msg(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}
	client := pb.NewMetricsClient(conn)

	return &GRPCSender{
		conn:   conn,
		client: client,
		log:    log,
	}, nil
}

func (s *GRPCSender) Send(ctx context.Context,
	jobs <-chan chan model.Metric, wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case metrics, ok := <-jobs:
			if !ok {
				return
			}
			s.doTheJob(metrics)
		case <-ctx.Done():
			return
		}
	}
}

func (s *GRPCSender) doTheJob(metrics chan model.Metric) {
	batch := toProtoMetrics(metrics)

	ctx, cancel := context.WithTimeout(
		context.Background(), constants.TimeoutAgentRequest)
	_, err := s.client.Batch(ctx, &pb.BatchRequest{
		Payload: &pb.BatchRequest_MetricList{
			MetricList: &pb.MetricList{
				Metrics: batch,
			},
		},
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

func toProtoMetrics(ch <-chan model.Metric) []*pb.Metric {
	batch := make([]*pb.Metric, len(ch))
	var i int
	for m := range ch {
		protoM := &pb.Metric{}
		protoM.Name = m.Name
		switch m.Type {
		case model.MetricTypeCounter:
			if m.Delta == nil {
				continue
			}
			protoM.Type = pb.Metric_Counter
			protoM.Delta = *m.Delta
		case model.MetricTypeGauge:
			if m.Value == nil {
				continue
			}
			protoM.Type = pb.Metric_Gauge
			protoM.Value = *m.Value
		default:
			protoM.Type = pb.Metric_Unspecified
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

func marshalMessage(grpcRequest any) ([]byte, error) {
	protoMsg, ok := grpcRequest.(*pb.BatchRequest)
	if !ok {
		return nil, fmt.Errorf(
			"request does not implement *pb.BathRequest: got %T", grpcRequest)
	}

	data, err := proto.Marshal(protoMsg)
	if err != nil {
		return nil, fmt.Errorf(
			"error in marshalling req to bytes: %w", err)
	}

	return data, nil
}

func NewSigningInterceptor(secret string, log *logger.ZeroLogger,
) grpc.UnaryClientInterceptor {
	interceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if secret == constants.NoSecret {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		data, err := marshalMessage(req)
		if err != nil {
			log.Error().Err(err).Msg("signing failed")
			return status.Errorf(
				codes.Internal, "signing failed: %v", err)
		}

		sig := trySign(data, secret)
		md := metadata.Pairs("signature", sig)
		mdCtx := metadata.NewOutgoingContext(ctx, md)

		return invoker(mdCtx, method, req, reply, cc, opts...)
	}

	return interceptor
}

func NewEncryptingInterceptor(encrypter *crypto.Encrypter, log *logger.ZeroLogger,
) grpc.UnaryClientInterceptor {
	interceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if encrypter == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		data, err := marshalMessage(req)
		if err != nil {
			log.Error().Err(err).Msg("encrypting failed")
			return status.Errorf(
				codes.Internal, "encrypting failed: %v", err)
		}

		encrypted, err := tryEncrypt(data, encrypter)
		if err != nil {
			log.Error().Err(err).Msg("encrypting failed")
			return status.Errorf(
				codes.Internal, "encrypting failed: %v", err)
		}
		md := metadata.Pairs("x-encrypted", "true")
		mdCtx := metadata.NewOutgoingContext(ctx, md)

		return invoker(mdCtx, method,
			&pb.BatchRequest{
				Payload: &pb.BatchRequest_EncryptedPayload{
					EncryptedPayload: encrypted,
				},
			},
			reply, cc, opts...)
	}

	return interceptor
}
