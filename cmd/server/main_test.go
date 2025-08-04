package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	l "github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server"
	"github.com/talx-hub/malerter/pkg/queue"
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
	//nolint:nilnil // it's tests
	return nil, nil
}

func (m *mockStorage) Ping(_ context.Context) error {
	return nil
}

func Test_metricDB_emptyDSN(t *testing.T) {
	logger := l.NewNopLogger()
	buffer := queue.New[model.Metric]()
	defer buffer.Close()

	database, err := metricDB(context.Background(), "", logger, &buffer)
	assert.Nil(t, database)
	assert.ErrorContains(t, err, "DB DSN is empty")
}

func Test_initStorage_fallbackToMemory(t *testing.T) {
	cfg := testConfig()
	cfg.DatabaseDSN = "bad-dsn"
	logger := l.NewNopLogger()
	buffer := queue.New[model.Metric]()
	defer buffer.Close()

	storage := initStorage(&cfg, logger, &buffer)
	_, ok := storage.(*memory.Memory)
	assert.True(t, ok)
}

func Test_shutdownServer_ok(t *testing.T) {
	storage := new(mockStorage)
	cfg := testConfig()
	logger := l.NewNopLogger()
	srv := server.Init(&cfg, storage, logger)
	cancelCalled := false
	err := shutdownServer(srv, func() { cancelCalled = true })
	assert.NoError(t, err)
	assert.True(t, cancelCalled)
}

func testConfig() serverCfg.Builder {
	return serverCfg.Builder{
		LogLevel:        "debug",
		TrustedSubnet:   "",
		DatabaseDSN:     "",
		RootAddress:     ":8080",
		StoreInterval:   time.Second * 10,
		FileStoragePath: "/tmp/test.json",
		Restore:         true,
		Secret:          "",
		CryptoKeyPath:   "",
	}
}
