// Package handlers предоставляет HTTP-обработчики для управления метриками.
// Он включает реализацию REST API, поддерживающую JSON и URL-параметры для операций чтения и записи метрик.
package handlers

import (
	"context"
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
)

const (
	errMsgPattern = `%s fails: %s`
)

// Storage определяет интерфейс для операций с хранилищем метрик.
type Storage interface {
	// Add сохраняет одну метрику.
	Add(ctx context.Context, metric model.Metric) error

	// Batch сохраняет несколько метрик за одну операцию.
	Batch(ctx context.Context, metrics []model.Metric) error

	// Find возвращает метрику по ключу.
	Find(ctx context.Context, key string) (model.Metric, error)

	// Get возвращает все метрики.
	Get(ctx context.Context) ([]model.Metric, error)

	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
}

// HTTPHandler реализует HTTP API для работы с метриками.
// Он использует хранилище метрик и логгер для обработки запросов.
type HTTPHandler struct {
	storage Storage
	log     *logger.ZeroLogger
}

// NewHTTPHandler создаёт новый экземпляр HTTPHandler.
func NewHTTPHandler(s Storage, log *logger.ZeroLogger) *HTTPHandler {
	return &HTTPHandler{storage: s, log: log}
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

// DumpMetricList сохраняет список метрик, переданный в теле запроса в формате JSON.
//
// Пример запроса: POST /updates/.
func (h *HTTPHandler) DumpMetricList(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.extractJSONs(r.Body)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to extract metrics from JSON")
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	wrappedBatch := func(args ...any) (any, error) {
		return nil, h.storage.Batch(r.Context(), metrics)
	}
	if _, err = db.WithConnectionCheck(wrappedBatch); err != nil {
		h.log.Error().Err(err).Msg("failed to dump metrics in repo")
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DumpMetricJSON сохраняет метрику, переданную в теле запроса в формате JSON.
//
// Пример запроса: POST /update/.
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
		return nil, h.storage.Add(r.Context(), metric)
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
		return h.storage.Find(r.Context(), dummyKey)
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

// DumpMetric сохраняет метрику, переданную в виде URL-параметров.
//
// Пример запроса: POST /update/{type}/{name}/{value}.
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
		return nil, h.storage.Add(r.Context(), metric)
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

// GetMetric возвращает значение метрики по имени и типу, переданным в URL.
//
// Пример запроса: GET /value/{type}/{name}.
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
		return h.storage.Find(r.Context(), dummyKey)
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

// GetMetricJSON возвращает значение метрики, переданной в теле запроса в формате JSON.
//
// Пример запроса: POST /value/.
func (h *HTTPHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric, err := extractJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()), st)
		return
	}

	dummyKey := metric.Type.String() + " " + metric.Name
	wrappedFind := func(args ...any) (any, error) {
		return h.storage.Find(r.Context(), dummyKey)
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

// GetAll возвращает все метрики в виде HTML-страницы.
//
// Пример запроса: GET /.
func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(constants.KeyContentType, constants.ContentTypeHTML)
	wrappedGet := func(args ...any) (any, error) {
		return h.storage.Get(r.Context())
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
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Ping проверяет доступность хранилища.
//
// Пример запроса: GET /ping.
func (h *HTTPHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.storage == nil {
		err := errors.New("the dumping service is not initialised")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrappedPing := func(args ...any) (any, error) {
		return nil, h.storage.Ping(r.Context())
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
