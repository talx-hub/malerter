package api

import (
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

	valueStr := fmt.Sprintf("%v", m.Value)
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e := "only GET requests are allowed"
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	metrics := h.service.GetAll()
	page := createMetricsPage(metrics)
	_, err := w.Write([]byte(page))
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
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
