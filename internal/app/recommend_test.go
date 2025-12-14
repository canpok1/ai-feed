package app

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
	"github.com/canpok1/ai-feed/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// makeSecretString はテスト用にSecretStringを作成する。
func makeSecretString(value string) entity.SecretString {
	return entity.NewSecretString(value)
}

// testMessageSenderFactory はテスト用のMessageSenderFactoryを返す
func testMessageSenderFactory(outputConfig *entity.OutputConfig) ([]domain.MessageSender, error) {
	// テスト用に空のスライスを返す（Sendersはモックでオーバーライドされることがほとんど）
	return []domain.MessageSender{}, nil
}

// testRecommendCacheFactory はテスト用のRecommendCacheFactoryを返す
func testRecommendCacheFactory(cacheConfig *entity.CacheConfig) (domain.RecommendCache, error) {
	return &mockNopCache{}, nil
}

// mockNopCache はテスト用のノーオペレーションキャッシュ
type mockNopCache struct{}

func (c *mockNopCache) Initialize() error                { return nil }
func (c *mockNopCache) IsCached(url string) bool         { return false }
func (c *mockNopCache) AddEntry(url, title string) error { return nil }
func (c *mockNopCache) Close() error                     { return nil }

// createMockConfig はテスト用にモックのentity.Profileを作成する。
func createMockConfig(promptConfig *entity.PromptConfig, outputConfig *entity.OutputConfig) *entity.Profile {
	return &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{Type: "test-type", APIKey: makeSecretString("test-key")},
		},
		Prompt: promptConfig,
		Output: outputConfig,
	}
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
			name:         "正常系: Sender無しで作成成功",
			outputConfig: &entity.OutputConfig{},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "正常系: SlackAPI Senderで作成成功",
			outputConfig: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:         true,
					APIToken:        makeSecretString("test-token"),
					Channel:         "#test",
					MessageTemplate: testutil.StringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		{
			name: "正常系: Misskey Senderで作成成功",
			outputConfig: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					Enabled:         true,
					APIToken:        makeSecretString("test-token"),
					APIURL:          "https://test.misskey.io/api",
					MessageTemplate: testutil.StringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
				},
			},
			promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
			expectError:  false,
		},
		// 異常系: ファクトリエラーのテストはcmd層で実施
		// runner層ではファクトリがDIされるため、ファクトリ内部のエラーはrunner層のテスト対象外
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
				testMessageSenderFactory,
				testRecommendCacheFactory,
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
				// ファクトリパターン導入により、sendersはファクトリ関数から作成される
				// テスト用ファクトリは空のsendersを返すため、数のチェックはスキップ
				assert.NotNil(t, runner.senders)
			}
		})
	}
}

func TestRecommendRunner_Run(t *testing.T) {
	// デフォルト設定を変数として定義し、複数のテストケースで再利用
	defaultOutputConfig := &entity.OutputConfig{}
	defaultPromptConfig := &entity.PromptConfig{CommentPromptTemplate: "test-prompt-template"}

	// createDefaultProfile はテスト用のデフォルトProfileを作成する。
	createDefaultProfile := func() *entity.Profile {
		return createMockConfig(defaultPromptConfig, defaultOutputConfig)
	}

	tests := []struct {
		name                        string
		mockFetchClientExpectations func(m *mock_domain.MockFetchClient)
		mockRecommenderExpectations func(m *mock_domain.MockRecommender)
		setupProfile                func() *entity.Profile // runner.Run()の引数として使用
		outputConfig                *entity.OutputConfig   // NewRecommendRunnerの引数として使用
		promptConfig                *entity.PromptConfig   // NewRecommendRunnerの引数として使用
		params                      *RecommendParams
		expectedErrorMessage        *string
	}{
		{
			name: "正常系: 推薦成功",
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
			setupProfile: createDefaultProfile,
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: nil,
		},
		{
			name: "異常系: 記事が見つからない",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// 記事が見つからない場合は呼び出されない
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Times(0)
			},
			setupProfile: createDefaultProfile,
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
			params: &RecommendParams{
				URLs: []string{"http://example.com/empty.xml"},
			},
			expectedErrorMessage: testutil.StringPtr("no articles found in the feed"),
		},
		{
			name: "異常系: フェッチエラー",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return(nil, fmt.Errorf("mock fetch error")).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Times(0)
			},
			setupProfile: createDefaultProfile,
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
			params: &RecommendParams{
				URLs: []string{"http://invalid.com/feed.xml"},
			},
			expectedErrorMessage: testutil.StringPtr("no articles found in the feed"),
		},
		{
			name: "異常系: 推薦エラー",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock recommend error")).Times(1)
			},
			setupProfile: createDefaultProfile,
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: testutil.StringPtr("failed to recommend article: mock recommend error"),
		},
		{
			name: "正常系: AIモデル未設定",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommendは呼び出される - 設定はRecommenderのコンストラクタで処理される
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			setupProfile: func() *entity.Profile {
				profile := createDefaultProfile()
				profile.AI = nil
				return profile
			},
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
			params: &RecommendParams{
				URLs: []string{"http://example.com/feed.xml"},
			},
			expectedErrorMessage: nil,
		},
		{
			name: "正常系: プロンプト未設定",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"}}, nil).AnyTimes()
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Recommendは呼び出される - 設定はRecommenderのコンストラクタで処理される
				m.EXPECT().Recommend(gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			setupProfile: func() *entity.Profile {
				profile := createDefaultProfile()
				profile.Prompt = nil
				return profile
			},
			outputConfig: defaultOutputConfig,
			promptConfig: defaultPromptConfig,
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
			stdoutBuffer := new(bytes.Buffer)

			profile := tt.setupProfile()
			runner, runErr := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, tt.outputConfig, tt.promptConfig, nil, testMessageSenderFactory, testRecommendCacheFactory)

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

// TestRecommendRunner_Run_LogOutput は推薦成功時のslogログ出力をテストする
func TestRecommendRunner_Run_LogOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// テスト用のslogハンドラーをセットアップしてログ出力をキャプチャ
	var logBuffer bytes.Buffer
	handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	originalLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger) // テスト後に元のロガーを復元

	mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
	mockRecommender := mock_domain.NewMockRecommender(ctrl)

	stderrBuffer := new(bytes.Buffer)
	mockProfile := createMockConfig(&entity.PromptConfig{CommentPromptTemplate: "test-prompt-template", FixedMessage: "Test Fixed Message"}, &entity.OutputConfig{})

	stdoutBuffer := new(bytes.Buffer)
	runner, runErr := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, mockProfile.Output, mockProfile.Prompt, nil, testMessageSenderFactory, testRecommendCacheFactory)
	assert.NoError(t, runErr)

	// テストデータをセットアップ
	testArticles := []entity.Article{
		{Title: "Test Article", Link: "https://example.com/test"},
	}
	testComment := "This is a test comment"
	testRecommend := &entity.Recommend{
		Article: testArticles[0],
		Comment: &testComment,
	}

	// モックの期待値をセットアップ
	mockFetchClient.EXPECT().Fetch(gomock.Any()).Return(testArticles, nil)
	mockRecommender.EXPECT().Recommend(gomock.Any(), testArticles).Return(testRecommend, nil)

	// テストを実行
	params := &RecommendParams{URLs: []string{"https://example.com/feed"}}
	profile := mockProfile
	err := runner.Run(context.Background(), params, profile)

	assert.NoError(t, err)

	// ログ出力を検証
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
			APIToken:        makeSecretString("test-token"),
			Channel:         "#test",
			MessageTemplate: testutil.StringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
		},
		Misskey: &entity.MisskeyConfig{
			Enabled:         false,
			APIToken:        makeSecretString("test-token"),
			APIURL:          "https://test.misskey.io",
			MessageTemplate: testutil.StringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
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
		testMessageSenderFactory,
		testRecommendCacheFactory,
	)

	assert.NoError(t, err)
	assert.NotNil(t, runner)
	assert.Equal(t, 0, len(runner.senders)) // テスト用ファクトリは空のsendersを返す

	// テストデータをセットアップ
	testArticles := []entity.Article{
		{Title: "Test Article", Link: "https://example.com/test"},
	}
	testRecommend := &entity.Recommend{
		Article: testArticles[0],
	}

	// モックの期待値をセットアップ
	mockFetchClient.EXPECT().Fetch(gomock.Any()).Return(testArticles, nil)
	mockRecommender.EXPECT().Recommend(gomock.Any(), testArticles).Return(testRecommend, nil)

	// テストを実行 - エラーにならないことを確認
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
	testOutputConfig := &entity.OutputConfig{SlackAPI: &entity.SlackAPIConfig{Enabled: true, APIToken: makeSecretString("slack-token"), Channel: "#general", MessageTemplate: testutil.StringPtr("test-template")}}
	testPromptConfig := &entity.PromptConfig{CommentPromptTemplate: "test-prompt", FixedMessage: "test-fixed-message"}
	testCacheConfig := &entity.CacheConfig{Enabled: false, FilePath: "/tmp/test-cache"}
	testProfile := &entity.Profile{
		AI:     &entity.AIConfig{Gemini: &entity.GeminiConfig{Type: "gemini", APIKey: makeSecretString("gemini-key")}},
		Prompt: testPromptConfig,
		Output: testOutputConfig,
	}

	// NewRecommendRunner の引数として渡す
	runner, err := NewRecommendRunner(mockFetchClient, mockRecommender, stderrBuffer, stdoutBuffer, testOutputConfig, testPromptConfig, testCacheConfig, testMessageSenderFactory, testRecommendCacheFactory)
	require.NoError(t, err)
	require.NotNil(t, runner)

	// runner.senders をモックに差し替える
	mockSender := mock_domain.NewMockMessageSender(ctrl)
	mockSender.EXPECT().SendRecommend(gomock.Any(), gomock.Any()).Return(nil)
	mockSender.EXPECT().ServiceName().Return("MockService").AnyTimes()
	runner.senders = []domain.MessageSender{mockSender}

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
