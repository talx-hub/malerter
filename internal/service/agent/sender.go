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
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/pkg/compressor"
	"github.com/talx-hub/malerter/pkg/crypto"
	"github.com/talx-hub/malerter/pkg/retry"
	"github.com/talx-hub/malerter/pkg/signature"
)

type Sender struct {
	client    *http.Client
	log       *logger.ZeroLogger
	encrypter *crypto.Encrypter
	host      string
	secret    string
	compress  bool
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

		batch := []byte(join(s.toJSONs(j)))
		compressed, err := s.tryCompress(batch)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to send data")
			continue
		}
		sig := s.trySign(compressed)
		encrypted, err := s.tryEncrypt(compressed)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to send data")
			continue
		}

		s.batch(encrypted, sig, s.compress, s.encrypter != nil)
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

func (s *Sender) tryCompress(data []byte) ([]byte, error) {
	if !s.compress {
		return data, nil
	}
	body, err := compressor.Compress(data)
	if err != nil {
		s.log.Error().Err(err).
			Msgf("unable to compress json %s", string(data))
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}
	return body.Bytes(), nil
}

func (s *Sender) trySign(data []byte) string {
	if s.secret != constants.NoSecret {
		return signature.Hash(data, s.secret)
	}
	return ""
}

func (s *Sender) tryEncrypt(data []byte) ([]byte, error) {
	if s.encrypter == nil {
		return data, nil
	}
	encryptedPayload, err := s.encrypter.Encrypt(data)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to encrypt data")
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return encryptedPayload, nil
}

func (s *Sender) batch(batch []byte, sig string, isCompressed, isEncrypted bool) {
	const unableFormat = "unable to send json %s to %s"

	ctx, cancel := context.WithTimeout(
		context.Background(), constants.TimeoutAgentRequest)
	defer cancel()

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, s.host+"/updates/", bytes.NewBuffer(batch))
	if err != nil {
		s.log.Error().Err(err).Msgf(unableFormat, batch, s.host)
		return
	}
	request.Header.Set(constants.KeyContentType, constants.ContentTypeJSON)

	if sig != "" {
		request.Header.Set(constants.KeyHashSHA256, sig)
	}
	if isCompressed {
		request.Header.Set(constants.KeyContentEncoding, constants.EncodingGzip)
	}
	if isEncrypted {
		request.Header.Set("X-Encrypted", "true")
	}

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
