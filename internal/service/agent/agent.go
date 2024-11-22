package agent

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(metric model.Metric) error
	Find(key string) (model.Metric, error)
	Get() []model.Metric
}

type Agent struct {
	config  *agent.Builder
	storage Storage
	poller  Poller
	sender  Sender
}

func NewAgent(storage Storage, cfg *agent.Builder, client *http.Client) *Agent {
	return &Agent{
		config:  cfg,
		storage: storage,
		poller:  Poller{storage: storage},
		sender: Sender{
			storage:  storage,
			host:     "http://" + cfg.ServerAddress,
			client:   client,
			compress: true,
		},
	}
}

func (a *Agent) Run() {
	var i = 1
	var updateToSendRatio = int(a.config.ReportInterval / a.config.PollInterval)
	for {
		if err := a.poller.update(); err != nil {
			if _, e := os.Stderr.WriteString(err.Error()); e != nil {
				log.Fatal(e)
			}
		}

		if i%updateToSendRatio == 0 {
			if err := a.sender.send(); err != nil {
				if _, e := os.Stderr.WriteString(err.Error()); e != nil {
					log.Fatal(e)
				}
			}
			i = 0
		}
		i++
		time.Sleep(a.config.PollInterval)
	}
}
