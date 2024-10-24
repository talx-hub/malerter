package config

import "log"

type Director struct {
    builder Builder
}

func NewAgentDirector() *Director {
    return &Director{
        builder: AgentBuilder{},
    }
}

func NewServerDirector() *Director {
    return &Director{
        builder: ServerBuilder{},
    }
}

func (d *Director) Build() Config {
    cfg, err := d.builder.
        loadFromFlags().
        loadFromEnv().
        isValid()

    if err != nil {
        log.Fatal(err)
    }
    return cfg.build()
}
