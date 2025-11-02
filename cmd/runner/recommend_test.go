package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// createMockConfig creates a mock entity.Config for testing purposes.
func createMockConfig(promptConfig *entity.PromptConfig, outputConfig *entity.OutputConfig) *entity.Profile {
	return &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{Type: "test-type", APIKey: "test-key"},
		},
		Prompt: promptConfig,
		Output: outputConfig,
	}
}

func toStringP(value string) *string {
	return &value
}

func TestNewRecommendRunner(t *testing.T) {
	tests := []struct {
		name             string
		outputConfig     *entity.OutputConfig
		promptConfig     *entity.PromptConfig
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:         "Successful creation with no viewers",
			outputConfig: &entity.OutputConfig{},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with SlackAPI viewer",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         true,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Successful creation with Misskey viewer",
			outputConfig: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					Enabled:         true,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io/api",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "Error creating Misskey viewer with invalid URL",
			outputConfig: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					Enabled:  true,
					APIToken: "test-token",
					APIURL:   "invalid-url",
				},
			},
			promptConfig:     &entity.PromptConfig{CommentPromptTemplate: "test-template"},
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

			stdoutBuffer := new(bytes.Buffer)
			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
				stderrBuffer,
				stdoutBuffer,
				tt.outputConfig,
				tt.promptConfig,
				nil, // cacheConfig
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
			expectedErrorMessage: toStringP("no articles found in the feed"),
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

			var profile *entity.Profile
			mockProfile := createMockConfig(&entity.PromptConfig{CommentPromptTemplate: "test-prompt-template"}, &entity.OutputConfig{})

			switch tt.name {
			case "Successful recommendation", "No articles found", "Recommend error", "Fetch error":
				stdoutBuffer := new(bytes.Buffer)
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, mockProfile.Output, mockProfile.Prompt, nil)
				profile = mockProfile
			case "AI model not configured":
				stdoutBuffer := new(bytes.Buffer)
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, &entity.OutputConfig{}, &entity.PromptConfig{}, nil)
				profile = &entity.Profile{
					AI:     nil,
					Prompt: mockProfile.Prompt,
					Output: &entity.OutputConfig{},
				}
			case "Prompt not configured":
				stdoutBuffer := new(bytes.Buffer)
				runner, runErr = NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, &entity.OutputConfig{}, &entity.PromptConfig{}, nil)
				profile = &entity.Profile{
					AI:     mockProfile.AI,
					Prompt: nil,
					Output: &entity.OutputConfig{},
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
	mockProfile := createMockConfig(&entity.PromptConfig{CommentPromptTemplate: "test-prompt-template", FixedMessage: "Test Fixed Message"}, &entity.OutputConfig{})

	stdoutBuffer := new(bytes.Buffer)
	runner, runErr := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, mockProfile.Output, mockProfile.Prompt, nil)
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
	profile := mockProfile
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
		if len(line) > 0 && bytes.Contains(line, []byte("Recommendation article selected")) {
			recommendLogLine = line
			break
		}
	}

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(recommendLogLine, &logEntry))

	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "Recommendation article selected", logEntry["msg"])
	assert.Equal(t, "Test Article", logEntry["title"])
	assert.Equal(t, "https://example.com/test", logEntry["link"])
	assert.Equal(t, "This is a test comment", logEntry["comment"])
	assert.Equal(t, "Test Fixed Message", logEntry["fixed_message"])
}

func TestNewRecommendRunner_EnabledFlags(t *testing.T) {
	tests := []struct {
		name            string
		outputConfig    *entity.OutputConfig
		expectedViewers int
	}{
		{
			name: "SlackAPI有効、Misskey有効（default）",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         true,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &entity.MisskeyConfig{
					Enabled:         true,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 2,
		},
		{
			name: "SlackAPI有効、Misskey無効",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         true,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &entity.MisskeyConfig{
					Enabled:         false,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 1,
		},
		{
			name: "SlackAPI無効、Misskey有効",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         false,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &entity.MisskeyConfig{
					Enabled:         true,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedViewers: 1,
		},
		{
			name: "両方無効",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         false,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &entity.MisskeyConfig{
					Enabled:         false,
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

			stdoutBuffer := new(bytes.Buffer)
			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
				stderrBuffer,
				stdoutBuffer,
				tt.outputConfig,
				&entity.PromptConfig{},
				nil, // cacheConfig
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
		outputConfig *entity.OutputConfig
		expectedLogs []string
	}{
		{
			name: "Slack無効ログ確認",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         false,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{"Slack API output is disabled (enabled: false)"},
		},
		{
			name: "Misskey無効ログ確認",
			outputConfig: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					Enabled:         false,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{"Misskey output is disabled (enabled: false)"},
		},
		{
			name: "両方無効ログ確認",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         false,
					APIToken:        "test-token",
					Channel:         "#test",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
				Misskey: &entity.MisskeyConfig{
					Enabled:         false,
					APIToken:        "test-token",
					APIURL:          "https://test.misskey.io",
					MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			expectedLogs: []string{
				"Slack API output is disabled (enabled: false)",
				"Misskey output is disabled (enabled: false)",
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
			stdoutBuffer := new(bytes.Buffer)
			runner, err := NewRecommendRunner(
				mockFetchClient,
				mockRecommender,
				stderrBuffer,
				stdoutBuffer,
				tt.outputConfig,
				&entity.PromptConfig{},
				nil, // cacheConfig
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
	outputConfig := &entity.OutputConfig{
		SlackAPI: &entity.SlackAPIConfig{
			Enabled:         false,
			APIToken:        "test-token",
			Channel:         "#test",
			MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
		},
		Misskey: &entity.MisskeyConfig{
			Enabled:         false,
			APIToken:        "test-token",
			APIURL:          "https://test.misskey.io",
			MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
		},
	}

	runner, err := NewRecommendRunner(
		mockFetchClient,
		mockRecommender,
		stderrBuffer,
		new(bytes.Buffer), // stdout buffer
		outputConfig,
		&entity.PromptConfig{},
		nil, // cacheConfig
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
	profile := &entity.Profile{}
	err = runner.Run(context.Background(), params, profile)

	assert.NoError(t, err) // 全出力先無効でもエラーにならない
}

func TestRecommendRunner_Run_ConfigLogging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// テスト用のslogハンドラーをセットアップしてログ出力をキャプチャ
	var logBuffer bytes.Buffer
	// DEBUGレベルのログをキャプチャするためにLevelをDebugに設定
	handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	originalLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger) // テスト後に元のロガーを復元

	mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
	mockRecommender := mock_domain.NewMockRecommender(ctrl)

	// モックの期待値を設定
	mockFetchClient.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
	mockRecommender.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{Article: entity.Article{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	// テスト用の設定値を作成
	testOutputConfig := &entity.OutputConfig{SlackAPI: &entity.SlackAPIConfig{Enabled: true, APIToken: "slack-token", Channel: "#general", MessageTemplate: toStringP("test-template")}}
	testPromptConfig := &entity.PromptConfig{CommentPromptTemplate: "test-prompt", FixedMessage: "test-fixed-message"}
	testCacheConfig := &entity.CacheConfig{Enabled: false, FilePath: "/tmp/test-cache"}
	testProfile := &entity.Profile{
		AI:     &entity.AIConfig{Gemini: &entity.GeminiConfig{Type: "gemini", APIKey: "gemini-key"}},
		Prompt: testPromptConfig,
		Output: testOutputConfig,
	}

	// NewRecommendRunner の引数として渡す
	runner, err := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, testOutputConfig, testPromptConfig, testCacheConfig)
	require.NoError(t, err)
	require.NotNil(t, runner)

	// runner.viewers をモックに差し替える
	mockViewer := mock_domain.NewMockMessageSender(ctrl)
	mockViewer.EXPECT().SendRecommend(gomock.Any(), gomock.Any()).Return(nil)
	runner.viewers = []domain.MessageSender{mockViewer}

	params := &RecommendParams{URLs: []string{"http://example.com/feed"}}

	// Run メソッドを実行
	err = runner.Run(context.Background(), params, testProfile)
	require.NoError(t, err)

	// ログ出力を検証
	logOutput := logBuffer.String()
	t.Logf("Log output: %s", logOutput)

	// "RecommendRunner.Run parameters" のログエントリを探す
	lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
	var configLogLine []byte
	for _, line := range lines {
		if len(line) > 0 && bytes.Contains(line, []byte("RecommendRunner.Run parameters")) {
			configLogLine = line
			break
		}
	}
	require.NotNil(t, configLogLine, "Config logging line not found")

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(configLogLine, &logEntry))

	assert.Equal(t, "DEBUG", logEntry["level"])
	assert.Equal(t, "RecommendRunner.Run parameters", logEntry["msg"])

	// ログに出力された設定値の検証
	// LogValue()による機密情報のマスク処理が正しく動作することを確認
	profileLog, ok := logEntry["profile"].(map[string]any)
	require.True(t, ok)

	// Gemini APIKeyがマスクされていることを確認
	aiLog, ok := profileLog["AI"].(map[string]any)
	require.True(t, ok)
	geminiLog, ok := aiLog["Gemini"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "[REDACTED]", geminiLog["APIKey"], "Gemini APIKey should be masked")

	// 非機密情報（Type）は正しく出力されることを確認
	assert.Equal(t, testProfile.AI.Gemini.Type, geminiLog["Type"])

	// SlackAPI APITokenがマスクされていることを確認
	outputLog, ok := profileLog["Output"].(map[string]any)
	require.True(t, ok)
	slackLog, ok := outputLog["SlackAPI"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "[REDACTED]", slackLog["APIToken"], "SlackAPI APIToken should be masked")

	// SlackAPIの非機密情報（Channel）は正しく出力されることを確認
	assert.Equal(t, testOutputConfig.SlackAPI.Channel, slackLog["Channel"])
}
