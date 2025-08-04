package customgrpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/pkg/signature"
	pb "github.com/talx-hub/malerter/proto"
)

const addr = "localhost:8085"

func initConn() (*grpc.ClientConn, error) {
	//nolint:wrapcheck // tests
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func TestServer_Batch(t *testing.T) {
	conn, err := initConn()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	client := pb.NewMetricsClient(conn)
	tests := []struct {
		metrics     []*pb.Metric
		wantMetrics int
		wantCode    codes.Code
	}{
		{
			metrics: []*pb.Metric{
				{Name: "m1", Type: pb.Metric_Gauge, Value: 3.14},
				{Name: "m2", Type: pb.Metric_Gauge, Value: 2.72},
				{Name: "m3", Type: pb.Metric_Counter, Value: 42},
				{Name: "m3", Type: pb.Metric_Counter, Value: 42},
			},
			wantMetrics: 3,
			wantCode:    codes.OK},
	}

	storage := memory.New(logger.NewNopLogger(), nil)
	srv := New(
		storage,
		logger.NewNopLogger(),
		nil,
		addr,
		constants.NoSecret,
		nil)
	defer func() {
		ctxTO, cancel := context.WithTimeout(
			context.Background(),
			constants.TimeoutShutdown)
		defer cancel()
		_ = srv.Stop(ctxTO)
	}()
	go func() {
		err := srv.Start()
		require.NoError(t, err)
	}()
	time.Sleep(1 * time.Second)

	for _, tt := range tests {
		_, err := client.Batch(context.Background(), &pb.BatchRequest{
			Payload: &pb.BatchRequest_MetricList{
				MetricList: &pb.MetricList{
					Metrics: tt.metrics,
				},
			},
		})
		require.NoError(t, err)
		result, err := storage.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, tt.wantMetrics, len(result))
	}
}

func TestNewVerifySignatureInterceptor_ValidSignature(t *testing.T) {
	log := logger.NewNopLogger()
	secret := "key"
	req := &pb.BatchRequest{Payload: &pb.BatchRequest_MetricList{}}
	data, _ := proto.Marshal(req)
	sig := signature.Hash(data, secret)

	md := metadata.Pairs("signature", sig)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := NewVerifySignatureInterceptor(secret, log)

	hit := false
	resp, err := interceptor(
		ctx,
		req,
		&grpc.UnaryServerInfo{},
		func(ctx context.Context, r interface{}) (interface{}, error) {
			hit = true
			return "ok", nil
		},
	)

	require.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "ok", resp)
}

func TestNewVerifySignatureInterceptor_InvalidSignature(t *testing.T) {
	log := logger.NewNopLogger()
	secret := "key"
	req := &pb.BatchRequest{Payload: &pb.BatchRequest_MetricList{}}
	md := metadata.Pairs("signature", "bad-sig")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := NewVerifySignatureInterceptor(secret, log)
	_, err := interceptor(
		ctx,
		req,
		&grpc.UnaryServerInfo{},
		nil,
	)

	assert.ErrorContains(t, err, "invalid signature")
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestNewCheckNetworkInterceptor_AllowedIP(t *testing.T) {
	log := logger.NewNopLogger()
	_, ipNet, _ := net.ParseCIDR("192.168.0.0/24")

	md := metadata.Pairs("x-real-ip", "192.168.0.42")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := NewCheckNetworkInterceptor(ipNet, log)

	hit := false
	resp, err := interceptor(
		ctx,
		"req",
		&grpc.UnaryServerInfo{},
		func(ctx context.Context, r interface{}) (interface{}, error) {
			hit = true
			return "ok", nil
		},
	)

	require.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "ok", resp)
}

func TestNewCheckNetworkInterceptor_ForbiddenIP(t *testing.T) {
	log := logger.NewNopLogger()
	_, ipNet, _ := net.ParseCIDR("192.168.0.0/24")

	md := metadata.Pairs("x-real-ip", "10.0.0.1")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := NewCheckNetworkInterceptor(ipNet, log)
	_, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{}, nil)

	assert.ErrorContains(t, err, "forbidden")
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
