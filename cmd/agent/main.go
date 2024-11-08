package main

import (
	"log"

	agentCfg "github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service/agent"
)

// TODO: сделать клиент модульным:
//		+ модуль сбора метрик
//		+ модуль отправки метрик
//		+ первый модуль собирает метрики
//		? затем оповещает модуль отпраки, что данные готовы
//		? но как сделать нотификацию???

func main() {
	// TODO: тут какие-то кошмары с указателями(см. config/agent/builder/.Build())... разобраться
	cfg, ok := agentCfg.NewDirector().Build().(agentCfg.Builder)
	if !ok {
		log.Fatal("unable to load agent config")
	}

	rep := repo.NewMemRepository()
	agt := agent.NewAgent(rep, &cfg)

	agt.Run()
}
