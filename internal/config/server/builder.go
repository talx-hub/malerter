package server

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/talx-hub/malerter/internal/config"
)

const (
	AddressDefault       = "localhost:8080"
	LogLevelDefault      = "InfoLevel"
	StoreIntervalDefault = 300
	RestoreDefault       = true
)

const (
	EnvAddress         = "ADDRESS"
	EnvLogLevel        = "LOG_LEVEL"
	EnvStoreInterval   = "STORE_INTERVAL"
	EnvFileStoragePath = "FILE_STORAGE_PATH"
	EnvRestore         = "RESTORE"
)

func FileStorageDefault() string {
	return time.Now().UTC().String() + ".bk"
}

func NewDirector() *config.Director {
	return &config.Director{
		Builder: &Builder{},
	}
}

type Builder struct {
	RootAddress     string
	LogLevel        string
	FileStoragePath string
	StoreInterval   time.Duration
	Restore         bool
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.RootAddress, "a", AddressDefault, "server root address")
	flag.StringVar(&b.LogLevel, "l", LogLevelDefault, "server log level")
	flag.StringVar(&b.FileStoragePath, "f", FileStorageDefault(), "backup file path")
	var backupInterval int64
	flag.Int64Var(&backupInterval, "i", StoreIntervalDefault, "interval in seconds of repository backup")
	b.Restore = *flag.Bool("r", RestoreDefault, "restore backup while start")
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
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	if b.StoreInterval < 0 {
		return nil, fmt.Errorf("store interval must be positive")
	}
	return b, nil
}

func (b *Builder) Build() config.Config {
	// TODO: wtf?? почему вынужден возвращать значение?
	return *b
}
