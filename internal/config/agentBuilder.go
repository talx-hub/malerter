package config

type AgentBuilder struct {
	agent Agent
}

func (b AgentBuilder) loadFromFlags() Builder {
	b.agent.loadFromFlags()
	return b
}

func (b AgentBuilder) loadFromEnv() Builder {
	b.agent.loadFromEnv()
	return b
}

func (b AgentBuilder) isValid() (Builder, error) {
	if ok, err := b.agent.isValid(); ok {
		return b, nil
	} else {
		return b, err
	}
}

func (b AgentBuilder) build() Config {
	return b.agent
}
