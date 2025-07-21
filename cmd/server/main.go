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
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server/backup"
	"github.com/talx-hub/malerter/internal/service/server/router"
	"github.com/talx-hub/malerter/pkg/queue"
	"github.com/talx-hub/malerter/pkg/shutdown"
)

func main() {
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server serverCfg")
	}

	zeroLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %v", err)
	}

	var storage handlers.Storage
	buffer := queue.New[model.Metric]()
	defer buffer.Close()
	database, err := metricDB(
		context.Background(),
		cfg.DatabaseDSN,
		zeroLogger,
		&buffer)
	if err != nil {
		zeroLogger.Warn().Err(err).Msg("store metrics in memory")
		storage = memory.New(zeroLogger, &buffer)
	} else {
		defer database.Close()
		storage = database
	}

	ctxBackup, cancelBackup := context.WithCancel(context.Background())
	defer cancelBackup()
	if bk := backup.New(&cfg, &buffer, storage, zeroLogger); bk != nil {
		go bk.Run(ctxBackup)
	} else {
		zeroLogger.Warn().Msg("unable to load backup service")
		buffer.Close()
	}

	zeroLogger.Info().
		Str(`"address"`, cfg.RootAddress).
		Dur(`"backup interval"`, cfg.StoreInterval).
		Bool(`"restore backup"`, cfg.Restore).
		Str(`"backup path"`, cfg.FileStoragePath).
		Bool(`"signature check"'`, cfg.Secret != constants.NoSecret).
		Str(`dsn`, cfg.DatabaseDSN).
		Msg("Starting server")

	chiRouter := router.New(zeroLogger, cfg.Secret)
	chiRouter.SetRouter(handlers.NewHTTPHandler(storage, zeroLogger))

	srv := http.Server{
		Addr:    cfg.RootAddress,
		Handler: chiRouter.GetRouter(),
	}

	idleConnectionsClosed := make(chan struct{})
	go shutdown.IdleShutdown(
		idleConnectionsClosed,
		zeroLogger,
		func(args ...any) error {
			return shutdownServer(&srv, cancelBackup)
		},
	)

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		zeroLogger.Fatal().
			Err(err).Msg("error during HTTP server ListenAndServe")
	}
	<-idleConnectionsClosed
}

func metricDB(
	ctx context.Context,
	dsn string,
	loggr *logger.ZeroLogger,
	buffer *queue.Queue[model.Metric],
) (*db.DB, error) {
	if len(dsn) == 0 {
		return nil, errors.New("DB DSN is empty")
	}

	database, err := db.New(ctx, dsn, loggr, buffer)
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
