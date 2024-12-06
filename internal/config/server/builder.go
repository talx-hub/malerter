package server

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/talx-hub/malerter/internal/config"
)

const (
	AddressDefault       = "localhost:8080"
	LogLevelDefault      = "Info"
	StoreIntervalDefault = 300
	RestoreDefault       = true
)

const (
	EnvAddress         = "ADDRESS"
	EnvLogLevel        = "LOG_LEVEL"
	EnvStoreInterval   = "STORE_INTERVAL"
	EnvFileStoragePath = "FILE_STORAGE_PATH"
	EnvRestore         = "RESTORE"
	EnvDatabaseDSN     = "DATABASE_DSN"
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
	DatabaseDSN     string
	FileStoragePath string
	LogLevel        string
	RootAddress     string
	StoreInterval   time.Duration
	Restore         bool
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.RootAddress, "a", AddressDefault, "server root address")
	flag.StringVar(&b.LogLevel, "l", LogLevelDefault, "server log level")
	flag.StringVar(&b.FileStoragePath, "f", FileStorageDefault(), "backup file path")
	var backupInterval int64
	flag.Int64Var(&backupInterval, "i", StoreIntervalDefault, "interval in seconds of repository backup")
	flag.BoolVar(&b.Restore, "r", RestoreDefault, "restore backup while start")
	flag.StringVar(&b.DatabaseDSN, "d", "", "database source name")
	flag.Parse()

	b.StoreInterval = time.Duration(backupInterval) * time.Second
	return b
}

func (b *Builder) LoadFromEnv() config.Builder {
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
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	if b.StoreInterval < 0 {
		return nil, errors.New("store interval must be positive")
	}
	return b, nil
}

func (b *Builder) Build() config.Config {
	// TODO: wtf?? почему вынужден возвращать значение?
	return *b
}
