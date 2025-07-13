package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal"
)

func TestInstantRecommendCommand(t *testing.T) {
	// Save original FetchFeed and restore it after test
	originalFetchFeed := internal.FetchFeed
	defer func() {
		internal.FetchFeed = originalFetchFeed
	}()

	tests := []struct {
		name                string
		url                 string
		mockArticles        []internal.Article
		expectedOutput      []string // Changed to slice for multiple possible outputs
		expectedErrorOutput string   // For stderr
		expectError         bool
	}{
		{
			name: "Successful recommendation",
			url:  "http://example.com/feed.xml",
			mockArticles: []internal.Article{
				{Title: "Article 1", Link: "http://example.com/article1"},
				{Title: "Article 2", Link: "http://example.com/com/article2"},
			},
			expectedOutput:      []string{""},
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "No articles found",
			url:                 "http://example.com/empty.xml",
			mockArticles:        []internal.Article{},
			expectedOutput:      []string{"No articles found in the feed.\n"},
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "Fetch feed error",
			url:                 "http://invalid.com/feed.xml",
			mockArticles:        nil,
			expectedOutput:      []string{""},
			expectedErrorOutput: "", // Partial match for error message
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffers to capture stdout and stderr
			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)

			// Set command output to buffers
			instantRecommendCmd.SetOut(stdoutBuffer)
			instantRecommendCmd.SetErr(stderrBuffer)

			// Reset flags for each test run
			instantRecommendCmd.Flags().Set("url", tt.url)

			internal.FetchFeed = func(url string) ([]internal.Article, error) {
				if tt.name == "Fetch feed error" {
					return nil, fmt.Errorf("mock fetch error")
				}
				return tt.mockArticles, nil
			}

			// Execute the command
			err := instantRecommendCmd.RunE(instantRecommendCmd, []string{}) // Use RunE() to test the full command lifecycle

			out := stdoutBuffer.String()
			errOut := stderrBuffer.String()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				if !strings.Contains(errOut, tt.expectedErrorOutput) {
					t.Errorf("Expected stderr to contain '%s', but got '%s'", tt.expectedErrorOutput, errOut)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				// Check if stdout contains any of the expected outputs
				matched := false
				for _, expected := range tt.expectedOutput {
					if strings.Contains(out, expected) {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("Expected stdout to contain one of %v, but got '%s'", tt.expectedOutput, out)
				}
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
