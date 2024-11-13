package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repo"
)

type Sender struct {
	repo repo.Repository
	host string
}

func (s *Sender) send() error {
	metrics := s.get()
	jsons := convertToJSONs(metrics)
	send(s.host, jsons)
	return nil
}

func (s *Sender) get() []model.Metric {
	return s.repo.GetAll()
}

func convertToJSONs(metrics []model.Metric) []string {
	var jsons []string
	for _, m := range metrics {
		mJSON, err := json.Marshal(m)
		if err != nil {
			// TODO: хочу логировать всё в одном месте, в main. Формировать пачку ошибок и её возвращать?
			log.Printf("unable to convert metric %s to JSON: %v", m.String(), err)
			continue
		}
		jsons = append(jsons, string(mJSON))
	}
	return jsons
}

func send(host string, jsons []string) {
	for _, j := range jsons {
		body := bytes.NewBuffer([]byte(j))
		response, err := http.Post(host+"/update/", "application/json", body)
		if err != nil {
			// TODO: хочу логировать всё в одном месте, в main. Формировать пачку ошибок и её возвращать?
			log.Printf("unable to send json %s to %s: %v", j, host, err)
			continue
		}
		if err = response.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
