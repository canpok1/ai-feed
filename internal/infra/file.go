package infra

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func loadYaml[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var v T
	err = yaml.Unmarshal(data, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &v, nil
}
