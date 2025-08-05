package customhttp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Add(ctx context.Context, metric model.Metric) error {
	args := m.Called(ctx, metric)
	//nolint:wrapcheck // it's tests
	return args.Error(0)
}

func (m *mockStorage) Batch(_ context.Context, _ []model.Metric) error {
	return nil
}

func (m *mockStorage) Find(_ context.Context, _ string) (model.Metric, error) {
	return model.Metric{}, nil
}

func (m *mockStorage) Get(_ context.Context) ([]model.Metric, error) {
	return nil, nil
}

func (m *mockStorage) Ping(_ context.Context) error {
	return nil
}

func TestNewCustomHTTP(t *testing.T) {
	log := logger.NewNopLogger()
	storage := new(mockStorage)

	srv := New(storage, log, nil, ":9999", "secret", nil)

	assert.Equal(t, ":9999", srv.Addr)
	assert.NotNil(t, srv.Handler)
}

func TestStop_OK(t *testing.T) {
	log := logger.NewNopLogger()
	storage := new(mockStorage)

	srv := New(storage, log, nil, ":0", "", nil)

	go func() {
		_ = srv.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := srv.Stop(ctx)
	assert.NoError(t, err)
}
