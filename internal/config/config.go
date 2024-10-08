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

func LoadServerConfig() (*ServerConfig, error) {
	var config ServerConfig
	flag.StringVar(&config.RootAddress, "a", "localhost:8080",
		"server root address")
	flag.Parse()

	if a, found := os.LookupEnv("ADDRESS"); found {
		config.RootAddress = a
	}

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

	var tempReportI int
	flag.IntVar(&tempReportI, "r", 10,
		"interval in seconds of sending metrics to alert server")
	if tempReportI < 0 {
		log.Fatal("report interval must be positive")
	}
	config.ReportInterval = time.Duration(tempReportI)

	var tempPollI int
	flag.IntVar(&tempPollI, "p", 2,
		"interval in seconds of polling and metric collection")
	if tempPollI < 0 {
		log.Fatal("poll interval must be positive")
	}
	config.PollInterval = time.Duration(tempPollI)
	flag.Parse()

	if a, found := os.LookupEnv("ADDRESS"); found {
		config.ServerAddress = a
	}
	if ri, found := os.LookupEnv("REPORT_INTERVAL"); found {
		riInt, err := strconv.Atoi(ri)
		if err != nil {
			log.Fatal(err)
		}
		if riInt < 0 {
			log.Fatal("report interval must be positive")
		}
		config.ReportInterval = time.Duration(riInt)
	}
	if rp, found := os.LookupEnv("POLL_INTERVAL"); found {
		rpInt, err := strconv.Atoi(rp)
		if err != nil {
			log.Fatal(err)
		}
		if rpInt < 0 {
			log.Fatal("poll interval must be positive")
		}
		config.PollInterval = time.Duration(rpInt)
	}

	return &config, nil
}
