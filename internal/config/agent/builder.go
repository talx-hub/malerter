package agent

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/talx-hub/malerter/internal/config"
	"github.com/talx-hub/malerter/internal/constants"
)

const (
	HostDefault           = "localhost:8080"
	PoolIntervalDefault   = 2
	RateLimitDefault      = 3
	ReportIntervalDefault = 10
)

const (
	EnvCryptoKeyPath  = "CRYPTO_KEY"
	EnvHost           = "ADDRESS"
	EnvSecretKey      = "KEY"
	EnvPollInterval   = "POLL_INTERVAL"
	EnvRateLimit      = "RATE_LIMIT"
	EnvReportInterval = "REPORT_INTERVAL"
)

func NewDirector() *config.Director {
	return &config.Director{
		Builder: &Builder{},
	}
}

type Builder struct {
	CryptoKeyPath  string
	LogLevel       string
	Secret         string
	ServerAddress  string
	RateLimit      int
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.CryptoKeyPath, "crypto-key", constants.EmptyPath, "absolute path to public crypto key")
	flag.StringVar(&b.LogLevel, "ll", constants.LogLevelDefault, "server log level")
	flag.StringVar(&b.ServerAddress, "a", HostDefault, "alert-host address")
	flag.StringVar(&b.Secret, "k", constants.NoSecret, "secret key")

	flag.IntVar(&b.RateLimit, "l", RateLimitDefault, "outgoing requests count")

	var pi int64
	flag.Int64Var(&pi, "p", PoolIntervalDefault, "interval in seconds of polling and collecting metrics")

	var ri int64
	flag.Int64Var(&ri, "r", ReportIntervalDefault, "interval in seconds of sending metrics to alert server")

	flag.Parse()

	b.ReportInterval = time.Duration(ri) * time.Second
	b.PollInterval = time.Duration(pi) * time.Second
	return b
}

func (b *Builder) LoadFromEnv() config.Builder {
	if cryptoKeyPath, found := os.LookupEnv(EnvCryptoKeyPath); found {
		b.CryptoKeyPath = cryptoKeyPath
	}
	if addr, found := os.LookupEnv(EnvHost); found {
		b.ServerAddress = addr
	}
	if rateLimitStr, found := os.LookupEnv(EnvRateLimit); found {
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			log.Fatal(err)
		}
		b.RateLimit = rateLimit
	}
	if pi, found := os.LookupEnv(EnvPollInterval); found {
		piInt, err := strconv.Atoi(pi)
		if err != nil {
			log.Fatal(err)
		}
		b.PollInterval = time.Duration(piInt) * time.Second
	}
	if ri, found := os.LookupEnv(EnvReportInterval); found {
		riInt, err := strconv.Atoi(ri)
		if err != nil {
			log.Fatal(err)
		}
		b.ReportInterval = time.Duration(riInt) * time.Second
	}
	if secret, found := os.LookupEnv(EnvSecretKey); found {
		b.Secret = secret
	}
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	if b.ReportInterval < 0 {
		return nil, errors.New("report interval must be positive")
	}
	if b.PollInterval < 0 {
		return nil, errors.New("poll interval must be positive")
	}
	return b, nil
}

func (b *Builder) Build() config.Config {
	return *b
}
