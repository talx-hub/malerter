package main

import (
	"log"
	"net/http"

	agentCfg "github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/agent"
)

// TODO: сделать клиент модульным:
//		+ модуль сбора метрик
//		+ модуль отправки метрик
//		+ первый модуль собирает метрики
//		? затем оповещает модуль отпраки, что данные готовы
//		? но как сделать нотификацию???

func main() {
	// TODO: тут какие-то кошмары с указателями(см. config/agent/builder/.Build())
	cfg, ok := agentCfg.NewDirector().Build().(agentCfg.Builder)
	if !ok {
		log.Fatal("unable to load agent config")
	}

	rep := memory.New()
	agt := agent.NewAgent(rep, &cfg, &http.Client{})

	agt.Run()
}
