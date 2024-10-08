package main

import (
	"github.com/alant1t/metricscoll/internal/api"
	"github.com/alant1t/metricscoll/internal/config"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	conf, err := config.LoadServerConfig()
	if err != nil {
		log.Fatal(err)
	}

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

	err = http.ListenAndServe(conf.RootAddress, router)
	if err != nil {
		panic(err)
	}
}
