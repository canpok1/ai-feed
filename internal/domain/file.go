package domain

import (
	"bufio"
	"net/url"
	"os"
	"strings"
)

func ReadURLsFromFile(filePath string, errCallback func(filePath, line string, err error) error) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// URLバリデーション
		if _, err := url.ParseRequestURI(line); err != nil {
			if err := errCallback(filePath, line, err); err != nil {
				return nil, err
			}
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
