package infra

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlProfileRepository struct {
	filePath string
}

func NewYamlProfileRepository(filePath string) *YamlProfileRepository {
	return &YamlProfileRepository{
		filePath: filePath,
	}
}

func (r *YamlProfileRepository) LoadProfile() (*Profile, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}
