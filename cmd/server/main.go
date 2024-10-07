package main

import (
	"github.com/alant1t/metricscoll/internal/api"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"net/http"
)

func main() {
	rep := repo.NewMemRepository()
	serv := service.NewMetricsDumper(rep)
	handler := api.NewHTTPHandler(serv)

	var updateHandler http.Handler = http.HandlerFunc(handler.DumpMetric)
	var getHandler http.Handler = http.HandlerFunc(handler.GetMetric)

	mux := http.NewServeMux()
	mux.Handle("/update/", updateHandler)
	mux.Handle("value", getHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
