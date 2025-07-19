package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CreateTempFile creates a temporary file with the given content.
func CreateTempFile(content string) (string, error) {
	file, err := os.CreateTemp(os.TempDir(), "test_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	return file.Name(), nil
}

// ReadURLsFromFile reads URLs from a given file, one URL per line.
func ReadURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return urls, nil
}
