package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/api/middlewares"
	"github.com/talx-hub/malerter/internal/backup"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server"
)

func main() {
	// TODO: тут какие-то кошмары с указателями
	// (см. config/server/builder/.Build())... разобраться
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server serverCfg")
	}

	zeroLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %v", err)
	}

	var storage server.Storage
	database, err := metricDB(context.Background(), cfg.DatabaseDSN, zeroLogger)
	if err != nil {
		zeroLogger.Warn().Err(err).Msg("store metrics in memory")
		storage = memory.New(zeroLogger)
	} else {
		defer func() {
			if err = database.Close(); err != nil {
				zeroLogger.Error().Err(err).Msg("unable to close DB")
			}
		}()
		storage = database
	}

	ctxBackup, cancelBackup := context.WithCancel(context.Background())
	defer cancelBackup()
	if bk := backup.New(&cfg, storage, zeroLogger); bk != nil {
		bk.Run(ctxBackup)
	}

	zeroLogger.Info().
		Str(`"address"`, cfg.RootAddress).
		Dur(`"backup interval"`, cfg.StoreInterval).
		Bool(`"restore backup"`, cfg.Restore).
		Str(`"backup path"`, cfg.FileStoragePath).
		Str(`dsn`, cfg.DatabaseDSN).
		Msg("Starting server")
	srv := http.Server{
		Addr:    cfg.RootAddress,
		Handler: metricRouter(storage, zeroLogger),
	}
	idleConnectionsClosed := make(chan struct{})
	go idleShutdown(&srv, idleConnectionsClosed, zeroLogger, cancelBackup)

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		zeroLogger.Fatal().
			Err(err).Msg("error during HTTP server ListenAndServe")
	}
	<-idleConnectionsClosed
}

func metricRouter(
	repo server.Storage,
	loggr *logger.ZeroLogger,
) chi.Router {
	dumper := server.NewMetricsDumper(repo)
	handler := handlers.NewHTTPHandler(dumper, loggr)

	var getHandler = handler.GetMetric
	var getAllHandler = handler.GetAll
	var getJSONHandler = handler.GetMetricJSON
	var pingDBHandler = handler.Ping
	var updateHandler = handler.DumpMetric
	var updateJSONHandler = handler.DumpMetricJSON
	var batchHandler = handler.DumpMetricList

	router := chi.NewRouter()
	router.Use(
		middlewares.Logging(loggr),
		middlewares.Gzip(loggr),
	)

	router.Route("/", func(r chi.Router) {
		r.Get("/", getAllHandler)
		r.Route("/value", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AllowContentType(constants.ContentTypeJSON))
				r.Post("/", getJSONHandler)
			})
			r.Get("/{type}/{name}", getHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AllowContentType(constants.ContentTypeJSON))
				r.Post("/", updateJSONHandler)
			})
			r.Post("/", updateJSONHandler)
			r.Post("/{type}/{name}/{val}", updateHandler)
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", pingDBHandler)
		})
		r.Route("/updates", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AllowContentType(constants.ContentTypeJSON))
				r.Post("/", batchHandler)
			})
		})
	})

	return router
}

func metricDB(
	ctx context.Context,
	dsn string,
	loggr *logger.ZeroLogger,
) (*db.DB, error) {
	if len(dsn) == 0 {
		return nil, errors.New("DB DSN is empty")
	}

	database, err := db.New(ctx, dsn, loggr)
	if err != nil {
		return nil, fmt.Errorf("unable to create DB instance: %w", err)
	}
	return database, nil
}

func idleShutdown(
	s *http.Server,
	channel chan struct{},
	loggr *logger.ZeroLogger,
	cancelBackup context.CancelFunc,
) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	loggr.Info().Msg("shutdown signal received. Exiting...")
	ctxServer, cancelSrv := context.WithTimeout(
		context.Background(), constants.Timeout*time.Second)
	defer cancelSrv()
	if err := s.Shutdown(ctxServer); err != nil {
		loggr.Error().Err(err).Msg("error during HTTP server Shutdown")
	}

	cancelBackup()

	close(channel)
}
