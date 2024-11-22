package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/compressor"
	"github.com/talx-hub/malerter/internal/model"
)

type Sender struct {
	storage  Storage
	host     string
	client   *http.Client
	compress bool
}

func (s *Sender) send() error {
	metrics := s.get()
	jsons := convertToJSONs(metrics)
	send(s.client, s.host, jsons, s.compress)
	return nil
}

func (s *Sender) get() []model.Metric {
	return s.storage.Get()
}

func convertToJSONs(metrics []model.Metric) []string {
	var jsons []string
	for _, m := range metrics {
		mJSON, err := json.Marshal(m)
		if err != nil {
			log.Printf("unable to convert metric %s to JSON: %v", m.String(), err)
			continue
		}
		jsons = append(jsons, string(mJSON))
	}
	return jsons
}

func send(client *http.Client, host string, jsons []string, compress bool) {
	for _, j := range jsons {
		var body *bytes.Buffer
		var err error
		if compress {
			body, err = compressor.Compress([]byte(j))
			if err != nil {
				log.Printf("unable to compress json %s: %v", j, err)
			}
		} else {
			body = bytes.NewBuffer([]byte(j))
		}
		request, err := http.NewRequest(http.MethodPost, host+"/update/", body)
		if err != nil {
			log.Printf("unable to send json %s to %s: %v", j, host, err)
			continue
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
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
