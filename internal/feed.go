package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
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

type Article struct {
	Title     string
	Link      string
	Published *time.Time
	Content   string
}

// FetchFeedFunc defines the signature for the FetchFeed function.
type FetchFeedFunc func(url string) ([]Article, error)

// defaultFetchFeed is the default implementation of FetchFeed.
var defaultFetchFeed FetchFeedFunc = func(url string) ([]Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	var articles []Article
	for _, item := range feed.Items {
		content := ""
		if item.Content != "" {
			content = item.Content
		} else if item.Description != "" {
			content = item.Description
		}

		articles = append(articles, Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: item.PublishedParsed,
			Content:   content,
		})
	}
	return articles, nil
}

// FetchFeed is the function that should be called to fetch feeds. It can be overridden for testing.
var FetchFeed = defaultFetchFeed

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
