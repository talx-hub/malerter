package service

import (
    "log"
    "os"
    "time"

    "github.com/talx-hub/malerter/internal/config"
    "github.com/talx-hub/malerter/internal/repo"
)

type Agent struct {
    config *config.Agent
    repo   repo.Repository
    poller Poller
    sender Sender
}

func NewAgent(repo repo.Repository, cfg *config.Agent) *Agent {
    poller := Poller{repo: repo}
    sender := Sender{repo: repo, host: cfg.ServerAddress}
    return &Agent{
        config: cfg,
        repo:   repo,
        poller: poller,
        sender: sender,
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
