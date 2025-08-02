package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server/customgrpc"
	pb "github.com/talx-hub/malerter/proto"
)

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestGRPCSender_Send(t *testing.T) {
	metrics := []model.Metric{
		{Name: "m1", Type: model.MetricTypeGauge},
		{Name: "m2", Type: model.MetricTypeGauge, Value: ptrFloat64(3.14)},
		{Name: "m3", Type: model.MetricTypeGauge, Value: ptrFloat64(3.15)},
		{Name: "m4", Type: model.MetricTypeGauge, Value: ptrFloat64(3.16)},
		{Name: "m5", Type: model.MetricTypeGauge, Value: ptrFloat64(3.17)},
		{Name: "m6", Type: model.MetricTypeCounter, Delta: ptrInt64(42)},
		{Name: "m6", Type: model.MetricTypeCounter, Delta: ptrInt64(42)},
		{Name: "m7", Type: model.MetricTypeCounter, Delta: ptrInt64(21)},
	}
	wantCount := 6

	var j = make(chan model.Metric, len(metrics))
	for _, m := range metrics {
		j <- m
	}
	close(j)

	var jobs = make(chan chan model.Metric, 1)
	jobs <- j
	close(jobs)

	storage := memory.New(logger.NewNopLogger(), nil)
	const addr = "localhost:8081"
	srv := customgrpc.New(
		storage,
		logger.NewNopLogger(),
		nil,
		addr,
		constants.NoSecret,
		nil,
	)
	defer func() {
		ctxTO, cancel := context.WithTimeout(
			context.Background(),
			constants.TimeoutShutdown)
		defer cancel()
		err := srv.Shutdown(ctxTO)
		require.NoError(t, err)
	}()
	go func() {
		err := srv.ListenAndServe()
		require.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)
	sender, err := NewGRPCSender(
		logger.NewNopLogger(),
		nil,
		addr,
		constants.NoSecret,
	)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sender.Send(ctx, jobs, &wg)
	defer func() {
		err := sender.Close()
		require.NoError(t, err)
	}()
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	result, err := storage.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, wantCount, len(result))
}
func TestMarshalMessage_Valid(t *testing.T) {
	msg1 := &pb.BatchRequest{}
	data, err := marshalMessage(msg1)
	require.NoError(t, err)
	assert.Empty(t, data)

	msg2 := &pb.BatchRequest{
		Payload: &pb.BatchRequest_MetricList{
			MetricList: &pb.MetricList{
				Metrics: []*pb.Metric{
					{}, {},
				},
			}},
	}
	data, err = marshalMessage(msg2)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMarshalMessage_Invalid(t *testing.T) {
	data, err := marshalMessage("not a proto message")
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestSigningInterceptor_NoSecret(t *testing.T) {
	log := logger.NewNopLogger()
	interceptor := NewSigningInterceptor(constants.NoSecret, log)

	called := false
	err := interceptor(context.Background(), "/pb.Metrics/Send", &pb.BatchRequest{}, nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			called = true
			return nil
		},
	)

	require.NoError(t, err)
	assert.True(t, called)
}

func TestSigningInterceptor_WithSecret(t *testing.T) {
	log := logger.NewNopLogger()
	interceptor := NewSigningInterceptor("my-secret", log)

	var signature string
	err := interceptor(context.Background(), "/pb.Metrics/Send", &pb.BatchRequest{}, nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			signature = md.Get("signature")[0]
			return nil
		},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestSigningInterceptor_MarshalFails(t *testing.T) {
	log := logger.NewNopLogger()
	interceptor := NewSigningInterceptor("secret", log)

	err := interceptor(context.Background(), "/pb.Metrics/Send", "invalid req", nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)

	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, s.Code())
}

func TestEncryptingInterceptor_NilEncrypter(t *testing.T) {
	log := logger.NewNopLogger()
	interceptor := NewEncryptingInterceptor(nil, log)

	called := false
	err := interceptor(context.Background(), "/pb.Metrics/Send", &pb.BatchRequest{}, nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			called = true
			return nil
		},
	)

	require.NoError(t, err)
	assert.True(t, called)
}
