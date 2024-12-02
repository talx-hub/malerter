package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

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
	metrics := s.get()
	jsons := convertToJSONs(metrics)
	send(s.client, s.host, jsons, s.compress)
}

func (s *Sender) get() []model.Metric {
	return s.storage.Get()
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

func send(client *http.Client, host string, jsons []string, compress bool) {
	for _, j := range jsons {
		var body *bytes.Buffer
		var err error
		if compress {
			body, err = gzip.Compress([]byte(j))
			if err != nil {
				log.Printf("unable to compress json %s: %v", j, err)
			}
		} else {
			body = bytes.NewBufferString(j)
		}
		request, err := http.NewRequest(http.MethodPost, host+"/update/", body)
		if err != nil {
			log.Printf("unable to send json %s to %s: %v", j, host, err)
			continue
		}
		request.Header.Set(constants.KeyContentType, constants.ContentTypeJSON)
		request.Header.Set(constants.KeyContentEncoding, "gzip")
		response, err := client.Do(request)
		if err != nil {
			log.Printf("unable to send json %s to %s: %v", j, host, err)
			continue
		}
		err = response.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
