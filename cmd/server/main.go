package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talx-hub/malerter/internal/api"
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
	serv := server.NewMetricsDumper(rep)
	handler := api.NewHTTPHandler(serv)

	router := chi.NewRouter()

	lggr := logger.New()
	var updateHandler = lggr.WrapHandler(handler.DumpMetric)
	var updateJSONHandler = lggr.WrapHandler(compressor.GzipMiddleware(handler.DumpMetricJSON))
	var getHandler = lggr.WrapHandler(handler.GetMetric)
	var getAllHandler = lggr.WrapHandler(compressor.GzipMiddleware(handler.GetAll))
	var getJSONHandler = lggr.WrapHandler(compressor.GzipMiddleware(handler.GetMetricJSON))

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

	err := http.ListenAndServe(cfg.RootAddress, router)
	if err != nil {
		panic(err)
	}
}
