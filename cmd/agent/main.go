package main

import (
	"github.com/alant1t/metricscoll/internal/config"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"log"
)

// TODO: сделать клиент модульным:
//		- модуль сбора метрик
//		- модуль отправки метрик
//		- первый модуль собирает метрики
//		- затем оповещает модуль отпраки, что данные готовы
//		- но как сделать нотификацию???

func main() {
	conf, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatal(err)
	}

	rep := repo.NewMemRepository()
	agent := service.NewAgent(rep, conf)

	agent.Run()

}
