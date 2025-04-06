package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/db"
	"github.com/talx-hub/malerter/internal/service"
)

const (
	errMsgPattern = `%s fails: %s`
)

type HTTPHandler struct {
	service service.Service
	log     *logger.ZeroLogger
}

func NewHTTPHandler(s service.Service, log *logger.ZeroLogger) *HTTPHandler {
	return &HTTPHandler{service: s, log: log}
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

func extractJSON(body io.Reader) (model.Metric, error) {
	m := model.NewMetric()
	if err := json.NewDecoder(body).Decode(m); err != nil {
		return model.Metric{},
			fmt.Errorf("unable to decode metric: %w", err)
	}
	if err := m.CheckValid(); err != nil {
		return model.Metric{},
			&customerror.InvalidArgumentError{
				Info: fmt.Sprintf("decoded metric is invalid: %v", err)}
	}

	return *m, nil
}

func (h *HTTPHandler) extractJSONs(body io.Reader) ([]model.Metric, error) {
	var metrics []model.Metric
	if err := json.NewDecoder(body).Decode(&metrics); err != nil {
		return nil,
			fmt.Errorf("unable to decode batch: %w", err)
	}

	validList := make([]model.Metric, 0)
	for _, m := range metrics {
		if err := m.CheckValid(); err != nil || m.IsEmpty() {
			h.log.Error().Err(err).Msg("decoded metric is invalid")
			continue
		}
		validList = append(validList, m)
	}
	return validList, nil
}

func (h *HTTPHandler) DumpMetricList(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.extractJSONs(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	wrappedBatch := func(args ...any) (any, error) {
		return nil, h.service.Batch(r.Context(), metrics)
	}
	if _, err = db.WithConnectionCheck(wrappedBatch); err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) DumpMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric, err := extractJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}
	if metric.IsEmpty() {
		http.Error(w, "failed to dump empty metric", http.StatusBadRequest)
		return
	}

	wrappedAdd := func(args ...any) (any, error) {
		return nil, h.service.Add(r.Context(), metric)
	}
	_, err = db.WithConnectionCheck(wrappedAdd)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			http.StatusNotFound)
		return
	}

	dummyKey := metric.Type.String() + " " + metric.Name
	wrappedFind := func(args ...any) (any, error) {
		return h.service.Find(r.Context(), dummyKey)
	}
	m, err := db.WithConnectionCheck(wrappedFind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	metric, ok := m.(model.Metric)
	if !ok {
		h.log.Error().Msg("failed to convert 'any' to model.Metric")
		http.Error(
			w,
			"failed to convert find result",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeJSON)
	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	mName := chi.URLParam(r, "name")
	mType := chi.URLParam(r, "type")
	mValue := chi.URLParam(r, "val")
	metric, err := model.NewMetric().FromValues(
		mName, model.MetricType(mType), mValue)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}
	if metric.IsEmpty() {
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, "metric value is empty"),
			http.StatusNotFound)
		return
	}

	wrappedAdd := func(args ...any) (any, error) {
		return nil, h.service.Add(r.Context(), metric)
	}
	_, err = db.WithConnectionCheck(wrappedAdd)
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
	mName := chi.URLParam(r, "name")
	mType := chi.URLParam(r, "type")
	metric, err := model.NewMetric().FromValues(
		mName, model.MetricType(mType), "0")
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}
	dummyKey := mType + " " + mName
	wrappedFind := func(args ...any) (any, error) {
		return h.service.Find(r.Context(), dummyKey)
	}
	m, err := db.WithConnectionCheck(wrappedFind)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}
	metric, ok := m.(model.Metric)
	if !ok {
		h.log.Error().Msg("failed to convert 'any' to model.Metric")
		http.Error(w, "failed to convert 'find' result", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeText)
	valueStr := fmt.Sprintf("%v", metric.ActualValue())
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		h.log.Error().Err(err).Msg("failed to write response")
	}
}

func (h *HTTPHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric, err := extractJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}

	dummyKey := metric.Type.String() + " " + metric.Name
	wrappedFind := func(args ...any) (any, error) {
		return h.service.Find(r.Context(), dummyKey)
	}
	m, err := db.WithConnectionCheck(wrappedFind)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}
	metric, ok := m.(model.Metric)
	if !ok {
		h.log.Error().Msg("failed to convert 'any' to model.Metric")
		http.Error(w, "failed to convert 'find' result", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeJSON)
	if err = json.NewEncoder(w).Encode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(constants.KeyContentType, constants.ContentTypeHTML)
	wrappedGet := func(args ...any) (any, error) {
		return h.service.Get(r.Context())
	}
	metrics, err := db.WithConnectionCheck(wrappedGet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m, ok := metrics.([]model.Metric)
	if !ok {
		h.log.Error().Msg("failed to convert 'any' to []model.Metric")
		http.Error(w, "failed to convert 'GetAll' result", http.StatusInternalServerError)
		return
	}

	page := createMetricsPage(m)
	_, err = w.Write([]byte(page))
	if err != nil {
		h.log.Error().Err(err).Msg("failed to write response")
	}
}

func (h *HTTPHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		err := errors.New("the dumping service is not initialised")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrappedPing := func(args ...any) (any, error) {
		return nil, h.service.Ping(r.Context())
	}
	_, err := db.WithConnectionCheck(wrappedPing)
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
