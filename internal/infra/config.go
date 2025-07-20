package infra

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type YamlConfigRepository struct {
	filePath string
}

const DefaultConfigFilePath = "./config.yml"

func NewYamlConfigRepository(filePath *string) domain.ConfigRepository {
	p := DefaultConfigFilePath
	if filePath != nil {
		p = *filePath
	}

	return &YamlConfigRepository{
		filePath: p,
	}
}

func (r *YamlConfigRepository) Save(config *domain.Config) error {
	file, err := os.Create(DefaultConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func (r *YamlConfigRepository) Load() (*domain.Config, error) {
	viper.SetConfigFile(r.filePath)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config domain.Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
