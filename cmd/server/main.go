package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/api/middlewares"
	"github.com/talx-hub/malerter/internal/backup"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server"
	"github.com/talx-hub/malerter/internal/utils/queue"
	"github.com/talx-hub/malerter/internal/utils/shutdown"
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

	var storage server.Storage
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
	srv := http.Server{
		Addr:    cfg.RootAddress,
		Handler: metricRouter(storage, zeroLogger, cfg.Secret),
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

func metricRouter(
	repo server.Storage,
	loggr *logger.ZeroLogger,
	secret string,
) chi.Router {
	dumper := server.NewMetricsDumper(repo)
	handler := handlers.NewHTTPHandler(dumper, loggr)

	router := chi.NewRouter()
	router.Use(middlewares.Logging(loggr))

	router.Route("/", func(r chi.Router) {
		r.
			With(middlewares.WriteSignature(secret)).
			With(middlewares.Compress(loggr)).
			Get("/", handler.GetAll)
		r.Route("/value", func(r chi.Router) {
			r.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.WriteSignature(secret)).
				With(middlewares.Decompress(loggr)).
				With(middlewares.Compress(loggr)).
				Post("/", handler.GetMetricJSON)
			r.Get("/{type}/{name}", handler.GetMetric)
		})
		r.Route("/update", func(r chi.Router) {
			r.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.WriteSignature(secret)).
				With(middlewares.Decompress(loggr)).
				With(middlewares.Compress(loggr)).
				Post("/", handler.DumpMetricJSON)
			r.Post("/{type}/{name}/{val}", handler.DumpMetric)
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handler.Ping)
		})
		r.Route("/updates", func(r chi.Router) {
			r.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.CheckSignature(secret)).
				With(middlewares.Decompress(loggr)).
				With(middlewares.Compress(loggr)).
				Post("/", handler.DumpMetricList)
		})
		r.Route("/debug/pprof", func(r chi.Router) {
			r.HandleFunc("/", pprof.Index)
			r.HandleFunc("cmdline", pprof.Cmdline)
			r.HandleFunc("profile", pprof.Profile)
			r.HandleFunc("symbol", pprof.Symbol)
			r.HandleFunc("trace", pprof.Trace)

			for _, p := range []string{
				"allocs", "block", "goroutine", "heap", "mutex", "threadcreate",
			} {
				r.HandleFunc("/"+p, pprof.Handler(p).ServeHTTP)
			}
		})
	})

	return router
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
