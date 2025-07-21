package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spf13/cobra"
)

// createMockConfig creates a mock entity.Config for testing purposes.
func createMockConfig(modelName, promptName string) *entity.Config {
	return &entity.Config{
		General: entity.GeneralConfig{
			DefaultExecutionProfile: "default",
		},
		AIModels: map[string]entity.AIModelConfig{
			modelName: {Type: "test-type", APIKey: "test-key"},
		},
		Prompts: map[string]entity.PromptConfig{
			promptName: {SystemMessage: "test-system-message", CommentPromptTemplate: "test-prompt-template"},
		},
		ExecutionProfiles: map[string]entity.ExecutionProfile{
			"default": {AIModel: modelName, Prompt: promptName},
		},
	}
}

func toStringP(value string) *string {
	return &value
}

func TestInstantRecommendRunner_Run(t *testing.T) {
	tests := []struct {
		name                        string
		mockFetchClientExpectations func(m *mock_domain.MockFetchClient)
		mockRecommenderExpectations func(m *mock_domain.MockRecommender)
		params                      *instantRecommendParams
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Recommended Article", Link: "http://example.com/recommended"},
				}, nil).Times(1)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/empty.xml"},
				config: createMockConfig("test-model", "test-prompt"),
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://invalid.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout:       "No articles found in the feed.\n",
			expectedStderr:       "Error fetching feed from http://invalid.com/feed.xml: mock fetch error\n",
			expectedErrorMessage: nil,
		},
		{
			name: "Recommend error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock recommend error")).Times(1)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout:       "",
			expectedStderr:       "",
			expectedErrorMessage: toStringP("failed to recommend article: mock recommend error"),
		},
		{
			name:                        "GetDefaultAIModel error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: &entity.Config{}, // No default AI model set
			},
			expectedStdout:       "",
			expectedErrorMessage: toStringP("failed to get default AI model: default execution profile not found: "),
		},
		{
			name:                        "GetDefaultPrompt error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {},
			params: &instantRecommendParams{
				urls: []string{"http://example.com/feed.xml"},
				config: &entity.Config{
					General: entity.GeneralConfig{DefaultExecutionProfile: "default"},
					AIModels: map[string]entity.AIModelConfig{
						"test-model": {Type: "test-type"},
					},
					ExecutionProfiles: map[string]entity.ExecutionProfile{
						"default": {AIModel: "test-model", Prompt: ""}, // No prompt set
					},
				},
			},
			expectedStdout:       "",
			expectedStderr:       "",
			expectedErrorMessage: toStringP("failed to get default prompt: prompt not found: "),
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

			runner, err := newInstantRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer)
			assert.NoError(t, err, "newInstantRecommendRunner should not return an error")

			// Create a dummy cobra.Command for the Run method, as it's required but not directly used in runner.Run logic
			cmd := &cobra.Command{}
			cmd.SetOut(stdoutBuffer)
			cmd.SetErr(stderrBuffer)

			err = runner.Run(cmd, tt.params)
			hasError := err != nil
			expectedHasError := tt.expectedErrorMessage != nil

			assert.Equal(t, expectedHasError, hasError, "Expected error state mismatch")
			if hasError {
				assert.Contains(t, err.Error(), *tt.expectedErrorMessage, "Error message mismatch")
			}

			assert.Equal(t, tt.expectedStdout, stdoutBuffer.String(), "Stdout mismatch")
			assert.Equal(t, tt.expectedStderr, stderrBuffer.String(), "Stderr mismatch")
		})
	}
}
