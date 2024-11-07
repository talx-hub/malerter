package api

import (
	"encoding/json"
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
	switch err.(type) {
	case *customerror.NotFoundError:
		return http.StatusNotFound
	case *customerror.InvalidArgumentError:
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

	var metric repo.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Store(metric); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	metric, err := h.service.Get(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := "only POST requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	rawMetric := r.URL.Path
	metric, err := repo.FromURL(rawMetric)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	err = h.service.Store(metric)
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

	rawMetric := r.URL.Path
	metric, err := repo.FromURL(rawMetric)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	m, err := h.service.Get(metric)
	if err != nil {
		switch err.(type) {
		case *customerror.NotFoundError:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *customerror.InvalidArgumentError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
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
