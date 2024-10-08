package main

import (
	"flag"
	"github.com/alant1t/metricscoll/internal/api"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var options struct {
	addr string
}

func init() {
	flag.StringVar(
		&options.addr,
		"a",
		"localhost:8080",
		"server root address")
}

func main() {
	flag.Parse()
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

	err := http.ListenAndServe(options.addr, router)
	if err != nil {
		panic(err)
	}
}
