package main

import (
	"github.com/alant1t/metricscoll/internal/api"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"net/http"
)

func main() {
	rep := repo.MemRepository{}
	serv := service.NewMetricsDumper(&rep)
	handler := api.NewHTTPHandler(serv)

	var updateHandler http.Handler = http.HandlerFunc(handler.DumpMetric)

	mux := http.NewServeMux()
	mux.Handle("/update/", updateHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
