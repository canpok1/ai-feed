package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/canpok1/ai-feed/internal"
)

func TestInstantRecommendCommand(t *testing.T) {
	// Save original FetchFeed and restore it after test
	originalFetchFeed := internal.FetchFeed
	defer func() {
		internal.FetchFeed = originalFetchFeed
	}()

	// Save original log output and restore it after test
	oldLogOutput := log.Writer()
	defer func() {
		log.SetOutput(oldLogOutput)
	}()

	tests := []struct {
		name                string
		url                 string
		sourceFileContent   string
		mockArticles        map[string][]internal.Article // Map URL to articles
		expectedOutput      []string
		expectedErrorOutput string
		expectError         bool
	}{
		{
			name: "Successful recommendation with URL flag",
			url:  "http://example.com/feed.xml",
			mockArticles: map[string][]internal.Article{
				"http://example.com/feed.xml": {
					{Title: "Article 1", Link: "http://example.com/article1"},
					{Title: "Article 2", Link: "http://example.com/article2"},
				},
			},
			expectedOutput:      []string{"Title: Article 1\nLink: http://example.com/article1\n", "Title: Article 2\nLink: http://example.com/article2\n"},
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "No articles found with URL flag",
			url:                 "http://example.com/empty.xml",
			mockArticles:        map[string][]internal.Article{"http://example.com/empty.xml": {}},
			expectedOutput:      []string{"No articles found in the feed.\n"},
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "Fetch feed error with URL flag",
			url:                 "http://invalid.com/feed.xml",
			mockArticles:        map[string][]internal.Article{},
			expectedOutput:      []string{""},
			expectedErrorOutput: "Error: failed to fetch feed: mock fetch error\n",
			expectError:         true,
		},
		{
			name:              "Successful recommendation with source file",
			sourceFileContent: "http://example.com/feed1.xml\nhttp://example.com/feed2.xml",
			mockArticles: map[string][]internal.Article{
				"http://example.com/feed1.xml": {
					{Title: "Article A", Link: "http://example.com/articleA"},
				},
				"http://example.com/feed2.xml": {
					{Title: "Article B", Link: "http://example.com/articleB"},
				},
			},
			expectedOutput: []string{
				"Title: Article A\nLink: http://example.com/articleA\n",
				"Title: Article B\nLink: http://example.com/articleB\n",
			},
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "Source file with no URLs",
			sourceFileContent:   "\n\n",
			mockArticles:        map[string][]internal.Article{},
			expectedOutput:      []string{""},
			expectedErrorOutput: "Error: source file contains no URLs\n",
			expectError:         true,
		},
		{
			name:                "Non-existent source file",
			url:                 "", // Not used, but to satisfy struct
			sourceFileContent:   "", // Will be replaced by a non-existent path
			mockArticles:        map[string][]internal.Article{},
			expectedOutput:      []string{""},
			expectedErrorOutput: "Error: failed to read URLs from file: open", // Partial match
			expectError:         true,
		},
		{
			name:              "Mixed success and error with source file",
			sourceFileContent: "http://example.com/good.xml\nhttp://example.com/bad.xml",
			mockArticles: map[string][]internal.Article{
				"http://example.com/good.xml": {
					{Title: "Good Article", Link: "http://example.com/good_article"},
				},
				"http://example.com/bad.xml": nil, // Simulate error for this URL
			},
			expectedOutput: []string{
				"Title: Good Article\nLink: http://example.com/good_article\n",
			},
			expectedErrorOutput: "Error fetching feed from http://example.com/bad.xml: mock fetch error\n",
			expectError:         false, // Command itself should not error, just log fetch error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command flags and output buffers for each test
			cmd := instantRecommendCmd
			cmd.SetArgs([]string{})
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))

			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)
			cmd.SetOut(stdoutBuffer)
			cmd.SetErr(stderrBuffer)

			// Prepare command arguments
			args := []string{"instant-recommend"}
			if tt.url != "" {
				args = append(args, "--url", tt.url)
			}

			sourceFilePath := ""
			if tt.sourceFileContent != "" {
				var err error
				sourceFilePath, err = internal.CreateTempFile(tt.sourceFileContent)
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(sourceFilePath)
				args = append(args, "--source", sourceFilePath)
			} else if tt.name == "Non-existent source file" {
				// Use a path that is guaranteed not to exist
				sourceFilePath = "/path/to/nonexistent/file.txt"
				args = append(args, "--source", sourceFilePath)
			}

			rootCmd.SetArgs(args)

			// Mock FetchFeed to return articles based on the URL
			internal.FetchFeed = func(url string) ([]internal.Article, error) {
				if tt.name == "Fetch feed error with URL flag" || (tt.name == "Mixed success and error with source file" && url == "http://example.com/bad.xml") {
					return nil, fmt.Errorf("mock fetch error")
				}
				if articles, ok := tt.mockArticles[url]; ok {
					return articles, nil
				}
				return nil, fmt.Errorf("unexpected URL: %s", url)
			}
		})
	}
}

func TestDisplayArticle(t *testing.T) {
	var buf bytes.Buffer
	article := internal.Article{
		Title: "Test Article",
		Link:  "http://test.com/article",
	}
	displayArticle(&buf, article)

	expected := "Title: Test Article\nLink: http://test.com/article\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}
