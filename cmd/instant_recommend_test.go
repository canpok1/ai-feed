package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestInstantRecommendCommand(t *testing.T) {
	// Save original log output and restore it after test
	oldLogOutput := log.Writer()
	defer func() {
		log.SetOutput(oldLogOutput)
	}()

	tests := []struct {
		name                string
		url                 string
		sourceFileContent   string
		mockArticles        map[string][]entity.Article // Map URL to articles
		expectedOutput      string
		expectedErrorOutput string
		expectError         bool
	}{
		{
			name: "Successful recommendation with URL flag",
			url:  "http://example.com/feed.xml",
			mockArticles: map[string][]entity.Article{
				"http://example.com/feed.xml": {
					{Title: "Article 1", Link: "http://example.com/article1"},
				},
			},
			expectedOutput: strings.Join([]string{
				"Title: Article 1",
				"Link: http://example.com/article1",
				"",
			}, "\n"),
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "No articles found with URL flag",
			url:                 "http://example.com/empty.xml",
			mockArticles:        map[string][]entity.Article{"http://example.com/empty.xml": {}},
			expectedOutput:      "No articles found in the feed.\n",
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "Fetch feed error with URL flag",
			url:                 "http://invalid.com/feed.xml",
			mockArticles:        nil,
			expectedOutput:      "No articles found in the feed.\n",
			expectedErrorOutput: "Error fetching feed from http://invalid.com/feed.xml",
			expectError:         false,
		},
		{
			name:              "Successful recommendation with source file",
			sourceFileContent: "http://example.com/feed1.xml",
			mockArticles: map[string][]entity.Article{
				"http://example.com/feed1.xml": {
					{Title: "Article A", Link: "http://example.com/articleA"},
				},
			},
			expectedOutput: strings.Join([]string{
				"Title: Article A",
				"Link: http://example.com/articleA",
				"",
			}, "\n"),
			expectedErrorOutput: "",
			expectError:         false,
		},
		{
			name:                "Source file with no URLs",
			sourceFileContent:   " ",
			mockArticles:        map[string][]entity.Article{},
			expectedErrorOutput: "Error: source file contains no URLs",
			expectError:         true,
		},
		{
			name:                "Non-existent source file",
			url:                 "", // Not used, but to satisfy struct
			sourceFileContent:   "", // Will be replaced by a non-existent path
			mockArticles:        map[string][]entity.Article{},
			expectedErrorOutput: "Error: failed to read URLs from file: open", // Partial match
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFetchClient := mock_domain.NewMockFetchClient(ctrl)

			for url, articles := range tt.mockArticles {
				mockFetchClient.EXPECT().Fetch(url).Return(articles, nil).AnyTimes()
			}
			if len(tt.mockArticles) == 0 {
				mockFetchClient.EXPECT().Fetch(tt.url).Return(nil, fmt.Errorf("mock fetch error")).AnyTimes()
			}

			cmd := makeInstantRecommendCmd(mockFetchClient, domain.NewFirstRecommender())

			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)
			cmd.SetOut(stdoutBuffer)
			cmd.SetErr(stderrBuffer)

			// Prepare command arguments
			args := []string{}
			if tt.url != "" {
				args = append(args, "--url", tt.url)
			}

			if tt.sourceFileContent != "" {
				sourceFilePath, err := internal.CreateTempFile(tt.sourceFileContent)
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(sourceFilePath)
				args = append(args, "--source", sourceFilePath)
			} else if tt.name == "Non-existent source file" {
				// Use a path that is guaranteed not to exist
				sourceFilePath := "/path/to/nonexistent/file.txt"
				args = append(args, "--source", sourceFilePath)
			}

			cmd.SetArgs(args)
			err := cmd.Execute()
			hasError := err != nil

			assert.Equal(t, tt.expectError, hasError)
			if tt.expectedOutput != "" {
				assert.Equal(t, tt.expectedOutput, stdoutBuffer.String())
			}
			if tt.expectedErrorOutput != "" {
				assert.Contains(t, stderrBuffer.String(), tt.expectedErrorOutput)
			}
		})
	}
}
