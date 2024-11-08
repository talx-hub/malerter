package config

type Builder interface {
	LoadFromFlags() Builder
	LoadFromEnv() Builder
	IsValid() (Builder, error)
	Build() Config
}
