package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	RootAddress string
}

const (
	hostDefault           = "localhost:8080"
	reportIntervalDefault = 10
	poolIntervalDefault   = 2
)

const (
	envAddress        = "ADDRESS"
	envReportInterval = "REPORT_INTERVAL"
	envPollInterval   = "POLL_INTERVAL"
)

func LoadServerConfig() *ServerConfig {
	var config ServerConfig
	flag.StringVar(&config.RootAddress, "a", hostDefault,
		"server root address")
	flag.Parse()

	if a, found := os.LookupEnv(envAddress); found {
		config.RootAddress = a
	}

	return &config
}

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func LoadAgentConfig() *AgentConfig {
	var config AgentConfig
	loadAgentCLConfig(&config)
	loadAgentEnvConfig(&config)

	return &config
}

func loadAgentCLConfig(config *AgentConfig) {
	flag.StringVar(&config.ServerAddress, "a", hostDefault,
		"alert host address")

	flag.DurationVar(&config.ReportInterval, "r", reportIntervalDefault,
		"interval in seconds of sending metrics to alert server")
	if config.ReportInterval < 0 {
		log.Fatal("report interval must be positive")
	}

	flag.DurationVar(&config.PollInterval, "p", poolIntervalDefault,
		"interval in seconds of polling and metric collection")
	if config.PollInterval < 0 {
		log.Fatal("poll interval must be positive")
	}

	flag.Parse()
}

func loadAgentEnvConfig(config *AgentConfig) {
	if a, found := os.LookupEnv(envAddress); found {
		config.ServerAddress = a
	}
	if ri, found := os.LookupEnv(envReportInterval); found {
		riInt, err := strconv.Atoi(ri)
		if err != nil {
			log.Fatal(err)
		}
		if riInt < 0 {
			log.Fatal("report interval must be positive")
		}
		config.ReportInterval = time.Duration(riInt)
	}
	if rp, found := os.LookupEnv(envPollInterval); found {
		rpInt, err := strconv.Atoi(rp)
		if err != nil {
			log.Fatal(err)
		}
		if rpInt < 0 {
			log.Fatal("poll interval must be positive")
		}
		config.PollInterval = time.Duration(rpInt)
	}
}
