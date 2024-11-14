package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/talx-hub/malerter/internal/api"
	"github.com/talx-hub/malerter/internal/backup"
	"github.com/talx-hub/malerter/internal/compressor"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service/server"
)

func main() {
	// TODO: тут какие-то кошмары с указателями(см. config/server/builder/.Build())... разобраться
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	if !ok {
		log.Fatal("unable to load server serverCfg")
	}

	rep := repo.NewMemRepository()
	bk, err := backup.New(cfg, rep)
	if err != nil {
		log.Fatalf("unable to load backup: %v", err)
	}
	if cfg.Restore {
		bk.Restore()
	}
	dumper := server.NewMetricsDumper(rep)
	handler := api.NewHTTPHandler(dumper)

	zeroLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	var updateHandler = zeroLogger.WrapHandler(bk.Middleware(handler.DumpMetric))
	var updateJSONHandler = zeroLogger.WrapHandler(compressor.GzipMiddleware(bk.Middleware(handler.DumpMetricJSON)))
	var getHandler = zeroLogger.WrapHandler(handler.GetMetric)
	var getAllHandler = zeroLogger.WrapHandler(compressor.GzipMiddleware(handler.GetAll))
	var getJSONHandler = zeroLogger.WrapHandler(compressor.GzipMiddleware(handler.GetMetricJSON))

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", getAllHandler)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", getJSONHandler)
			r.Get("/{type}/{name}", getHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", updateJSONHandler)
			r.Post("/{type}/{name}/{val}", updateHandler)
		})
	})

	zeroLogger.Logger.Info().
		Str(`"address"`, cfg.RootAddress).
		Dur(`"backup interval"`, cfg.StoreInterval).
		Bool(`"restore backup"`, cfg.Restore).
		Str(`"backup path"`, cfg.FileStoragePath).
		Msg("Starting server")
	srv := http.Server{Addr: cfg.RootAddress, Handler: router}
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err = srv.Shutdown(context.Background()); err != nil {
			zeroLogger.Logger.Fatal().
				Err(err).
				Msg("HTTP server Shutdown")
		}

		bk.Backup()
		if err = bk.Close(); err != nil {
			zeroLogger.Logger.Fatal().
				Err(err).
				Msg("Backup Close")
		}
		close(idleConnsClosed)
	}()

	if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		zeroLogger.Logger.Fatal().
			Err(err)
	}
	<-idleConnsClosed
}
