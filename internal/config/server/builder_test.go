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
	_ = os.Setenv(EnvTrustedSubnet, "127.0.0.0/24")

	defer func() {
		_ = os.Unsetenv(EnvCryptoKeyPath)
		_ = os.Unsetenv(EnvAddress)
		_ = os.Unsetenv(EnvLogLevel)
		_ = os.Unsetenv(EnvFileStoragePath)
		_ = os.Unsetenv(EnvStoreInterval)
		_ = os.Unsetenv(EnvRestore)
		_ = os.Unsetenv(EnvDatabaseDSN)
		_ = os.Unsetenv(EnvSecretKey)
		_ = os.Unsetenv(EnvTrustedSubnet)
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
	assert.Equal(t, "127.0.0.0/24", b.TrustedSubnet)
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
		TrustedSubnet:   "127.0.0.0/24",
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
	assert.Equal(t, b.TrustedSubnet, cfg.TrustedSubnet)
}

func TestFileStorageDefault(t *testing.T) {
	path := FileStorageDefault()
	assert.Contains(t, path, ".bk")
	assert.Greater(t, len(path), 10)
}

func TestBuilder_LoadFromFile_Success(t *testing.T) {
	jsonData := `{
		"crypto_key_path": "testkey.pem",
		"database_dsn": "postgres://localhost/db",
		"trusted_subnet": "127.0.0.0/24",
		"file_storage_path": "storage.bk",
		"log_level": "debug",
		"root_address": "0.0.0.0:9090",
		"secret": "secretkey",
		"store_interval": 120000000000,
		"restore": true
	}`

	tmpFile, err := os.CreateTemp("", "config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_, _ = tmpFile.WriteString(jsonData)
	_ = tmpFile.Close()

	initial := &Builder{
		Config:          tmpFile.Name(),
		RootAddress:     "localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "default.bk",
		StoreInterval:   60 * time.Second,
		Restore:         false,
	}

	updated, _ := initial.LoadFromFile().(*Builder)

	if updated.RootAddress != "0.0.0.0:9090" {
		t.Errorf("expected RootAddress to be updated, got %s", updated.RootAddress)
	}
	if updated.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be updated, got %s", updated.LogLevel)
	}
	if updated.StoreInterval != 120*time.Second {
		t.Errorf("expected StoreInterval to be 120s, got %v", updated.StoreInterval)
	}
	if updated.Restore != true {
		t.Errorf("expected Restore to be true, got %v", updated.Restore)
	}
	if updated.Secret != "secretkey" {
		t.Errorf("expected Secret to be updated, got %s", updated.Secret)
	}
	if updated.TrustedSubnet != "127.0.0.0/24" {
		t.Errorf("expected TrustedSubnet to be updated, got %s", updated.TrustedSubnet)
	}
}

func TestBuilder_LoadFromFile_FileDoesNotExist(t *testing.T) {
	b := &Builder{
		Config:        "/non/existent/path.json",
		RootAddress:   "localhost:8080",
		StoreInterval: 30 * time.Second,
	}

	result, _ := b.LoadFromFile().(*Builder)

	if result.RootAddress != "localhost:8080" {
		t.Errorf("expected RootAddress unchanged, got %s", result.RootAddress)
	}
	if result.StoreInterval != 30*time.Second {
		t.Errorf("expected StoreInterval unchanged, got %v", result.StoreInterval)
	}
}

func TestBuilder_LoadFromFile_PartialOverride(t *testing.T) {
	jsonData := `{
		"log_level": "error"
	}`

	tmpFile, err := os.CreateTemp("", "partial_config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_, _ = tmpFile.WriteString(jsonData)
	_ = tmpFile.Close()

	initial := &Builder{
		Config:          tmpFile.Name(),
		LogLevel:        "info",
		RootAddress:     "localhost:8080",
		StoreInterval:   45 * time.Second,
		FileStoragePath: "default.bk",
	}

	updated, _ := initial.LoadFromFile().(*Builder)

	if updated.LogLevel != "error" {
		t.Errorf("expected LogLevel to be overridden to 'error', got %s", updated.LogLevel)
	}
	if updated.RootAddress != "localhost:8080" {
		t.Errorf("expected RootAddress to stay the same, got %s", updated.RootAddress)
	}
	if updated.StoreInterval != 45*time.Second {
		t.Errorf("expected StoreInterval to stay the same, got %v", updated.StoreInterval)
	}
}
