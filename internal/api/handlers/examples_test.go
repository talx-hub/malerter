package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service/server/logger"
)

// mockStorage is a test implementation of the Storage interface.
type mockStorage struct {
	metrics   []model.Metric
	metric    model.Metric
	failAdd   bool
	failBatch bool
	failFind  bool
	failGet   bool
}

func (m *mockStorage) Add(_ context.Context, _ model.Metric) error {
	if m.failAdd {
		return errors.New("add failed")
	}
	return nil
}

func (m *mockStorage) Batch(_ context.Context, _ []model.Metric) error {
	if m.failBatch {
		return errors.New("batch failed")
	}
	return nil
}

func (m *mockStorage) Find(_ context.Context, _ string) (model.Metric, error) {
	if m.failFind {
		return model.Metric{}, &customerror.NotFoundError{Info: "not found"}
	}
	return m.metric, nil
}

func (m *mockStorage) Get(_ context.Context) ([]model.Metric, error) {
	if m.failGet {
		return nil, &customerror.NotFoundError{Info: "repo error"}
	}
	return m.metrics, nil
}

func (m *mockStorage) Ping(_ context.Context) error {
	return nil
}

// ExampleHTTPHandler_DumpMetricList_success демонстрирует успешный случай вызова DumpMetricList.
func ExampleHTTPHandler_DumpMetricList_success() {
	mock := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	metrics := []model.Metric{
		{Name: "m1", Type: model.MetricTypeGauge, Value: new(float64)},
		{Name: "m2", Type: model.MetricTypeCounter, Delta: new(int64)},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricList(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 200
}

// ExampleHTTPHandler_DumpMetricList_invalidJSON демонстрирует случай, когда передан некорректный JSON.
func ExampleHTTPHandler_DumpMetricList_invalidJSON() {
	mock := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	body := []byte(`{ this is not valid JSON ]`)
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricList(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}

// ExampleHTTPHandler_DumpMetricList_invalidMetric демонстрирует случай, когда передана невалидная метрика.
func ExampleHTTPHandler_DumpMetricList_invalidMetric() {
	mock := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	// Пустое имя и невалидное значение
	invalidMetric := model.Metric{Type: "counter", Name: "", Delta: nil}
	body, _ := json.Marshal([]model.Metric{invalidMetric})

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricList(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 200
}

// ExampleHTTPHandler_DumpMetricList_storageFailure демонстрирует случай ошибки при сохранении в хранилище.
func ExampleHTTPHandler_DumpMetricList_storageFailure() {
	mock := &mockStorage{failBatch: true}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	metrics := []model.Metric{
		{Name: "m1", Type: model.MetricTypeGauge, Value: new(float64)},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricList(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}

// ExampleHTTPHandler_DumpMetricJSON_success демонстрирует успешную отправку метрики.
func ExampleHTTPHandler_DumpMetricJSON_success() {
	st := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	metric := model.Metric{
		Name:  "m1",
		Type:  model.MetricTypeGauge,
		Value: new(float64),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 200
}

// ExampleHTTPHandler_DumpMetricJSON_invalidJSON демонстрирует ошибку при некорректном JSON.
func ExampleHTTPHandler_DumpMetricJSON_invalidJSON() {
	st := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	body := []byte(`{ invalid json ]`)
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}

// ExampleHTTPHandler_DumpMetricJSON_invalidMetric демонстрирует ошибку валидации (невалидная метрика).
func ExampleHTTPHandler_DumpMetricJSON_invalidMetric() {
	st := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	// Отсутствует значение и delta
	invalidMetric := model.Metric{
		Name: "m1",
		Type: model.MetricTypeCounter,
	}
	body, _ := json.Marshal(invalidMetric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 400
}

// ExampleHTTPHandler_DumpMetricJSON_emptyMetric демонстрирует ошибку при пустой метрике.
func ExampleHTTPHandler_DumpMetricJSON_emptyMetric() {
	st := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	emptyMetric := model.Metric{}
	body, _ := json.Marshal(emptyMetric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 400
}

// ExampleHTTPHandler_DumpMetricJSON_storageAddError демонстрирует ошибку при сохранении в Add.
func ExampleHTTPHandler_DumpMetricJSON_storageAddError() {
	st := &mockStorage{failAdd: true}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	metric := model.Metric{
		Name:  "m1",
		Type:  model.MetricTypeGauge,
		Value: new(float64),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 404
}

// ExampleHTTPHandler_DumpMetricJSON_findError демонстрирует ошибку при поиске метрики.
func ExampleHTTPHandler_DumpMetricJSON_findError() {
	st := &mockStorage{failFind: true}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(st, log)

	metric := model.Metric{
		Name:  "m2",
		Type:  model.MetricTypeCounter,
		Delta: new(int64),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.DumpMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}

// ExampleHTTPHandler_GetMetricJSON_success — успешный запрос на получение метрики.
func ExampleHTTPHandler_GetMetricJSON_success() {
	val := 123.456
	mock := &mockStorage{
		metric: model.Metric{
			Name:  "m1",
			Type:  model.MetricTypeGauge,
			Value: &val,
		},
	}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	reqBody, _ := json.Marshal(model.Metric{
		Name: "m1",
		Type: model.MetricTypeGauge,
	})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	handler.GetMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 200
}

// ExampleHTTPHandler_GetMetricJSON_invalidJSON — некорректный JSON.
func ExampleHTTPHandler_GetMetricJSON_invalidJSON() {
	mock := &mockStorage{}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	body := []byte(`{ invalid json ]`)
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.GetMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}

// ExampleHTTPHandler_GetMetricJSON_notFound — метрика не найдена.
func ExampleHTTPHandler_GetMetricJSON_notFound() {
	mock := &mockStorage{
		failFind: true,
	}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	body, _ := json.Marshal(model.Metric{
		Name: "nonexistent",
		Type: model.MetricTypeCounter,
	})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.GetMetricJSON(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 404
}

func ExampleHTTPHandler_GetAll_success() {
	val := 42.0
	mock := &mockStorage{
		metrics: []model.Metric{
			{Name: "cpu", Type: model.MetricTypeGauge, Value: &val},
			{Name: "reqs", Type: model.MetricTypeCounter, Delta: new(int64)},
		},
	}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	handler.GetAll(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()

	fmt.Println("Status code:", result.StatusCode)
	fmt.Println("Content-Type:", w.Header().Get(constants.KeyContentType))

	// Output:
	// Status code: 200
	// Content-Type: text/html
}

func ExampleHTTPHandler_GetAll_storageError() {
	mock := &mockStorage{
		failGet: true,
	}
	log := logger.NewNopLogger()
	handler := handlers.NewHTTPHandler(mock, log)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	handler.GetAll(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log := logger.NewNopLogger()
			log.Error().Err(err).Msg("fail to close")
		}
	}()
	fmt.Println("Status code:", result.StatusCode)

	// Output:
	// Status code: 500
}
