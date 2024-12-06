package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

func fromJSON(body io.Reader) (model.Metric, error) {
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

func fromJSONs(body io.Reader) ([]model.Metric, error) {
	var metrics []model.Metric
	if err := json.NewDecoder(body).Decode(&metrics); err != nil {
		return nil,
			fmt.Errorf("unable to decode batch: %w", err)
	}

	validList := make([]model.Metric, 0)
	for _, m := range metrics {
		if err := m.CheckValid(); err != nil || m.IsEmpty() {
			continue
		}
		validList = append(validList, m)
	}
	return validList, nil
}

func (h *HTTPHandler) DumpMetricList(w http.ResponseWriter, r *http.Request) {
	metrics, err := fromJSONs(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}

	if err = h.service.Batch(context.Background(), metrics); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			if err = h.service.Batch(context.Background(), metrics); err != nil {
				st := getStatusFromError(err)
				http.Error(w, err.Error(), st)
				return
			}
		} else {
			st := getStatusFromError(err)
			http.Error(w, err.Error(), st)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) DumpMetricJSON(w http.ResponseWriter, r *http.Request) {
	metric, err := fromJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(w, err.Error(), st)
		return
	}
	if metric.IsEmpty() {
		http.Error(w, "failed to dump empty metric", http.StatusBadRequest)
		return
	}

	if err = h.service.Add(context.Background(), metric); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			if err = h.service.Add(context.Background(), metric); err != nil {
				http.Error(
					w,
					fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
					http.StatusNotFound)
				return
			}
		} else {
			http.Error(
				w,
				fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
				http.StatusNotFound)
			return
		}
	}

	dummyKey := metric.Type.String() + " " + metric.Name
	metric, err = h.service.Find(context.Background(), dummyKey)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			metric, err = h.service.Find(context.Background(), dummyKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	err = h.service.Add(context.Background(), metric)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			err = h.service.Add(context.Background(), metric)
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
					http.StatusNotFound)
				return
			}
		} else {
			http.Error(
				w,
				fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
				http.StatusNotFound)
			return
		}
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

	dummyKey := metric.Type.String() + " " + metric.Name
	m, err := h.service.Find(context.Background(), dummyKey)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			m, err = h.service.Find(context.Background(), dummyKey)
			if err != nil {
				st := getStatusFromError(err)
				http.Error(
					w,
					fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
					st)
				return
			}
		} else {
			st := getStatusFromError(err)
			http.Error(
				w,
				fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
				st)
			return
		}
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
	metric, err := fromJSON(r.Body)
	if err != nil {
		st := getStatusFromError(err)
		http.Error(
			w,
			fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
			st)
		return
	}

	dummyKey := metric.Type.String() + " " + metric.Name
	metric, err = h.service.Find(context.Background(), dummyKey)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			metric, err = h.service.Find(context.Background(), dummyKey)
			if err != nil {
				st := getStatusFromError(err)
				http.Error(
					w,
					fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
					st)
				return
			}
		} else {
			st := getStatusFromError(err)
			http.Error(
				w,
				fmt.Sprintf(errMsgPattern, r.URL.Path, err.Error()),
				st)
			return
		}
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
	metrics, err := h.service.Get(context.Background())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
			metrics, err = h.service.Get(context.Background())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	page := createMetricsPage(metrics)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(page))
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry(context.Background(), 0, h, w)
			if err != nil {
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func retry(ctx context.Context, count int, h *HTTPHandler, w http.ResponseWriter) error {
	const maxAttemptCount = 4
	if count == maxAttemptCount {
		err := errors.New("all attempts to retry are out")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err := h.service.Ping(ctx)
	if err != nil && errors.Is(err, syscall.ECONNREFUSED) {
		time.Sleep((time.Duration(count*2 + 1)) * time.Second) // count: 0 1 2 -> seconds: 1 3 5.
		return retry(ctx, count+1, h, w)
	}
	return nil
}
