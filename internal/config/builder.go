package config

type Builder interface {
    loadFromFlags() Builder
    loadFromEnv() Builder
    isValid() (Builder, error)
    build() Config
}
