package main

import (
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/talx-hub/malerter/internal/api"
	"github.com/talx-hub/malerter/internal/config"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service"
)

func main() {
	director := config.NewServerDirector()
	cfg, ok := director.Build().(config.Server)
	if !ok {
		log.Fatal("unable to load server config")
	}

	rep := repo.NewMemRepository()
	serv := service.NewMetricsDumper(rep)
	handler := api.NewHTTPHandler(serv)

	router := chi.NewRouter()

	lggr := logger.New()
	var updateHandler = lggr.WrapHandler(handler.DumpMetric)
	var getHandler = lggr.WrapHandler(handler.GetMetric)
	var getAllHandler = lggr.WrapHandler(handler.GetAll)

	router.Route("/", func(r chi.Router) {
		r.Get("/", getAllHandler)
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", getHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{val}", updateHandler)
		})
	})

	err := http.ListenAndServe(cfg.RootAddress, router)
	if err != nil {
		panic(err)
	}
}
