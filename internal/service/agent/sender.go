package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/talx-hub/malerter/internal/compressor/gzip"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
)

type Sender struct {
	client   *http.Client
	storage  Storage
	host     string
	compress bool
}

func (s *Sender) send() {
	metrics, _ := s.get()
	jsons := convertToJSONs(metrics)
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

func convertToJSONs(metrics []model.Metric) []string {
	jsons := make([]string, len(metrics))
	for i, m := range metrics {
		mJSON, err := json.Marshal(m)
		if err != nil {
			log.Printf("unable to convert metric %s to JSON: %v", m.String(), err)
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
		body, err = gzip.Compress([]byte(batch))
		if err != nil {
			log.Printf("unable to compress json %s: %v", batch, err)
			return
		}
	} else {
		body = bytes.NewBufferString(batch)
	}
	request, err := http.NewRequest(http.MethodPost, s.host+"/updates/", body)
	if err != nil {
		log.Printf("unable to send json %s to %s: %v", batch, s.host, err)
		return
	}
	request.Header.Set(constants.KeyContentType, constants.ContentTypeJSON)
	request.Header.Set(constants.KeyContentEncoding, "gzip")
	response, err := s.client.Do(request)
	if err != nil {
		log.Printf("unable to send json %s to %s: %v", batch, s.host, err)
		return
	}
	err = response.Body.Close()
	if err != nil {
		log.Fatalf("unable to close the body: %v", err)
	}
}
