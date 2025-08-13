package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// ログ出力をデバッグのために表示
	t.Logf("Log output: %s", logOutput)

	// 複数行のJSONログから推薦記事選択のログエントリを取得
	lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
	var recommendLogLine []byte
	for _, line := range lines {
		if len(line) > 0 && bytes.Contains(line, []byte("推薦記事を選択しました")) {
			recommendLogLine = line
			break
		}
	}

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(recommendLogLine, &logEntry))

	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "推薦記事を選択しました", logEntry["msg"])
	assert.Equal(t, "Test Article", logEntry["title"])
	assert.Equal(t, "https://example.com/test", logEntry["link"])
	assert.Equal(t, "This is a test comment", logEntry["comment"])
	assert.Equal(t, "Test Fixed Message", logEntry["fixed_message"])
}

func TestNewRecommendRunner_EnabledFlags(t *testing.T) {
	tests := []struct {
		name            string
		outputConfig    *infra.OutputConfig
		expectedViewers int
	}{
		{
			name: "SlackAPI有効、Misskey有効（default）",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &infra.MisskeyConfig{
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 2,
		},
		{
			name: "SlackAPI有効、Misskey無効",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &infra.MisskeyConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 1,
		},
		{
			name: "SlackAPI無効、Misskey有効",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &infra.MisskeyConfig{
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 1,
		},
		{
			name: "両方無効",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &infra.MisskeyConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 0,
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
				&infra.PromptConfig{},
			)

			assert.NoError(t, err)
			assert.NotNil(t, runner)
			assert.Equal(t, tt.expectedViewers, len(runner.viewers))
		})
	}
}

func TestRecommendRunner_Run_EnabledFlagsLogging(t *testing.T) {
	tests := []struct {
		name         string
		outputConfig *infra.OutputConfig
		expectedLogs []string
	}{
		{
			name: "Slack無効ログ確認",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{"Slack API出力が無効化されています (enabled: false)"},
		},
		{
			name: "Misskey無効ログ確認",
			outputConfig: &infra.OutputConfig{
				Misskey: &infra.MisskeyConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{"Misskey出力が無効化されています (enabled: false)"},
		},
		{
			name: "両方無効ログ確認",
			outputConfig: &infra.OutputConfig{
				SlackAPI: &infra.SlackAPIConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &infra.MisskeyConfig{
					Enabled:         testutil.BoolPtr(false),
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{
				"Slack API出力が無効化されています (enabled: false)",
				"Misskey出力が無効化されています (enabled: false)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Set up a test slog handler to capture log output
			var logBuffer bytes.Buffer
			handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo})
			logger := slog.New(handler)
			originalLogger := slog.Default()
			slog.SetDefault(logger)
			defer slog.SetDefault(originalLogger)

			mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
			mockRecommender := mock_domain.NewMockRecommender(ctrl)

			stderrBuffer := new(bytes.Buffer)

			// Runner作成時にログが出力される
			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
				stderrBuffer,
				tt.outputConfig,
				&infra.PromptConfig{},
			)

			assert.NoError(t, err)
			assert.NotNil(t, runner)

			// ログ出力を確認
			logOutput := logBuffer.String()
			for _, expectedLog := range tt.expectedLogs {
				assert.Contains(t, logOutput, expectedLog)
			}
		})
	}
}

func TestRecommendRunner_Run_AllOutputsDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
	mockRecommender := mock_domain.NewMockRecommender(ctrl)

	stderrBuffer := new(bytes.Buffer)

	// 両方無効の設定
	outputConfig := &infra.OutputConfig{
		SlackAPI: &infra.SlackAPIConfig{
			Enabled:         testutil.BoolPtr(false),
			APIToken:        "test-token",
			Channel:         "#test",
			MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
		},
		Misskey: &infra.MisskeyConfig{
			Enabled:         testutil.BoolPtr(false),
			APIToken:        "test-token",
			APIURL:          "https://test.misskey.io",
			MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
		},
	}

	runner, err := NewRecommendRunner(
		mockFetchClient,
		mockRecommender,
		stderrBuffer,
		outputConfig,
		&infra.PromptConfig{},
	)

	assert.NoError(t, err)
	assert.NotNil(t, runner)
	assert.Equal(t, 0, len(runner.viewers)) // viewer数は0

	// Set up test data
	testArticles := []entity.Article{
		{Title: "Test Article", Link: "https://example.com/test"},
	}
	testRecommend := &entity.Recommend{
		Article: testArticles[0],
	}

	// Set up mock expectations
	mockFetchClient.EXPECT().Fetch(gomock.Any()).Return(testArticles, nil)
	mockRecommender.EXPECT().Recommend(gomock.Any(), testArticles).Return(testRecommend, nil)

	// Execute the test - エラーにならないことを確認
	params := &RecommendParams{URLs: []string{"https://example.com/feed"}}
	profile := infra.Profile{}
	err = runner.Run(context.Background(), params, profile)

	assert.NoError(t, err) // 全出力先無効でもエラーにならない
}
