package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talx-hub/malerter/internal/api"
	"github.com/talx-hub/malerter/internal/config"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service"
)

func main() {
	conf := config.LoadServerConfig()

	rep := repo.NewMemRepository()
	serv := service.NewMetricsDumper(rep)
	handler := api.NewHTTPHandler(serv)

	router := chi.NewRouter()

	var updateHandler = http.HandlerFunc(handler.DumpMetric)
	var getHandler = http.HandlerFunc(handler.GetMetric)
	var getAllHandler = http.HandlerFunc(handler.GetAll)

	router.Get("/", getAllHandler)
	router.Get("/value/{type}/{name}", getHandler)
	router.Post("/update/{type}/{name}/{val}", updateHandler)

	err := http.ListenAndServe(conf.RootAddress, router)
	if err != nil {
		panic(err)
	}
}
