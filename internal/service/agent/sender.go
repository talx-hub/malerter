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

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/utils/compressor"
	"github.com/talx-hub/malerter/internal/utils/retry"
	"github.com/talx-hub/malerter/internal/utils/signature"
)

type Sender struct {
	client   *http.Client
	storage  Storage
	log      *logger.ZeroLogger
	host     string
	secret   string
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
	const unableFormat = "unable to send json %s to %s"
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

	ctx, cancel := context.WithTimeout(
		context.Background(), constants.TimeoutAgentRequest)
	defer cancel()

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, s.host+"/updates/", body)
	if err != nil {
		s.log.Error().Err(err).Msgf(unableFormat, batch, s.host)
		return
	}

	if s.secret != constants.NoSecret {
		sign := signature.Hash(body.Bytes(), s.secret)
		request.Header.Set(constants.KeyHashSHA256, sign)
	}
	request.Header.Set(constants.KeyContentType, constants.ContentTypeJSON)
	request.Header.Set(constants.KeyContentEncoding, "gzip")

	wrappedDo := func(args ...any) (any, error) {
		response, e := s.client.Do(request)
		if e != nil {
			return nil, fmt.Errorf("request send failed: %w", e)
		}

		errBody := response.Body.Close()
		if errBody != nil {
			s.log.Fatal().Err(err).Msg("unable to close the body")
		}
		//nolint:nilnil // don't need any value from the func and have no error
		return nil, nil
	}
	connectionPred := func(err error) bool {
		return errors.Is(err, syscall.ECONNREFUSED)
	}
	_, err = retry.Try(wrappedDo, connectionPred, 0)
	if err != nil {
		s.log.Error().Err(err).Msgf(unableFormat, batch, s.host)
		return
	}
}
