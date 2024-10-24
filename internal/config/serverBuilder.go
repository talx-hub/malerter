package config

type ServerBuilder struct {
    server Server
}

func (b ServerBuilder) loadFromFlags() Builder {
    b.server.loadFromFlags()
    return b
}

func (b ServerBuilder) loadFromEnv() Builder {
    b.server.loadFromEnv()
    return b
}

func (b ServerBuilder) isValid() (Builder, error) {
    if ok, err := b.server.isValid(); ok {
        return b, nil
    } else {
        return b, err
    }
}

func (b ServerBuilder) build() Config {
    return b.server
}
