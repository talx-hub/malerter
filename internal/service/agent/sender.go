package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"syscall"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service/server/logger"
	"github.com/talx-hub/malerter/internal/utils/compressor"
	"github.com/talx-hub/malerter/internal/utils/retry"
	"github.com/talx-hub/malerter/internal/utils/signature"
)

type Sender struct {
	client   *http.Client
	log      *logger.ZeroLogger
	host     string
	secret   string
	compress bool
}

func (s *Sender) send(
	jobs <-chan chan model.Metric, m *sync.Mutex, wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		m.Lock()
		jobCount := len(jobs)
		if jobCount == 0 {
			m.Unlock()
			return
		}

		j, ok := <-jobs
		if !ok {
			m.Unlock()
			return
		}
		m.Unlock()

		batch := join(s.toJSONs(j))
		s.batch(batch)
	}
}

func (s *Sender) toJSONs(ch <-chan model.Metric) chan string {
	jsons := make(chan string)

	go func() {
		defer close(jsons)
		for m := range ch {
			mJSON, err := json.Marshal(m)
			if err != nil {
				s.log.Error().Err(err).
					Msgf("unable to convert metric %s to JSON", m.String())
				continue
			}
			jsons <- string(mJSON)
		}
	}()

	return jsons
}

func join(ch <-chan string) string {
	jsons := make([]string, 0)
	for j := range ch {
		jsons = append(jsons, j)
	}
	return "[" + strings.Join(jsons, ",") + "]"
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
	request.Header.Set(constants.KeyContentEncoding, constants.EncodingGzip)

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
