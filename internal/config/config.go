package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

type Cfg interface {
	Load()
}

const (
	hostDefault = "localhost:8080"
)

const (
	reportIntervalDefault = 10
	poolIntervalDefault   = 2
)

const (
	envAddress        = "ADDRESS"
	envReportInterval = "REPORT_INTERVAL"
	envPollInterval   = "POLL_INTERVAL"
)

type Server struct {
	RootAddress string
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Load() {
	flag.StringVar(&s.RootAddress, "a", hostDefault,
		"server root address")
	flag.Parse()

	if a, found := os.LookupEnv(envAddress); found {
		s.RootAddress = a
	}
}

type Agent struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func NewAgent() *Agent {
	return &Agent{}
}

func (a *Agent) Load() {
	loadAgentCLConfig(a)
	loadAgentEnvConfig(a)

	if a.ReportInterval < 0 {
		log.Fatal("report interval must be positive")
	}
	if a.PollInterval < 0 {
		log.Fatal("poll interval must be positive")
	}
}

func loadAgentCLConfig(config *Agent) {
	flag.StringVar(&config.ServerAddress, "a", hostDefault,
		"alert host address")

	var ri int64
	flag.Int64Var(&ri, "r", reportIntervalDefault,
		"interval in seconds of sending metrics to alert server")

	var pi int64
	flag.Int64Var(&pi, "p", poolIntervalDefault,
		"interval in seconds of polling and metric collection")
	flag.Parse()

	config.ReportInterval = time.Duration(ri) * time.Second
	config.PollInterval = time.Duration(pi) * time.Second
}

func loadAgentEnvConfig(config *Agent) {
	if addr, found := os.LookupEnv(envAddress); found {
		config.ServerAddress = addr
	}
	if ri, found := os.LookupEnv(envReportInterval); found {
		riInt, err := strconv.Atoi(ri)
		if err != nil {
			log.Fatal(err)
		}
		config.ReportInterval = time.Duration(riInt) * time.Second
	}
	if pi, found := os.LookupEnv(envPollInterval); found {
		piInt, err := strconv.Atoi(pi)
		if err != nil {
			log.Fatal(err)
		}
		config.PollInterval = time.Duration(piInt) * time.Second
	}
}
