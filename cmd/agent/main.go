package main

import (
    "log"

    "github.com/talx-hub/malerter/internal/config"
    "github.com/talx-hub/malerter/internal/repo"
    "github.com/talx-hub/malerter/internal/service"
)

// TODO: сделать клиент модульным:
//		- модуль сбора метрик
//		- модуль отправки метрик
//		- первый модуль собирает метрики
//		- затем оповещает модуль отпраки, что данные готовы
//		- но как сделать нотификацию???

func main() {
    director := config.NewAgentDirector()
    cfg, ok := director.Build().(config.Agent)
    if !ok {
        log.Fatal("unable to load agent config")
    }

    rep := repo.NewMemRepository()
    agent := service.NewAgent(rep, &cfg)

    agent.Run()
}
