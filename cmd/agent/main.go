package main

import (
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"time"
)

// TODO: сделать клиент модульным:
//		- модуль сбора метрик
//		- модуль отправки метрик
//		- первый модуль собирает метрики
//		- затем оповещает модуль отпраки, что данные готовы
//		- но как сделать нотификацию???

func main() {
	rep := repo.NewMemRepository()
	agent := service.NewAgent(rep)

	var i = 0
	for {
		agent.Update()

		if i%5 == 0 {
			agent.Send()
			i = 0
		}
		i++
		time.Sleep(2 * time.Second)
	}
}
