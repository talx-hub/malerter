package server

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
	AddressDefault       = "localhost:8080"
	RestoreDefault       = true
	StoreIntervalDefault = 300
)

const (
	EnvAddress         = "ADDRESS"
	EnvCryptoKeyPath   = "CRYPTO_KEY"
	EnvDatabaseDSN     = "DATABASE_DSN"
	EnvFileStoragePath = "FILE_STORAGE_PATH"
	EnvLogLevel        = "LOG_LEVEL"
	EnvRestore         = "RESTORE"
	EnvSecretKey       = "KEY"
	EnvStoreInterval   = "STORE_INTERVAL"
)

func FileStorageDefault() string {
	return time.Now().UTC().Format("2006-01-02_15:04:05_MST") + ".bk"
}

func NewDirector() *config.Director {
	return &config.Director{
		Builder: &Builder{},
	}
}

type Builder struct {
	CryptoKeyPath   string
	DatabaseDSN     string
	FileStoragePath string
	LogLevel        string
	RootAddress     string
	Secret          string
	StoreInterval   time.Duration
	Restore         bool
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.CryptoKeyPath, "crypto-key", constants.EmptyPath, "absolute path to public crypto key")
	flag.StringVar(&b.RootAddress, "a", AddressDefault, "server root address")
	flag.StringVar(&b.LogLevel, "l", constants.LogLevelDefault, "server log level")
	flag.StringVar(&b.FileStoragePath, "f", FileStorageDefault(), "backup file path")
	var backupInterval int64
	flag.Int64Var(&backupInterval, "i", StoreIntervalDefault, "interval in seconds of repository backup")
	flag.BoolVar(&b.Restore, "r", RestoreDefault, "restore backup while start")
	flag.StringVar(&b.DatabaseDSN, "d", "", "database source name")
	flag.StringVar(&b.Secret, "k", constants.NoSecret, "secret key")
	flag.Parse()

	b.StoreInterval = time.Duration(backupInterval) * time.Second
	return b
}

func (b *Builder) LoadFromEnv() config.Builder {
	if cryptoKeyPath, found := os.LookupEnv(EnvCryptoKeyPath); found {
		b.CryptoKeyPath = cryptoKeyPath
	}
	if a, found := os.LookupEnv(EnvAddress); found {
		b.RootAddress = a
	}
	if l, found := os.LookupEnv(EnvLogLevel); found {
		b.LogLevel = l
	}
	if f, found := os.LookupEnv(EnvFileStoragePath); found {
		b.FileStoragePath = f
	}
	if i, found := os.LookupEnv(EnvStoreInterval); found {
		backupInterval, err := strconv.Atoi(i)
		if err != nil {
			log.Fatal(err)
		}
		b.StoreInterval = time.Duration(backupInterval) * time.Second
	}
	if r, found := os.LookupEnv(EnvRestore); found {
		var err error
		b.Restore, err = strconv.ParseBool(r)
		if err != nil {
			log.Fatal(err)
		}
	}
	if d, found := os.LookupEnv(EnvDatabaseDSN); found {
		b.DatabaseDSN = d
	}
	if k, found := os.LookupEnv(EnvSecretKey); found {
		b.Secret = k
	}
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	if b.StoreInterval < 0 {
		return nil, errors.New("store interval must be positive")
	}
	return b, nil
}

func (b *Builder) Build() config.Config {
	return *b
}
