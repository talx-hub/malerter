package customgrpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/proto"
)

const addr = ":8085"

func initConn() (*grpc.ClientConn, error) {
	//nolint:staticcheck,wrapcheck //i'm tired boss
	return grpc.Dial(
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
		_ = srv.Shutdown(ctxTO)
	}()
	go func() {
		err := srv.ListenAndServe()
		require.NoError(t, err)
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
		require.NoError(t, err)
		result, err := storage.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, tt.wantMetrics, len(result))
	}
}
