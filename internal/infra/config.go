package infra

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type YamlConfigRepository struct {
	filePath string
}

func NewYamlConfigRepository(filePath string) domain.ConfigRepository {
	return &YamlConfigRepository{
		filePath: filePath,
	}
}

func (r *YamlConfigRepository) Save(config *entity.Config) error {
	if _, err := os.Stat(r.filePath); err == nil {
		return fmt.Errorf("config file already exists: %s", r.filePath)
	}

	file, err := os.Create(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %s, %w", r.filePath, err)
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

func (r *YamlConfigRepository) Load() (*entity.Config, error) {
	viper.SetConfigFile(r.filePath)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %s, %w", r.filePath, err)
		}
	}

	var config entity.Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
