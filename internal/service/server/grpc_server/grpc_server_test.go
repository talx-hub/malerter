package grpc_server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/proto"
)

func initConn() (*grpc.ClientConn, error) {
	return grpc.Dial(":8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		metrics     []*proto.Metric
		wantMetrics int
		wantCode    codes.Code
	}{
		{
			metrics: []*proto.Metric{
				{Name: "m1", Type: proto.Metric_Gauge, Value: 3.14},
				{Name: "m2", Type: proto.Metric_Gauge, Value: 2.72},
				{Name: "m3", Type: proto.Metric_Counter, Value: 42},
				{Name: "m3", Type: proto.Metric_Counter, Value: 42},
			},
			wantMetrics: 3,
			wantCode:    codes.OK},
	}

	storage := memory.New(logger.NewNopLogger(), nil)
	srv := New(storage, logger.NewNopLogger())
	go func() {
		lis, err := Serve(":8082", srv)
		require.NoError(t, err)
		defer func() {
			err = lis.Close()
			require.NoError(t, err)
		}()
	}()
	time.Sleep(1 * time.Second)

	for _, tt := range tests {
		_, err := client.Batch(context.Background(), &proto.BatchRequest{
			Payload: &proto.BatchRequest_MetricList{
				MetricList: &proto.MetricList{
					Metrics: tt.metrics,
				},
			},
		})
		assert.NoError(t, err)
		result, err := storage.Get(context.Background())
		assert.Equal(t, tt.wantMetrics, len(result))
	}
}
