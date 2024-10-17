package main

import (
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
    cfg := config.NewAgent()
    cfg.Load()

    rep := repo.NewMemRepository()
    agent := service.NewAgent(rep, cfg)

    agent.Run()
}
