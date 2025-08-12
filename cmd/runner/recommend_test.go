package runner

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
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
			name:         "Successful creation with no viewers",
			outputConfig: &infra.OutputConfig{},
			promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with SlackAPI viewer",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with Misskey viewer",
			outputConfig: &infra.OutputConfig{
				Misskey: &infra.MisskeyConfig{
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io/api",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
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

			stderrBuffer := new(bytes.Buffer)

			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
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
				// viewersスライスはStdSender削除により初期状態では空だが、外部連携設定により追加される
				if tt.outputConfig.SlackAPI != nil || tt.outputConfig.Misskey != nil {
					assert.Greater(t, len(runner.viewers), 0)
				} else {
					assert.Equal(t, 0, len(runner.viewers))
				}
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Times(0)
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Times(0)
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
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock recommend error")).Times(1)
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
				// Recommend is called - config is handled by Recommender constructor now.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: nil,
		},
		{
			name: "Prompt not configured",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommend is called - config is handled by Recommender constructor now.
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: nil,
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

			stderrBuffer := new(bytes.Buffer)

			var runner *RecommendRunner
			var runErr error

			var profile infra.Profile
			mockConfig := createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &infra.OutputConfig{})

			switch tt.name {
			case "Successful recommendation", "No articles found", "Recommend error", "Fetch error":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, mockConfig.DefaultProfile.Output, mockConfig.DefaultProfile.Prompt)
				profile = *mockConfig.DefaultProfile
			case "AI model not configured":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
				profile = infra.Profile{
					AI:     nil,
					Prompt: mockConfig.DefaultProfile.Prompt,
					Output: &infra.OutputConfig{},
				}
			case "Prompt not configured":
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, &infra.OutputConfig{}, &infra.PromptConfig{})
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

// TestRecommendRunner_Run_LogOutput tests slog output when recommendation is successful
func TestRecommendRunner_Run_LogOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Set up a test slog handler to capture log output
	var logBuffer bytes.Buffer
	handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	originalLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger) // Restore original logger after test

	mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
	mockRecommender := mock_domain.NewMockRecommender(ctrl)

	stderrBuffer := new(bytes.Buffer)
	mockConfig := createMockConfig(&infra.PromptConfig{CommentPromptTemplate: "test-prompt-template", FixedMessage: "Test Fixed Message"}, &infra.OutputConfig{})

	runner, runErr := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, mockConfig.DefaultProfile.Output, mockConfig.DefaultProfile.Prompt)
	assert.NoError(t, runErr)

	// Set up test data
	testArticles := []entity.Article{
		{Title: "Test Article", Link: "https://example.com/test"},
	}
	testComment := "This is a test comment"
	testRecommend := &entity.Recommend{
		Article: testArticles[0],
		Comment: &testComment,
	}

	// Set up mock expectations
	mockFetchClient.EXPECT().Fetch(gomock.Any()).Return(testArticles, nil)
	mockRecommender.EXPECT().Recommend(gomock.Any(), testArticles).Return(testRecommend, nil)

	// Execute the test
	params := &RecommendParams{URLs: []string{"https://example.com/feed"}}
	profile := *mockConfig.DefaultProfile
	err := runner.Run(context.Background(), params, profile)

	assert.NoError(t, err)

	// Verify log output
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "推薦記事を選択しました")
	assert.Contains(t, logOutput, "\"title\":\"Test Article\"")
	assert.Contains(t, logOutput, "\"link\":\"https://example.com/test\"")
	assert.Contains(t, logOutput, "\"comment\":\"This is a test comment\"")
	assert.Contains(t, logOutput, "\"fixed_message\":\"Test Fixed Message\"")
}
