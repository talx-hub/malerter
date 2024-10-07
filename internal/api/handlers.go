package api

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/service"
	"net/http"
)

// TODO: Зачем мы создаем свою структуру хэндлера?
//   - Почему просто не сделать пакет с экспортируемыми
//   - функциями?
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
	if err := h.service.DumpMetric(rawMetric); err != nil {
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
	m, err := h.service.GetMetric(rawMetric)
	if err != nil {
		return
	}

	w.Write([]byte(m.Value))
	w.WriteHeader(http.StatusOK)
}
