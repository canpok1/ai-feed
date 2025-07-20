package domain

type ConfigRepository interface {
	Save(config *Config) error
	Load() (*Config, error)
}
