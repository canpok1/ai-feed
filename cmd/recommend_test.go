package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spf13/cobra"
)

// createMockConfig creates a mock entity.Config for testing purposes.
func createMockConfig(promptConfig *infra.PromptConfig, outputConfig *infra.OutputConfig) *infra.Config {
	return &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{Type: "test-type", APIKey: "test-key"},
			},
			Prompt: promptConfig,
			Output: outputConfig,
		},
	}
}

func toStringP(value string) *string {
	return &value
}

func TestRecommendRunner_Run(t *testing.T) {
	tests := []struct {
		name                        string
		mockFetchClientExpectations func(m *mock_domain.MockFetchClient)
		mockRecommenderExpectations func(m *mock_domain.MockRecommender)
		params                      *recommendParams
		expectedStdout              string
		expectedStderr              string
		expectedErrorMessage        *string
	}{
		{
			name: "Successful recommendation",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Recommended Article", Link: "http://example.com/recommended"},
				}, nil).Times(1)
			},
			params: &recommendParams{
				urls: []string{"http://example.com/feed.xml"},
			},
			expectedStdout: strings.Join([]string{
				"Title: Recommended Article",
				"Link: http://example.com/recommended",
				"",
			}, "\n"),
			expectedStderr:       "",
			expectedErrorMessage: nil,
		},
		{
			name: "No articles found",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Should not be called if no articles are found
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &recommendParams{
				urls: []string{"http://example.com/empty.xml"},
			},
			expectedStdout:       "No articles found in the feed.\n",
			expectedStderr:       "",
			expectedErrorMessage: nil,
		},
		{
			name: "Fetch error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return(nil, fmt.Errorf("mock fetch error")).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &recommendParams{
				urls: []string{"http://invalid.com/feed.xml"},
			},
			expectedStdout:       "", // Changed to empty string
			expectedStderr:       "Error fetching feed from http://invalid.com/feed.xml: mock fetch error\n",
			expectedErrorMessage: toStringP("failed to fetch articles: mock fetch error"),
		},
		{
			name: "Recommend error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock recommend error")).Times(1)
			},
			params: &recommendParams{
				urls: []string{"http://example.com/feed.xml"},
			},
			expectedStdout:       "",
			expectedStderr:       "",
			expectedErrorMessage: toStringP("failed to recommend article: mock recommend error"),
		},
		{
			name: "GetDefaultAIModel error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				// This is called before the AI model is checked.
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommend is not called if AI model or prompt is not found.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &recommendParams{
				urls: []string{"http://example.com/feed.xml"},
			},
			expectedStdout:       "",
			expectedErrorMessage: toStringP("AI model or prompt is not configured"),
		},
		{
			name: "GetDefaultPrompt error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				// This is called before the prompt is checked.
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommend is not called if AI model or prompt is not found.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &recommendParams{
				urls: []string{"http://example.com/feed.xml"},
			},
			expectedStdout:       "",
			expectedErrorMessage: toStringP("AI model or prompt is not configured"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
			tt.mockFetchClientExpectations(mockFetchClient)

			mockRecommender := mock_domain.NewMockRecommender(ctrl)
			tt.mockRecommenderExpectations(mockRecommender)

			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)

			var runner *recommendRunner
			var runErr error

			var profile *infra.Profile

			switch tt.name {
			case "Successful recommendation", "No articles found", "Recommend error", "Fetch error":
				runner, runErr = newRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{}).DefaultProfile.Output, createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{}).DefaultProfile.Prompt)
				profile = createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{}).DefaultProfile
			case "GetDefaultAIModel error":
				runner, runErr = newRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
				profile = &infra.Profile{
					AI:     nil,
					Prompt: createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{}).DefaultProfile.Prompt,
					Output: &infra.OutputConfig{},
				}

			case "GetDefaultPrompt error":
				runner, runErr = newRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
				profile = &infra.Profile{
					AI:     createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{}).DefaultProfile.AI,
					Prompt: nil,
					Output: &infra.OutputConfig{},
				}

			}
			// No need for the if/else block here anymore, as expectations are set within the switch.

			// Create a dummy cobra.Command for the Run method, as it's required but not directly used in runner.Run logic
			cmd := &cobra.Command{}
			cmd.SetOut(stdoutBuffer)
			cmd.SetErr(stderrBuffer)

			if runErr == nil && runner != nil {
				runErr = runner.Run(cmd, tt.params, *profile)
			}

			hasError := runErr != nil
			expectedHasError := tt.expectedErrorMessage != nil

			assert.Equal(t, expectedHasError, hasError, "Expected error state mismatch")
			if expectedHasError {
				assert.Contains(t, runErr.Error(), *tt.expectedErrorMessage, "Error message mismatch")
			} else {
				assert.NoError(t, runErr)
			}

			assert.Equal(t, tt.expectedStdout, stdoutBuffer.String(), "Stdout mismatch")
			assert.Equal(t, tt.expectedStderr, stderrBuffer.String(), "Stderr mismatch")
		})
	}
}

func TestNewRecommendParams(t *testing.T) {
	tests := []struct {
		name         string
		urlFlag      string
		sourceFlag   string
		expectedURLs []string
		expectedErr  string
	}{
		{
			name:         "URL flag only",
			urlFlag:      "http://example.com/feed.xml",
			sourceFlag:   "",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "Source flag only with valid file",
			urlFlag:      "",
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml"},
		},
		{
			name:         "Both URL and source flags",
			urlFlag:      "http://example.com/feed.xml",
			sourceFlag:   "tmp_source.txt",
			expectedURLs: nil,
			expectedErr:  "cannot use --url and --source options together",
		},
		{
			name:         "Neither URL nor source flags",
			urlFlag:      "",
			sourceFlag:   "",
			expectedURLs: nil,
			expectedErr:  "either --url or --source must be specified",
		},
		{
			name:         "Source file not found",
			urlFlag:      "",
			sourceFlag:   "non_existent_file.txt",
			expectedURLs: nil,
			expectedErr:  "failed to read URLs from file: open non_existent_file.txt: no such file or directory",
		},
		{
			name:         "Empty source file",
			urlFlag:      "",
			sourceFlag:   "empty_source.txt",
			expectedURLs: nil,
			expectedErr:  "source file contains no URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy cobra.Command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
			cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

			if tt.urlFlag != "" {
				cmd.Flags().Set("url", tt.urlFlag)
			}
			if tt.sourceFlag != "" {
				// Create temporary source file if sourceFlag is used
				if tt.sourceFlag == "tmp_source.txt" {
					content := "http://example.com/from_file.xml\nhttp://another.com/from_file.xml"
					err := os.WriteFile(tt.sourceFlag, []byte(content), 0644)
					assert.NoError(t, err)
					defer os.Remove(tt.sourceFlag)
				} else if tt.sourceFlag == "empty_source.txt" {
					err := os.WriteFile(tt.sourceFlag, []byte(""), 0644)
					assert.NoError(t, err)
					defer os.Remove(tt.sourceFlag)
				}
				cmd.Flags().Set("source", tt.sourceFlag)
			}

			params, err := newRecommendParams(cmd)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, params)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
				assert.Equal(t, tt.expectedURLs, params.urls)
			}
		})
	}
}
