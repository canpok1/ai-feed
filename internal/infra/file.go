package infra

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadYAML はYAMLファイルを読み込んで指定された型にデコードする
func LoadYAML[T any](filePath string) (*T, error) {
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

// ReadURLsFromFile reads URLs from a given file, one URL per line.
func ReadURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning file %s: %w", filePath, err)
	}

	return urls, nil
}
