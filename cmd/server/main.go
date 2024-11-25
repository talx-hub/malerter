package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/talx-hub/malerter/internal/api"
	"github.com/talx-hub/malerter/internal/backup"
	"github.com/talx-hub/malerter/internal/compressor/gzip"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger/zerologger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server"
)

func main() {
	// TODO: тут какие-то кошмары с указателями(см. config/server/builder/.Build())... разобраться
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server serverCfg")
	}

	zeroLogger, err := zerologger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %s", err.Error())
	}

	storage := memory.New()
	bk, err := backup.New(cfg, storage)
	if err != nil {
		zeroLogger.Fatal().
			Err(err).
			Msg("unable to load Backup service")
	}
	defer func() {
		if err = bk.Close(); err != nil {
			zeroLogger.Fatal().
				Err(err).
				Msg("unable to close Backup service")
		}
	}()
	if cfg.Restore {
		bk.Restore()
	}

	zeroLogger.Info().
		Str(`"address"`, cfg.RootAddress).
		Dur(`"backup interval"`, cfg.StoreInterval).
		Bool(`"restore backup"`, cfg.Restore).
		Str(`"backup path"`, cfg.FileStoragePath).
		Msg("Starting server")
	srv := http.Server{
		Addr:    cfg.RootAddress,
		Handler: metricRouter(storage, zeroLogger, bk),
	}
	idleConnectionsClosed := make(chan struct{})
	go idleShutdown(&srv, idleConnectionsClosed, zeroLogger, bk)

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		zeroLogger.Fatal().
			Err(err).
			Msg("error during HTTP server ListenAndServe")
	}
	<-idleConnectionsClosed
}

func metricRouter(repo *memory.Metrics, logger *zerologger.ZeroLogger, backer *backup.File) chi.Router {
	dumper := server.NewMetricsDumper(repo)
	handler := api.NewHTTPHandler(dumper)

	var updateHandler = backer.Middleware(handler.DumpMetric)
	var updateJSONHandler = backer.Middleware(handler.DumpMetricJSON)
	var getHandler = handler.GetMetric
	var getAllHandler = handler.GetAll
	var getJSONHandler = handler.GetMetricJSON

	router := chi.NewRouter()
	router.Use(
		logger.Middleware,
		gzip.Middleware,
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
	})

	return router
}

func idleShutdown(s *http.Server, channel chan struct{},
	logger *zerologger.ZeroLogger, backer *backup.File) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	logger.Info().
		Msg("shutdown signal received. Exiting...")
	if err := s.Shutdown(context.Background()); err != nil {
		logger.Fatal().
			Err(err).
			Msg("error during HTTP server Shutdown")
	}

	backer.Backup()

	close(channel)
}
