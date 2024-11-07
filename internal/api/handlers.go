package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service"
)

type HTTPHandler struct {
	service service.Service
}

func NewHTTPHandler(service service.Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func getStatusFromError(err error) int {
	var notFoundError *customerror.NotFoundError
	var invalidArgumentError *customerror.InvalidArgumentError
	switch {
	case errors.As(err, &notFoundError):
		return http.StatusNotFound
	case errors.As(err, &invalidArgumentError):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *HTTPHandler) DumpMetricJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := "only POST requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	metric, err := repo.NewMetric().FromJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}
	if metric == nil {
		http.Error(w, "metric is nil", http.StatusInternalServerError)
		return
	}
	if metric.IsEmpty() {
		e := customerror.NotFoundError{
			MetricURL: metric.ToURL(),
			Info:      "metric value is empty",
		}
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	if err = h.service.Store(*metric); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	*metric, err = h.service.Get(*metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := "only POST requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	metric, err := repo.NewMetric().FromURL(r.URL.Path)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}
	if metric == nil {
		http.Error(w, "metric is nil", http.StatusInternalServerError)
		return
	}
	if metric.IsEmpty() {
		e := customerror.NotFoundError{
			MetricURL: metric.ToURL(),
			Info:      "metric value is empty",
		}
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}

	err = h.service.Store(*metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e := "only GET requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	metric, err := repo.NewMetric().FromURL(r.URL.Path)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}
	if metric == nil {
		http.Error(w, "metric is nil", http.StatusInternalServerError)
		return
	}

	m, err := h.service.Get(*metric)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	w.Header().Set("content-type", "text/plain")
	valueStr := fmt.Sprintf("%v", m.ActualValue())
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		log.Fatal(err)
	}
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e := "only GET requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "text/html")
	metrics := h.service.GetAll()
	page := createMetricsPage(metrics)
	_, err := w.Write([]byte(page))
	if err != nil {
		log.Fatal(err)
	}
}

func createMetricsPage(metrics []repo.Metric) string {
	var page = `<html>
	<body>
%s	</body>
</html>`

	var data string
	for _, m := range metrics {
		data += fmt.Sprintf("\t\t<p>%s</p>\n", m.String())
	}
	fmt.Printf(page, data)
	return fmt.Sprintf(page, data)
}
