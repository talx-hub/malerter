package config

import (
	"flag"
	"log"
	"strings"
	"time"
)

type ServerConfig struct {
	RootAddress string
}

func LoadServerConfig() (*ServerConfig, error) {
	var config ServerConfig
	flag.StringVar(&config.RootAddress, "a", "localhost:8080",
		"server root address")
	flag.Parse()

	return &config, nil
}

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func LoadAgentConfig() (*AgentConfig, error) {
	var config AgentConfig
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080",
		"alert server address")
	if strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress =
			strings.TrimPrefix(config.ServerAddress, "http://")
	}

	var tempReportI int
	flag.IntVar(&tempReportI, "r", 10,
		"interval in seconds of sending metrics to alert server")
	if tempReportI < 0 {
		log.Fatal("interval must be positive")
	}
	config.ReportInterval = time.Duration(tempReportI)

	var tempPollI int
	flag.IntVar(&tempPollI, "p", 2,
		"interval in seconds of polling and metric collection")
	if tempPollI < 0 {
		log.Fatal("interval must be positive")
	}
	config.PollInterval = time.Duration(tempPollI)
	flag.Parse()

	return &config, nil
}
