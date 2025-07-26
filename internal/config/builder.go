package config

type Builder interface {
	LoadFromFlags() Builder
	LoadFromEnv() Builder
	LoadFromFile() Builder
	IsValid() (Builder, error)
	Build() Config
}
