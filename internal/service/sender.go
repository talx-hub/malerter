package service

import (
    "github.com/talx-hub/malerter/internal/repo"
    "log"
    "net/http"
)

type Sender struct {
    repo repo.Repository
    host string
}

func (s *Sender) send() error {
    metrics := s.get()
    urls := convertToURLs(metrics, s.host)
    send(urls)
    return nil
}

func (s *Sender) get() []repo.Metric {
    return s.repo.GetAll()
}

func convertToURLs(metrics []repo.Metric, host string) []string {
    var urls []string
    for _, m := range metrics {
        url := "http://" + host + "/update/" + m.ToURL()
        urls = append(urls, url)
    }
    return urls
}

func send(urls []string) {
    for _, url := range urls {
        response, err := http.Post(url, "text/plain", nil)
        if err != nil {
            continue
        }
        if err := response.Body.Close(); err != nil {
            log.Fatal(err)
        }
    }
}
