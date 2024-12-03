package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/constants"

	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service"
)

const (
	errMsgPattern = `%s fails: %s`
)

type HTTPHandler struct {
	service service.Service
}

func NewHTTPHandler(s service.Service) *HTTPHandler {
	return &HTTPHandler{service: s}
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
	metric, err := model.NewMetric().FromJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	if metric.IsEmpty() {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, "metric value is empty"),
			http.StatusNotFound)
		return
	}
	if err = h.service.Add(metric); err != nil {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			http.StatusNotFound)
		return
	}

	dummyKey := metric.Type.String() + metric.Name
	metric, err = h.service.Find(dummyKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	metric, err := model.NewMetric().FromURL(r.URL.Path)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}
	if metric.IsEmpty() {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, "metric value is empty"),
			http.StatusNotFound)
		return
	}

	err = h.service.Add(metric)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metric, err := model.NewMetric().FromURL(r.URL.Path)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	dummyKey := metric.Type.String() + metric.Name
	m, err := h.service.Find(dummyKey)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeText)
	w.WriteHeader(http.StatusOK)
	valueStr := fmt.Sprintf("%v", m.ActualValue())
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		log.Fatal(err)
	}
}

func (h *HTTPHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric, err := model.NewMetric().FromJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	dummyKey := metric.Type.String() + metric.Name
	metric, err = h.service.Find(dummyKey)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.KeyContentType, constants.ContentTypeHTML)
	metrics := h.service.Get()
	page := createMetricsPage(metrics)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(page))
	if err != nil {
		log.Fatal(err)
	}
}

func (h *HTTPHandler) Ping(w http.ResponseWriter, _ *http.Request) {
	if h.service == nil {
		err := errors.New("the dumping service is not initialised")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.service.Ping(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func createMetricsPage(metrics []model.Metric) string {
	var page = `<html>
	<body>
%s	</body>
</html>`

	var data string
	for _, m := range metrics {
		data += fmt.Sprintf("\t\t<p>%s</p>\n", m.String())
	}
	return fmt.Sprintf(page, data)
}
