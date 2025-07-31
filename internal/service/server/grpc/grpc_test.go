package grpc

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
	"google.golang.org/grpc/status"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/proto"
)

const (
	scheme      = "metrics"
	serviceName = "metrics.service"
)

func initConn() (*grpc.ClientConn, error) {
	return grpc.Dial(":8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func TestServer_Batch(t *testing.T) {
	conn, err := initConn()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	client := proto.NewMetricsClient(conn)
	tests := []struct {
		metrics  []*proto.Metric
		wantCode codes.Code
	}{
		{
			metrics: []*proto.Metric{
				{Name: "m1", Type: proto.Metric_Gauge, Value: 3.14},
				{Name: "m2", Type: proto.Metric_Gauge, Value: 2.72},
				{Name: "m3", Type: proto.Metric_Counter, Value: 42},
				{Name: "m3", Type: proto.Metric_Counter, Value: 42},
			},
			wantCode: codes.OK},
		{
			metrics:  []*proto.Metric{},
			wantCode: codes.OK,
		},
	}

	storage := memory.New(logger.NewNopLogger(), nil)
	srv := New(storage, logger.NewNopLogger())
	go func() {
		err := Serve(t, srv)
		require.NoError(t, err)
	}()
	time.Sleep(5 * time.Second)

	var ctx context.Context
	var cancel context.CancelFunc
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	for _, tt := range tests {
		ctx, cancel = context.WithTimeout(context.Background(), constants.TimeoutStorage)
		_, err := client.Batch(ctx, &proto.BatchRequest{
			Metrics: tt.metrics,
		})
		cancel()
		if tt.wantCode == codes.OK {
			assert.NoError(t, err)
		} else {
			e, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantCode, e.Code())
		}
	}
}

func Serve(t *testing.T, srv *Server) error {
	t.Helper()

	lis, err := net.Listen("tcp", "localhost:8081")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	proto.RegisterMetricsServer(grpcServer, srv)

	return grpcServer.Serve(lis)
}
