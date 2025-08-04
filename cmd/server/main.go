package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/talx-hub/malerter/internal/api/handlers"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	l "github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server"
	"github.com/talx-hub/malerter/internal/service/server/backup"
	"github.com/talx-hub/malerter/internal/service/server/buildinfo"
	"github.com/talx-hub/malerter/pkg/queue"
	"github.com/talx-hub/malerter/pkg/shutdown"
)

func main() {
	cfg := loadConfig()
	logger := initLogger(&cfg)

	buffer := queue.New[model.Metric]()
	defer buffer.Close()

	storage := initStorage(&cfg, logger, &buffer)
	defer closeDatabase(storage)

	ctxBackup, cancelBackup := context.WithCancel(context.Background())
	defer cancelBackup()
	startBackupService(ctxBackup, &cfg, &buffer, storage, logger)

	printStartupInfo(&cfg, logger)

	srv := server.Init(&cfg, storage, logger)
	if srv == nil {
		logger.Fatal().Msg("Unable to start server. Exit")
		return
	}

	idleConnectionsClosed := make(chan struct{})
	go shutdown.IdleShutdown(
		idleConnectionsClosed,
		logger,
		func(args ...any) error {
			return shutdownServer(srv, cancelBackup)
		},
	)

	if err := srv.Start(); err != nil {
		logger.Fatal().Err(err).Msg("error during server start")
	}
	<-idleConnectionsClosed
}

func loadConfig() serverCfg.Builder {
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server config")
	}
	return cfg
}

func initLogger(cfg *serverCfg.Builder) *l.ZeroLogger {
	logger, err := l.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %v", err)
	}
	return logger
}

func initStorage(
	cfg *serverCfg.Builder,
	logger *l.ZeroLogger,
	buffer *queue.Queue[model.Metric],
) handlers.Storage {
	dbStorage, err := metricDB(context.Background(), cfg.DatabaseDSN, logger, buffer)
	if err != nil {
		logger.Warn().Err(err).Msg("store metrics in memory")
		return memory.New(logger, buffer)
	}
	return dbStorage
}

func closeDatabase(storage handlers.Storage) {
	if dbStorage, ok := storage.(*db.DB); ok {
		dbStorage.Close()
	}
}

func startBackupService(
	ctx context.Context,
	cfg *serverCfg.Builder,
	buffer *queue.Queue[model.Metric],
	storage handlers.Storage,
	logger *l.ZeroLogger,
) {
	bk := backup.New(cfg, buffer, storage, logger)
	if bk != nil {
		go bk.Run(ctx)
	} else {
		logger.Warn().Msg("unable to load backup service")
		buffer.Close()
	}
}

func printStartupInfo(cfg *serverCfg.Builder, logger *l.ZeroLogger) {
	logger.Info().
		Str("address", cfg.RootAddress).
		Str("trusted subnet", cfg.TrustedSubnet).
		Dur("backup interval", cfg.StoreInterval).
		Bool("restore backup", cfg.Restore).
		Str("backup path", cfg.FileStoragePath).
		Bool("signature check", cfg.Secret != constants.NoSecret).
		Str("dsn", cfg.DatabaseDSN).
		Str("buildVersion", buildinfo.Version).
		Str("buildCommit", buildinfo.Commit).
		Str("buildDate", buildinfo.Date).
		Msg("Starting server")
}

func metricDB(
	ctx context.Context,
	dsn string,
	logger *l.ZeroLogger,
	buffer *queue.Queue[model.Metric],
) (*db.DB, error) {
	if len(dsn) == 0 {
		return nil, errors.New("DB DSN is empty")
	}

	database, err := db.New(ctx, dsn, logger, buffer)
	if err != nil {
		return nil, fmt.Errorf("unable to create DB instance: %w", err)
	}
	return database, nil
}

func shutdownServer(s server.Server, cancelBackup context.CancelFunc) error {
	ctxTO, cancelSrv := context.WithTimeout(
		context.Background(), constants.TimeoutShutdown)
	defer cancelSrv()

	if err := s.Stop(ctxTO); err != nil {
		cancelBackup()
		return fmt.Errorf("server stop failed: %w", err)
	}
	cancelBackup()
	return nil
}
