package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/talx-hub/malerter/internal/compressor"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type Sender struct {
	client   *http.Client
	storage  Storage
	log      *logger.ZeroLogger
	host     string
	compress bool
}

func (s *Sender) send() {
	metrics, _ := s.get()
	jsons := s.convertToJSONs(metrics)
	batch := "[" + strings.Join(jsons, ",") + "]"
	s.batch(batch)
	s.storage.Clear()
}

func (s *Sender) get() ([]model.Metric, error) {
	metrics, err := s.storage.Get(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics from storage: %w", err)
	}
	return metrics, nil
}

func (s *Sender) convertToJSONs(metrics []model.Metric) []string {
	jsons := make([]string, len(metrics))
	for i, m := range metrics {
		mJSON, err := json.Marshal(m)
		if err != nil {
			s.log.Error().Err(err).
				Msgf("unable to convert metric %s to JSON", m.String())
			continue
		}
		jsons[i] = string(mJSON)
	}
	return jsons
}

func (s *Sender) batch(batch string) {
	var body *bytes.Buffer
	var err error
	if s.compress {
		body, err = compressor.Compress([]byte(batch))
		if err != nil {
			s.log.Error().Err(err).
				Msgf("unable to compress json %s", batch)
			return
		}
	} else {
		body = bytes.NewBufferString(batch)
	}

	const unableFormat = "unable to send json %s to %s"
	request, err := http.NewRequest(http.MethodPost, s.host+"/updates/", body)
	if err != nil {
		s.log.Error().Err(err).Msgf(unableFormat, batch, s.host)
		return
	}
	request.Header.Set(constants.KeyContentType, constants.ContentTypeJSON)
	request.Header.Set(constants.KeyContentEncoding, "gzip")
	response, err := s.try(request, 0)
	if err != nil {
		s.log.Error().Err(err).Msgf(unableFormat, batch, s.host)
		return
	}

	err = response.Body.Close()
	if err != nil {
		s.log.Fatal().Err(err).Msg("unable to close the body")
	}
}

func (s *Sender) try(r *http.Request, count int) (*http.Response, error) {
	response, err := s.client.Do(r)
	if err == nil {
		return response, nil
	}

	const maxAttemptCount = 3
	if count >= maxAttemptCount {
		return nil, errors.New("all attempts to retry request are out")
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		s.log.Info().Msg("connection refused, trying to retry send request...")
		time.Sleep((time.Duration(count*2 + 1)) * time.Second) // count: 0 1 2 -> seconds: 1 3 5.
		return s.try(r, count+1)
	}
	return nil, fmt.Errorf(
		"on attempt #%d error occurred: %w", count, err)
}
