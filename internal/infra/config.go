package infra

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
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
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("config file already exists: %s", r.filePath)
		}
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
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s, %w", r.filePath, err)
	}

	var config entity.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
