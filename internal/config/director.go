package config

import (
	"log"
)

type Director struct {
	Builder Builder
}

func (d *Director) Build() Config {
	cfg, err := d.Builder.
		LoadFromFlags().
		LoadFromEnv().
		LoadFromFile().
		IsValid()

	if err != nil {
		log.Fatal(err)
	}
	return cfg.Build()
}
