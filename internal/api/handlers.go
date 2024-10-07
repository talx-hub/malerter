package api

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"log"
	"net/http"
	"strconv"
)

type HTTPHandler struct {
	service service.Service
}

func NewHTTPHandler(service service.Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := "only POST requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	rawMetric := r.URL.Path
	if err := h.service.Store(rawMetric); err != nil {
		switch err.(type) {
		case *customerror.NotFoundError:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *customerror.InvalidArgumentError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
	m, err := h.service.Get(rawMetric)
	if err != nil {
		switch err.(type) {
		case *customerror.NotFoundError:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *customerror.InvalidArgumentError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	var valueStr string
	if m.Type == repo.MetricTypeCounter {
		valueStr = strconv.FormatInt(m.Value.(int64), 10)
	} else {
		valueStr = strconv.FormatFloat(
			m.Value.(float64), 'f', 2, 64)
	}

	_, err = w.Write([]byte(valueStr))
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
}
