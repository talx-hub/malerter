package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	l "github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/pkg/queue"
)

func Test_parseTrustedSubnet_valid(t *testing.T) {
	cfg := testConfig()
	cfg.TrustedSubnet = "192.168.0.0/24"
	subnet, err := parseTrustedSubnet(&cfg)
	require.NoError(t, err)
	assert.NotNil(t, subnet)
}

func Test_parseTrustedSubnet_empty(t *testing.T) {
	cfg := testConfig()
	cfg.TrustedSubnet = ""
	subnet, err := parseTrustedSubnet(&cfg)
	require.NoError(t, err)
	assert.Nil(t, subnet)
}

func Test_parseTrustedSubnet_invalid(t *testing.T) {
	cfg := testConfig()
	cfg.TrustedSubnet = "invalid-cidr"
	_, err := parseTrustedSubnet(&cfg)
	require.Error(t, err)
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
	srv := &http.Server{
		Addr: ":0",
	}
	cancelCalled := false
	err := shutdownServer(srv, func() { cancelCalled = true })
	assert.NoError(t, err)
	assert.True(t, cancelCalled)
}

func Test_initHTTPServer(t *testing.T) {
	cfg := testConfig()
	logger := l.NewNopLogger()
	buffer := queue.New[model.Metric]()
	defer buffer.Close()
	storage := initStorage(&cfg, logger, &buffer)
	srv := initHTTPServer(&cfg, logger, nil, storage, nil)
	assert.Equal(t, cfg.RootAddress, srv.Addr)
	assert.NotNil(t, srv.Handler)
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
