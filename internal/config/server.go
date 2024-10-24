package config

import (
	"flag"
	"os"
)

type Server struct {
	RootAddress string
}

func (s *Server) loadFromFlags() {
	flag.StringVar(&s.RootAddress, "a", hostDefault,
		"server root address")
	flag.Parse()
}

func (s *Server) loadFromEnv() {
	if a, found := os.LookupEnv(envAddress); found {
		s.RootAddress = a
	}
}

func (s *Server) isValid() (bool, error) {
	return true, nil
}
