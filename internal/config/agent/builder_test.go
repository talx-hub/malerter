package agent

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_LoadFromEnv(t *testing.T) {
	_ = os.Setenv(EnvCryptoKeyPath, "/keys/public.pem")
	_ = os.Setenv(EnvHost, "127.0.0.1:9000")
	_ = os.Setenv(EnvSecretKey, "my-secret")
	_ = os.Setenv(EnvPollInterval, "5")
	_ = os.Setenv(EnvRateLimit, "10")
	_ = os.Setenv(EnvReportInterval, "15")

	defer func() {
		_ = os.Unsetenv(EnvCryptoKeyPath)
		_ = os.Unsetenv(EnvHost)
		_ = os.Unsetenv(EnvSecretKey)
		_ = os.Unsetenv(EnvPollInterval)
		_ = os.Unsetenv(EnvRateLimit)
		_ = os.Unsetenv(EnvReportInterval)
	}()

	b := &Builder{}
	b.LoadFromEnv()

	assert.Equal(t, "/keys/public.pem", b.CryptoKeyPath)
	assert.Equal(t, "127.0.0.1:9000", b.ServerAddress)
	assert.Equal(t, "my-secret", b.Secret)
	assert.Equal(t, 10, b.RateLimit)
	assert.Equal(t, 5*time.Second, b.PollInterval)
	assert.Equal(t, 15*time.Second, b.ReportInterval)
}

func TestBuilder_IsValid_Positive(t *testing.T) {
	b := &Builder{
		PollInterval:   2 * time.Second,
		ReportInterval: 5 * time.Second,
	}
	_, err := b.IsValid()
	assert.NoError(t, err)
}

func TestBuilder_IsValid_Negative(t *testing.T) {
	tests := []struct {
		name    string
		builder Builder
		wantErr string
	}{
		{
			name:    "Negative ReportInterval",
			builder: Builder{ReportInterval: -1, PollInterval: 1},
			wantErr: "report interval must be positive",
		},
		{
			name:    "Negative PollInterval",
			builder: Builder{ReportInterval: 1, PollInterval: -1},
			wantErr: "poll interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.IsValid()
			assert.Error(t, err)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	b := &Builder{
		CryptoKeyPath:  "/keys/public.pem",
		LogLevel:       "info",
		Secret:         "key",
		ServerAddress:  "localhost:9000",
		RateLimit:      5,
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
	}
	cfg, ok := b.Build().(Builder)
	require.True(t, ok)

	assert.Equal(t, b.CryptoKeyPath, cfg.CryptoKeyPath)
	assert.Equal(t, b.LogLevel, cfg.LogLevel)
	assert.Equal(t, b.Secret, cfg.Secret)
	assert.Equal(t, b.ServerAddress, cfg.ServerAddress)
	assert.Equal(t, b.RateLimit, cfg.RateLimit)
	assert.Equal(t, b.ReportInterval, cfg.ReportInterval)
	assert.Equal(t, b.PollInterval, cfg.PollInterval)
}

func TestBuilder_LoadFromFile_Success(t *testing.T) {
	jsonData := `{
		"crypto_key_path": "key.pem",
		"log_level": "debug",
		"secret": "supersecret",
		"server_address": "192.168.0.1:9090",
		"rate_limit": 10,
		"report_interval": 15000000000,
		"poll_interval": 5000000000
	}`

	tmpFile, err := os.CreateTemp("", "agent_config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_, _ = tmpFile.WriteString(jsonData)
	_ = tmpFile.Close()

	initial := &Builder{
		Config:         tmpFile.Name(),
		ServerAddress:  "localhost:8080",
		RateLimit:      5,
		PollInterval:   2 * time.Second,
		ReportInterval: 5 * time.Second,
		Secret:         "",
	}

	result, _ := initial.LoadFromFile().(*Builder)

	if result.ServerAddress != "192.168.0.1:9090" {
		t.Errorf("expected ServerAddress to be overridden, got %s", result.ServerAddress)
	}
	if result.RateLimit != 10 {
		t.Errorf("expected RateLimit to be 10, got %d", result.RateLimit)
	}
	if result.PollInterval != 5*time.Second {
		t.Errorf("expected PollInterval to be 5s, got %v", result.PollInterval)
	}
	if result.ReportInterval != 15*time.Second {
		t.Errorf("expected ReportInterval to be 15s, got %v", result.ReportInterval)
	}
	if result.Secret != "supersecret" {
		t.Errorf("expected Secret to be 'supersecret', got %s", result.Secret)
	}
}

func TestBuilder_LoadFromFile_FileNotExists(t *testing.T) {
	b := &Builder{
		Config:         "/nonexistent/path.json",
		ServerAddress:  "localhost:8080",
		RateLimit:      3,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	result, _ := b.LoadFromFile().(*Builder)

	if result.ServerAddress != "localhost:8080" {
		t.Errorf("expected ServerAddress to stay the same, got %s", result.ServerAddress)
	}
	if result.RateLimit != 3 {
		t.Errorf("expected RateLimit to stay the same, got %d", result.RateLimit)
	}
}

func TestBuilder_LoadFromFile_PartialOverride(t *testing.T) {
	jsonData := `{
		"rate_limit": 20
	}`

	tmpFile, err := os.CreateTemp("", "agent_partial_config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_, _ = tmpFile.WriteString(jsonData)
	_ = tmpFile.Close()

	initial := &Builder{
		Config:         tmpFile.Name(),
		ServerAddress:  "localhost:8080",
		RateLimit:      3,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	result, _ := initial.LoadFromFile().(*Builder)

	if result.RateLimit != 20 {
		t.Errorf("expected RateLimit to be updated to 20, got %d", result.RateLimit)
	}
	if result.ServerAddress != "localhost:8080" {
		t.Errorf("expected ServerAddress to remain unchanged, got %s", result.ServerAddress)
	}
}
