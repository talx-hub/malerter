package config

import "flag"

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
