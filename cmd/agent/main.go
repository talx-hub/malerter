package main

import (
	"github.com/alant1t/metricscoll/internal/config"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
)

// TODO: сделать клиент модульным:
//		- модуль сбора метрик
//		- модуль отправки метрик
//		- первый модуль собирает метрики
//		- затем оповещает модуль отпраки, что данные готовы
//		- но как сделать нотификацию???

func main() {
	cfg := config.LoadAgentConfig()

	rep := repo.NewMemRepository()
	agent := service.NewAgent(rep, cfg)

	agent.Run()
}
