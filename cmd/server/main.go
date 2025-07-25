package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/api/handlers"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	l "github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server/backup"
	"github.com/talx-hub/malerter/internal/service/server/buildinfo"
	"github.com/talx-hub/malerter/internal/service/server/router"
	"github.com/talx-hub/malerter/pkg/queue"
	"github.com/talx-hub/malerter/pkg/shutdown"
)

func main() {
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server serverCfg")
	}

	logger, err := l.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %v", err)
	}

	var storage handlers.Storage
	buffer := queue.New[model.Metric]()
	defer buffer.Close()
	database, err := metricDB(
		context.Background(),
		cfg.DatabaseDSN,
		logger,
		&buffer)
	if err != nil {
		logger.Warn().Err(err).Msg("store metrics in memory")
		storage = memory.New(logger, &buffer)
	} else {
		defer database.Close()
		storage = database
	}

	ctxBackup, cancelBackup := context.WithCancel(context.Background())
	defer cancelBackup()
	if bk := backup.New(&cfg, &buffer, storage, logger); bk != nil {
		go bk.Run(ctxBackup)
	} else {
		logger.Warn().Msg("unable to load backup service")
		buffer.Close()
	}

	logger.Info().
		Str(`"address"`, cfg.RootAddress).
		Dur(`"backup interval"`, cfg.StoreInterval).
		Bool(`"restore backup"`, cfg.Restore).
		Str(`"backup path"`, cfg.FileStoragePath).
		Bool(`"signature check"'`, cfg.Secret != constants.NoSecret).
		Str(`dsn`, cfg.DatabaseDSN).
		Str("buildVersion", buildinfo.Version).
		Str("buildCommit", buildinfo.Commit).
		Str("buildDate", buildinfo.Date).
		Msg("Starting server")

	chiRouter := router.New(logger, cfg.Secret, cfg.CryptoKeyPath)
	chiRouter.SetRouter(handlers.NewHTTPHandler(storage, logger))

	srv := http.Server{
		Addr:    cfg.RootAddress,
		Handler: chiRouter.GetRouter(),
	}

	idleConnectionsClosed := make(chan struct{})
	go shutdown.IdleShutdown(
		idleConnectionsClosed,
		logger,
		func(args ...any) error {
			return shutdownServer(&srv, cancelBackup)
		},
	)

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal().
			Err(err).Msg("error during HTTP server ListenAndServe")
	}
	<-idleConnectionsClosed
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

func shutdownServer(s *http.Server, cancelBackup context.CancelFunc) error {
	ctxServer, cancelSrv := context.WithTimeout(
		context.Background(), constants.TimeoutShutdown)
	defer cancelSrv()

	if err := s.Shutdown(ctxServer); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	cancelBackup()
	return nil
}
