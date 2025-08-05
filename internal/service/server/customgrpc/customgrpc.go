package customgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/pkg/crypto"
	"github.com/talx-hub/malerter/pkg/signature"
	pb "github.com/talx-hub/malerter/proto"
)

type Server struct {
	pb.UnimplementedMetricsServer
	storage    handlers.Storage
	log        *logger.ZeroLogger
	decrypter  *crypto.Decrypter
	grpcServer *grpc.Server
	subnet     *net.IPNet
	address    string
	secret     string
}

func New(
	storage handlers.Storage,
	log *logger.ZeroLogger,
	decrypter *crypto.Decrypter,
	address, secret string,
	subnet *net.IPNet,
) *Server {
	return &Server{
		address:   address,
		storage:   storage,
		log:       log,
		decrypter: decrypter,
		secret:    secret,
		subnet:    subnet,
	}
}

func (s *Server) Batch(ctx context.Context, r *pb.BatchRequest,
) (*pb.BatchResponse, error) {
	metrics := s.parseMetrics(r)

	if err := s.storeMetrics(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "storage error: %v", err)
	}

	return &pb.BatchResponse{}, nil
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		errMsg := "failed to start listening " + s.address
		s.log.Fatal().Err(err).Msg(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			NewCheckNetworkInterceptor(s.subnet, s.log),
			NewVerifySignatureInterceptor(s.secret, s.log),
			NewDecryptingInterceptor(s.decrypter, s.log),
		))
	pb.RegisterMetricsServer(s.grpcServer, s)

	errCh := make(chan error)
	defer close(errCh)

	go func() {
		errCh <- s.grpcServer.Serve(lis)
	}()

	return <-errCh
}

func (s *Server) Stop(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)

	go func() {
		s.grpcServer.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context closed: %w", ctx.Err())
	case <-done:
		return nil
	}
}

func (s *Server) parseMetrics(r *pb.BatchRequest) []model.Metric {
	protoMetrics := r.GetMetricList().GetMetrics()
	metrics := make([]model.Metric, len(protoMetrics))
	var j = 0
	for _, protoMetric := range protoMetrics {
		m, err := fromGRPC(protoMetric)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to parse metric")
			continue
		}
		metrics[j] = m
		j++
	}
	return metrics[:j]
}

func (s *Server) storeMetrics(ctx context.Context, batch []model.Metric) error {
	ctxTO, cancel := context.WithTimeout(ctx, constants.TimeoutStorage)
	defer cancel()
	wrappedBatch := func(args ...any) (any, error) {
		return nil, s.storage.Batch(ctxTO, batch)
	}
	if _, err := db.WithConnectionCheck(wrappedBatch); err != nil {
		errMsg := "failed to store batch in repo"
		s.log.Error().Err(err).Msg(errMsg)
		return status.Errorf(codes.Internal, "%s: %v", errMsg, err)
	}
	return nil
}

func fromGRPC(pbMetric *pb.Metric) (model.Metric, error) {
	switch pbMetric.GetType() {
	case pb.Metric_Gauge:
		return model.Metric{
			Delta: nil,
			Value: &pbMetric.Value,
			Type:  model.MetricTypeGauge,
			Name:  pbMetric.GetName(),
		}, nil
	case pb.Metric_Counter:
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

func NewDecryptingInterceptor(decrypter *crypto.Decrypter, log *logger.ZeroLogger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if decrypter == nil {
			return handler(ctx, req)
		}

		batchReq, ok := req.(*pb.BatchRequest)
		if !ok {
			errMsg := "request does not implement *pb.BathRequest: got " +
				fmt.Sprintf("%T", req)
			log.Error().Msg(errMsg)
			return nil, status.Errorf(
				codes.InvalidArgument, "wrong message format: %s", errMsg)
		}

		encrypted := batchReq.GetEncryptedPayload()
		if encrypted == nil {
			return handler(ctx, req)
		}

		data, err := decrypter.Decrypt(encrypted)
		if err != nil {
			log.Error().Err(err).Msg("decryption failed")
			return nil, status.Errorf(
				codes.InvalidArgument, "decryption failed: %v", err)
		}

		var decryptedPayload pb.MetricList
		if err := proto.Unmarshal(data, &decryptedPayload); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal decrypted data")
			return nil, status.Errorf(codes.InvalidArgument,
				"invalid decrypted format: %v", err)
		}

		newReq := &pb.BatchRequest{
			Payload: &pb.BatchRequest_MetricList{
				MetricList: &decryptedPayload,
			},
		}

		return handler(ctx, newReq)
	}
}

func NewVerifySignatureInterceptor(secret string, log *logger.ZeroLogger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if secret == constants.NoSecret {
			return handler(ctx, req)
		}

		batchReq, ok := req.(*pb.BatchRequest)
		if !ok {
			errMsg := "request does not implement *pb.BathRequest: got " +
				fmt.Sprintf("%T", req)
			log.Error().Msg(errMsg)
			return nil, status.Errorf(
				codes.InvalidArgument, "wrong message format: %s", errMsg)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(
				codes.Unauthenticated, "missing metadata")
		}

		signatures := md.Get("signature")
		if len(signatures) == 0 {
			return nil, status.Errorf(
				codes.Unauthenticated, "missing signature")
		}

		data, err := proto.Marshal(batchReq)
		if err != nil {
			log.Error().Err(err).Msg(
				"failed to marshal request for verification")
			return nil, status.Errorf(
				codes.Internal, "verify failed: %v", err)
		}

		hash := signature.Hash(data, secret)
		if hash != signatures[0] {
			log.Warn().Msg("signature verification failed")
			return nil, status.Errorf(
				codes.PermissionDenied, "invalid signature")
		}

		return handler(ctx, req)
	}
}

func NewCheckNetworkInterceptor(ipNet *net.IPNet, log *logger.ZeroLogger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if ipNet == nil {
			return handler(ctx, req)
		}

		const forbiddenMsg = "forbidden"
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error().Msg("metadata not found in context")
			return nil, status.Error(codes.PermissionDenied, forbiddenMsg)
		}

		realIPs := md.Get("x-real-ip")
		if len(realIPs) == 0 {
			log.Error().Msg("x-real-ip header not present")
			return nil, status.Error(codes.PermissionDenied, forbiddenMsg)
		}

		ip := net.ParseIP(realIPs[0])
		if ip == nil {
			log.Error().Str("x-real-ip", realIPs[0]).Msg("unable to parse x-real-ip")
			return nil, status.Error(codes.PermissionDenied, forbiddenMsg)
		}

		if !ipNet.Contains(ip) {
			log.Error().Str("ip", ip.String()).Msg("agent IP not in allowed subnet")
			return nil, status.Error(codes.PermissionDenied, forbiddenMsg)
		}

		return handler(ctx, req)
	}
}
