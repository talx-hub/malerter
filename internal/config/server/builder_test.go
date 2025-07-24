package server

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_LoadFromEnv(t *testing.T) {
	_ = os.Setenv(EnvCryptoKeyPath, "/path/to/private.pem")
	_ = os.Setenv(EnvAddress, "127.0.0.1:9001")
	_ = os.Setenv(EnvLogLevel, "debug")
	_ = os.Setenv(EnvFileStoragePath, "data/backup.db")
	_ = os.Setenv(EnvStoreInterval, "60")
	_ = os.Setenv(EnvRestore, "true")
	_ = os.Setenv(EnvDatabaseDSN, "user:pass@tcp(localhost:3306)/dbname")
	_ = os.Setenv(EnvSecretKey, "my-secret")

	defer func() {
		_ = os.Unsetenv(EnvCryptoKeyPath)
		_ = os.Unsetenv(EnvAddress)
		_ = os.Unsetenv(EnvLogLevel)
		_ = os.Unsetenv(EnvFileStoragePath)
		_ = os.Unsetenv(EnvStoreInterval)
		_ = os.Unsetenv(EnvRestore)
		_ = os.Unsetenv(EnvDatabaseDSN)
		_ = os.Unsetenv(EnvSecretKey)
	}()

	b := &Builder{}
	b.LoadFromEnv()

	assert.Equal(t, "/path/to/private.pem", b.CryptoKeyPath)
	assert.Equal(t, "127.0.0.1:9001", b.RootAddress)
	assert.Equal(t, "debug", b.LogLevel)
	assert.Equal(t, "data/backup.db", b.FileStoragePath)
	assert.Equal(t, 60*time.Second, b.StoreInterval)
	assert.True(t, b.Restore)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/dbname", b.DatabaseDSN)
	assert.Equal(t, "my-secret", b.Secret)
}

func TestBuilder_IsValid_Positive(t *testing.T) {
	b := &Builder{
		StoreInterval: 10 * time.Second,
	}
	_, err := b.IsValid()
	assert.NoError(t, err)
}

func TestBuilder_IsValid_Negative(t *testing.T) {
	b := &Builder{
		StoreInterval: -1 * time.Second,
	}
	_, err := b.IsValid()
	assert.Error(t, err)
	assert.EqualError(t, err, "store interval must be positive")
}

func TestBuilder_Build(t *testing.T) {
	b := &Builder{
		CryptoKeyPath:   "/keys/private.pem",
		RootAddress:     "localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "backup.bk",
		StoreInterval:   30 * time.Second,
		Restore:         true,
		DatabaseDSN:     "dsn",
		Secret:          "secret",
	}

	cfg, ok := b.Build().(Builder)
	require.True(t, ok)

	assert.Equal(t, b.CryptoKeyPath, cfg.CryptoKeyPath)
	assert.Equal(t, b.RootAddress, cfg.RootAddress)
	assert.Equal(t, b.LogLevel, cfg.LogLevel)
	assert.Equal(t, b.FileStoragePath, cfg.FileStoragePath)
	assert.Equal(t, b.StoreInterval, cfg.StoreInterval)
	assert.Equal(t, b.Restore, cfg.Restore)
	assert.Equal(t, b.DatabaseDSN, cfg.DatabaseDSN)
	assert.Equal(t, b.Secret, cfg.Secret)
}

func TestFileStorageDefault(t *testing.T) {
	path := FileStorageDefault()
	assert.Contains(t, path, ".bk")
	assert.Greater(t, len(path), 10)
}
