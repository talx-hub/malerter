package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
	"os"
	"time"
)

type ServerConfig struct {
	RootAddress string
}

const (
	hostDefault = "localhost:8080"
	envAddress  = "ADDRESS"
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
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}
type proxy struct {
	A  string `env:"ADDRESS" envDefault:"localhost:8080"`
	RI int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	PI int64  `env:"POLL_INTERVAL" envDefault:"2"`
}

func LoadAgentConfig() *AgentConfig {
	var config AgentConfig
	loadAgentCLConfig(&config)
	loadAgentEnvConfig(&config)

	// convert because time.Duration is in nanoseconds
	config.PollInterval *= time.Second
	config.ReportInterval *= time.Second

	return &config
}

func loadAgentCLConfig(cfg *AgentConfig) {
	flag.StringVar(&cfg.Address, "a", cfg.Address, "alert host address")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval,
		"interval in seconds of sending metrics to alert server")
	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval,
		"interval in seconds of polling and metric collection")

	if cfg.ReportInterval < 0 || cfg.PollInterval < 0 {
		log.Fatal("interval flags must be positive")
	}

	flag.Parse()
}

func loadAgentEnvConfig(cfg *AgentConfig) {
	// unable to env.Parse directly into cfg due to time.Duration field
	var p proxy
	if err := env.Parse(&p); err != nil {
		log.Fatal(err)
	}
	if p.RI < 0 || p.PI < 0 {
		log.Fatal("interval environment variables must be positive")
	}

	cfg.Address = p.A
	cfg.ReportInterval = time.Duration(p.RI)
	cfg.PollInterval = time.Duration(p.PI)
}
