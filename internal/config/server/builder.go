package server

import (
	"flag"
	"os"

	"github.com/talx-hub/malerter/internal/config"
)

const (
	AddressDefault = "localhost:8080"
)

const (
	EnvAddress = "ADDRESS"
)

func NewDirector() *config.Director {
	return &config.Director{
		Builder: &Builder{},
	}
}

type Builder struct {
	RootAddress string
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.RootAddress, "a", AddressDefault, "server root address")
	flag.Parse()
	return b
}

func (b *Builder) LoadFromEnv() config.Builder {
	if a, found := os.LookupEnv(EnvAddress); found {
		b.RootAddress = a
	}
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	return b, nil
}

func (b *Builder) Build() config.Config {
	// TODO: wtf?? почему вынужден возвращать значение?
	return *b
}
