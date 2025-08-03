package runner

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

func TestNewRecommendRunner(t *testing.T) {
	tests := []struct {
		name             string
		outputConfig     *infra.OutputConfig
		promptConfig     *infra.PromptConfig
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:         "Successful creation with standard viewer only",
			outputConfig: &infra.OutputConfig{},
			promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with SlackAPI viewer",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					APIToken: "test-token",
					Channel:  "#test",
				},
			},
			promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with Misskey viewer",
			outputConfig: &infra.OutputConfig{
				Misskey: &infra.MisskeyConfig{
					APIToken: "test-token",
					APIURL:   "https://test.misskey.io/api",
				},
			},
			promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Error creating Misskey viewer with invalid URL",
			outputConfig: &infra.OutputConfig{
				Misskey: &infra.MisskeyConfig{
					APIToken: "test-token",
					APIURL:   "invalid-url",
				},
			},
			promptConfig:     &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:      true,
			expectedErrorMsg: "failed to create Misskey viewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
			mockRecommender := mock_domain.NewMockRecommender(ctrl)

			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)

			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
				stdoutBuffer,
				stderrBuffer,
				tt.outputConfig,
				tt.promptConfig,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, runner)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, runner)
				assert.NotNil(t, runner.fetcher)
				assert.NotNil(t, runner.recommender)
				assert.NotEmpty(t, runner.viewers)
			}
		})
	}
}

func TestRecommendRunner_Run(t *testing.T) {
	tests := []struct {
		name                        string
		mockFetchClientExpectations func(m *mock_domain.MockFetchClient)
		mockRecommenderExpectations func(m *mock_domain.MockRecommender)
		params                      *RecommendParams
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
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
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
			params: &RecommendParams{
				URLs: []string{"http://example.com/empty.xml"},
			},
			expectedErrorMessage: toStringP("no articles found in the feed"),
		},
		{
			name: "Fetch error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return(nil, fmt.Errorf("mock fetch error")).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &RecommendParams{
				URLs: []string{"http://invalid.com/feed.xml"},
			},
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
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: toStringP("failed to recommend article: mock recommend error"),
		},
		{
			name: "AI model not configured",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommend is not called if AI model or prompt is not found.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: toStringP("AI model or prompt is not configured"),
		},
		{
			name: "Prompt not configured",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommend is not called if AI model or prompt is not found.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
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

			var runner *RecommendRunner
			var runErr error

			var profile infra.Profile
			mockConfig := createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{})

			switch tt.name {
			case "Successful recommendation", "No articles found", "Recommend error", "Fetch error":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, mockConfig.DefaultProfile.Output, mockConfig.DefaultProfile.Prompt)
				profile = *mockConfig.DefaultProfile
			case "AI model not configured":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
				profile = infra.Profile{
					AI:     nil,
					Prompt: mockConfig.DefaultProfile.Prompt,
					Output: &infra.OutputConfig{},
				}
			case "Prompt not configured":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
				profile = infra.Profile{
					AI:     mockConfig.DefaultProfile.AI,
					Prompt: nil,
					Output: &infra.OutputConfig{},
				}
			}

			ctx := context.Background()

			if runErr == nil && runner != nil {
				runErr = runner.Run(ctx, tt.params, profile)
			}

			hasError := runErr != nil
			expectedHasError := tt.expectedErrorMessage != nil

			assert.Equal(t, expectedHasError, hasError, "Expected error state mismatch")
			if expectedHasError {
				assert.Contains(t, runErr.Error(), *tt.expectedErrorMessage, "Error message mismatch")
			} else {
				assert.NoError(t, runErr)
			}
		})
	}
}
