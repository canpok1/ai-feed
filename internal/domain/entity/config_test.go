package entity

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestLogValue_WithNilFields(t *testing.T) {
	// nilフィールドを含むProfileがログ出力時にエラーにならないことを確認
	var logBuffer bytes.Buffer
	handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	originalLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger)

	t.Run("異常系: AI.Geminiがnil", func(t *testing.T) {
		logBuffer.Reset()
		profileWithNilGemini := &Profile{
			AI:     &AIConfig{Gemini: nil},
			Prompt: &PromptConfig{FixedMessage: "test"},
			Output: &OutputConfig{},
		}
		slog.Debug("test nil gemini", slog.Any("profile", *profileWithNilGemini))
		output := logBuffer.String()
		assert.Contains(t, output, "test nil gemini")
	})

	t.Run("異常系: AIがnil", func(t *testing.T) {
		logBuffer.Reset()
		profileWithNilAI := &Profile{
			AI:     nil,
			Prompt: &PromptConfig{FixedMessage: "test"},
			Output: &OutputConfig{},
		}
		slog.Debug("test nil ai", slog.Any("profile", *profileWithNilAI))
		output := logBuffer.String()
		// ログが正常に出力されることを確認（パニックしないことが重要）
		assert.Contains(t, output, "test nil ai")
	})

	t.Run("異常系: Output.SlackAPIとMisskeyがnil", func(t *testing.T) {
		logBuffer.Reset()
		var apiKey SecretString
		apiKey.UnmarshalText([]byte("key"))
		profileWithNilOutput := &Profile{
			AI:     &AIConfig{Gemini: &GeminiConfig{Type: "test", APIKey: apiKey}},
			Prompt: &PromptConfig{FixedMessage: "test"},
			Output: &OutputConfig{SlackAPI: nil, Misskey: nil},
		}
		slog.Debug("test nil output configs", slog.Any("profile", *profileWithNilOutput))
		output := logBuffer.String()
		assert.Contains(t, output, "test nil output configs")
		// APIKeyがマスクされていることを確認
		assert.Contains(t, output, "[REDACTED]")
		assert.NotContains(t, output, "key")
	})

	t.Run("異常系: Output.SlackAPIがnil", func(t *testing.T) {
		logBuffer.Reset()
		var apiKey SecretString
		apiKey.UnmarshalText([]byte("key"))
		var misskeyToken SecretString
		misskeyToken.UnmarshalText([]byte("misskey-token"))
		messageTemplate := "{{.Article.Title}}"
		profileWithNilSlackAPI := &Profile{
			AI:     &AIConfig{Gemini: &GeminiConfig{Type: "test", APIKey: apiKey}},
			Prompt: &PromptConfig{FixedMessage: "test"},
			Output: &OutputConfig{
				SlackAPI: nil,
				Misskey: &MisskeyConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        misskeyToken,
					APIURL:          "https://misskey.example.com",
					MessageTemplate: &messageTemplate,
				},
			},
		}
		slog.Debug("test nil slackapi", slog.Any("profile", *profileWithNilSlackAPI))
		output := logBuffer.String()
		assert.Contains(t, output, "test nil slackapi")
	})

	t.Run("異常系: Output.Misskeyがnil", func(t *testing.T) {
		logBuffer.Reset()
		var apiKey SecretString
		apiKey.UnmarshalText([]byte("key"))
		var slackToken SecretString
		slackToken.UnmarshalText([]byte("slack-token"))
		messageTemplate := "{{.Article.Title}}"
		profileWithNilMisskey := &Profile{
			AI:     &AIConfig{Gemini: &GeminiConfig{Type: "test", APIKey: apiKey}},
			Prompt: &PromptConfig{FixedMessage: "test"},
			Output: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        slackToken,
					Channel:         "#test",
					MessageTemplate: &messageTemplate,
				},
				Misskey: nil,
			},
		}
		slog.Debug("test nil misskey", slog.Any("profile", *profileWithNilMisskey))
		output := logBuffer.String()
		assert.Contains(t, output, "test nil misskey")
	})
}

func TestGeminiConfig_Validate(t *testing.T) {
	// ヘルパー関数: SecretStringを作成
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	tests := []struct {
		name    string
		config  *GeminiConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_TypeとAPIKeyが適切に設定されている",
			config: &GeminiConfig{
				Type:   "gemini-pro",
				APIKey: makeSecretString("valid-api-key"),
			},
			wantErr: false,
		},
		{
			name: "異常系_Typeが空文字列",
			config: &GeminiConfig{
				Type:   "",
				APIKey: makeSecretString("valid-api-key"),
			},
			wantErr: true,
			errors:  []string{"Gemini設定のTypeが設定されていません"},
		},
		{
			name: "異常系_APIKeyが空文字列",
			config: &GeminiConfig{
				Type:   "gemini-pro",
				APIKey: SecretString{}, // ゼロ値 (空)
			},
			wantErr: true,
			errors:  []string{"Gemini APIキーが設定されていません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestProfile_Validate(t *testing.T) {
	// ヘルパー関数: SecretStringを作成
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	validAI := &AIConfig{
		Gemini: &GeminiConfig{
			Type:   "gemini-pro",
			APIKey: makeSecretString("valid-api-key"),
		},
	}
	validPrompt := &PromptConfig{
		SystemPrompt:          "システムプロンプト",
		CommentPromptTemplate: "コメントテンプレート",
		SelectorPrompt:        "記事選択プロンプト",
	}
	validTemplate := "{{.Article.Title}} {{.Article.Link}}"
	validOutput := &OutputConfig{
		SlackAPI: &SlackAPIConfig{
			Enabled:         testutil.BoolPtr(true),
			APIToken:        makeSecretString("valid-token"),
			Channel:         "#general",
			MessageTemplate: &validTemplate,
		},
	}

	tests := []struct {
		name    string
		profile *Profile
		wantErr bool
	}{
		{
			name: "正常系_すべての設定が適切",
			profile: &Profile{
				AI:     validAI,
				Prompt: validPrompt,
				Output: validOutput,
			},
			wantErr: false,
		},
		{
			name: "異常系_AI設定がnil",
			profile: &Profile{
				AI:     nil,
				Prompt: validPrompt,
				Output: validOutput,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.Validate()
			assert.Equal(t, !tt.wantErr, result.IsValid)
		})
	}
}

func TestMisskeyConfig_Validate(t *testing.T) {
	// ヘルパー関数: SecretStringを作成
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	validTemplate := "{{.Article.Title}} {{.Article.Link}}"
	invalidTemplate := "{{.Article.Title" // 不正な構文
	emptyTemplate := ""

	tests := []struct {
		name    string
		config  *MisskeyConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_必須項目すべて",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &validTemplate,
			},
			wantErr: false,
		},
		{
			name: "正常系_テンプレート付き",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &validTemplate,
			},
			wantErr: false,
		},
		{
			name: "異常系_MessageTemplateが未設定",
			config: &MisskeyConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: makeSecretString("valid-token"),
				APIURL:   "https://misskey.example.com",
			},
			wantErr: true,
			errors:  []string{"Misskeyメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nmisskey:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}"},
		},
		{
			name: "異常系_MessageTemplateが空文字列",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &emptyTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskeyメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nmisskey:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}"},
		},
		{
			name: "異常系_APITokenが空",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        SecretString{}, // ゼロ値 (空)
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &validTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskey APIトークンが設定されていません"},
		},
		{
			name: "異常系_APIURLが空",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "",
				MessageTemplate: &validTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskey API URLが設定されていません"},
		},
		{
			name: "異常系_APIURLが不正なURL",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "not-a-url",
				MessageTemplate: &validTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskey API URLが正しいURL形式ではありません"},
		},
		{
			name: "異常系_不正なテンプレート構文",
			config: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("valid-token"),
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &invalidTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskeyメッセージテンプレートが無効です: テンプレート構文エラー: template: misskey_message:1: unclosed action"},
		},
		{
			name: "異常系_複数のエラー",
			config: &MisskeyConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: SecretString{}, // ゼロ値 (空)
				APIURL:   "not-a-url",
			},
			wantErr: true,
			errors: []string{
				"Misskey APIトークンが設定されていません",
				"Misskey API URLが正しいURL形式ではありません",
				"Misskeyメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nmisskey:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestPromptConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PromptConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_すべてのフィールドが設定されている",
			config: &PromptConfig{
				SystemPrompt:          "システムプロンプト",
				CommentPromptTemplate: "コメントテンプレート",
				SelectorPrompt:        "記事選択プロンプト",
				FixedMessage:          "固定メッセージ",
			},
			wantErr: false,
		},
		{
			name: "正常系_FixedMessageは任意項目",
			config: &PromptConfig{
				SystemPrompt:          "システムプロンプト",
				CommentPromptTemplate: "コメントテンプレート",
				SelectorPrompt:        "記事選択プロンプト",
				FixedMessage:          "",
			},
			wantErr: false,
		},
		{
			name: "異常系_SystemPromptが空文字列",
			config: &PromptConfig{
				SystemPrompt:          "",
				CommentPromptTemplate: "コメントテンプレート",
				SelectorPrompt:        "記事選択プロンプト",
			},
			wantErr: true,
			errors:  []string{"システムプロンプトが設定されていません"},
		},
		{
			name: "異常系_CommentPromptTemplateが空文字列",
			config: &PromptConfig{
				SystemPrompt:          "システムプロンプト",
				CommentPromptTemplate: "",
				SelectorPrompt:        "記事選択プロンプト",
			},
			wantErr: true,
			errors:  []string{"コメントプロンプトテンプレートが設定されていません"},
		},
		{
			name: "異常系_SelectorPromptが空文字列",
			config: &PromptConfig{
				SystemPrompt:          "システムプロンプト",
				CommentPromptTemplate: "コメントテンプレート",
				SelectorPrompt:        "",
			},
			wantErr: true,
			errors:  []string{"記事選択プロンプトが設定されていません"},
		},
		{
			name: "異常系_複数のエラー",
			config: &PromptConfig{
				SystemPrompt:          "",
				CommentPromptTemplate: "",
				SelectorPrompt:        "",
			},
			wantErr: true,
			errors: []string{
				"システムプロンプトが設定されていません",
				"コメントプロンプトテンプレートが設定されていません",
				"記事選択プロンプトが設定されていません",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestIsValidMockSelectorMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want bool
	}{
		{
			name: "正常系_first",
			mode: "first",
			want: true,
		},
		{
			name: "正常系_random",
			mode: "random",
			want: true,
		},
		{
			name: "正常系_last",
			mode: "last",
			want: true,
		},
		{
			name: "異常系_空文字列",
			mode: "",
			want: false,
		},
		{
			name: "異常系_無効なモード",
			mode: "invalid",
			want: false,
		},
		{
			name: "異常系_大文字のFIRST",
			mode: "FIRST",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMockSelectorMode(tt.mode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMockConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *MockConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_Enabled_false_バリデーションスキップ",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(false),
				SelectorMode: "",
				Comment:      "",
			},
			wantErr: false,
		},
		{
			name: "正常系_有効なSelectorMode_first",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "first",
				Comment:      "テストコメント",
			},
			wantErr: false,
		},
		{
			name: "正常系_有効なSelectorMode_random",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "random",
				Comment:      "",
			},
			wantErr: false,
		},
		{
			name: "正常系_有効なSelectorMode_last",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "last",
				Comment:      "last comment",
			},
			wantErr: false,
		},
		{
			name: "異常系_無効なSelectorMode",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "invalid",
				Comment:      "",
			},
			wantErr: true,
			errors:  []string{"Mockの記事選択モードが不正です。first, random, lastのいずれかを指定してください"},
		},
		{
			name: "異常系_SelectorModeが空文字列",
			config: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "",
				Comment:      "",
			},
			wantErr: true,
			errors:  []string{"Mockの記事選択モードが不正です。first, random, lastのいずれかを指定してください"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestMockConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		target   *MockConfig
		source   *MockConfig
		expected *MockConfig
	}{
		{
			name: "正常系_nilをマージ",
			target: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "first",
				Comment:      "original",
			},
			source: nil,
			expected: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "first",
				Comment:      "original",
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &MockConfig{
				Enabled:      testutil.BoolPtr(false),
				SelectorMode: "first",
				Comment:      "original",
			},
			source: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "random",
				Comment:      "new comment",
			},
			expected: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "random",
				Comment:      "new comment",
			},
		},
		{
			name: "正常系_空文字列はマージしない",
			target: &MockConfig{
				Enabled:      testutil.BoolPtr(true),
				SelectorMode: "first",
				Comment:      "original",
			},
			source: &MockConfig{
				Enabled:      testutil.BoolPtr(false),
				SelectorMode: "",
				Comment:      "",
			},
			expected: &MockConfig{
				Enabled:      testutil.BoolPtr(false),
				SelectorMode: "first",
				Comment:      "original",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestCacheConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *CacheConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_FilePathが設定されている",
			config: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/path/to/cache.json",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			wantErr: false,
		},
		{
			name: "正常系_MaxEntriesとRetentionDaysがゼロでも有効",
			config: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/path/to/cache.json",
				MaxEntries:    0,
				RetentionDays: 0,
			},
			wantErr: false,
		},
		{
			name: "正常系_Enabledがfalseでも有効",
			config: &CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "/path/to/cache.json",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			wantErr: false,
		},
		{
			name: "異常系_FilePathが空文字列",
			config: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			wantErr: true,
			errors:  []string{"キャッシュファイルパスが設定されていません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestCacheConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		target   *CacheConfig
		source   *CacheConfig
		expected *CacheConfig
	}{
		{
			name: "正常系_nilをマージ",
			target: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			source: nil,
			expected: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			source: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/new/path",
				MaxEntries:    200,
				RetentionDays: 60,
			},
			expected: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/new/path",
				MaxEntries:    200,
				RetentionDays: 60,
			},
		},
		{
			name: "正常系_空文字列とゼロ値はマージしない",
			target: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			source: &CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "",
				MaxEntries:    0,
				RetentionDays: 0,
			},
			expected: &CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
		},
		{
			name: "正常系_部分的な上書き",
			target: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/original/path",
				MaxEntries:    100,
				RetentionDays: 30,
			},
			source: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/new/path",
				MaxEntries:    0,
				RetentionDays: 60,
			},
			expected: &CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/new/path",
				MaxEntries:    100,
				RetentionDays: 60,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestAIConfig_Validate(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	tests := []struct {
		name    string
		config  *AIConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_Gemini設定が有効",
			config: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "gemini-pro",
					APIKey: makeSecretString("valid-key"),
				},
			},
			wantErr: false,
		},
		{
			name: "正常系_Mock設定が有効な場合Gemini不要",
			config: &AIConfig{
				Mock: &MockConfig{
					Enabled:      testutil.BoolPtr(true),
					SelectorMode: "first",
				},
				Gemini: nil,
			},
			wantErr: false,
		},
		{
			name: "正常系_MockとGemini両方設定_Mock有効",
			config: &AIConfig{
				Mock: &MockConfig{
					Enabled:      testutil.BoolPtr(true),
					SelectorMode: "random",
				},
				Gemini: &GeminiConfig{
					Type:   "gemini-pro",
					APIKey: makeSecretString("valid-key"),
				},
			},
			wantErr: false,
		},
		{
			name: "正常系_Mock無効の場合Gemini必須",
			config: &AIConfig{
				Mock: &MockConfig{
					Enabled:      testutil.BoolPtr(false),
					SelectorMode: "first",
				},
				Gemini: &GeminiConfig{
					Type:   "gemini-pro",
					APIKey: makeSecretString("valid-key"),
				},
			},
			wantErr: false,
		},
		{
			name: "異常系_Gemini設定がnil",
			config: &AIConfig{
				Gemini: nil,
			},
			wantErr: true,
			errors:  []string{"Gemini設定が設定されていません"},
		},
		{
			name: "異常系_Mock無効でGeminiがnil",
			config: &AIConfig{
				Mock: &MockConfig{
					Enabled: testutil.BoolPtr(false),
				},
				Gemini: nil,
			},
			wantErr: true,
			errors:  []string{"Gemini設定が設定されていません"},
		},
		{
			name: "異常系_Mock有効だがSelectorModeが無効",
			config: &AIConfig{
				Mock: &MockConfig{
					Enabled:      testutil.BoolPtr(true),
					SelectorMode: "invalid",
				},
				Gemini: nil,
			},
			wantErr: true,
			errors:  []string{"Mockの記事選択モードが不正です。first, random, lastのいずれかを指定してください"},
		},
		{
			name: "異常系_Gemini設定のType空",
			config: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "",
					APIKey: makeSecretString("valid-key"),
				},
			},
			wantErr: true,
			errors:  []string{"Gemini設定のTypeが設定されていません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestAIConfig_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	tests := []struct {
		name     string
		target   *AIConfig
		source   *AIConfig
		validate func(t *testing.T, result *AIConfig)
	}{
		{
			name: "正常系_nilをマージ",
			target: &AIConfig{
				Gemini: &GeminiConfig{Type: "original", APIKey: makeSecretString("original-key")},
			},
			source: nil,
			validate: func(t *testing.T, result *AIConfig) {
				assert.Equal(t, "original", result.Gemini.Type)
			},
		},
		{
			name: "正常系_Geminiをマージ",
			target: &AIConfig{
				Gemini: &GeminiConfig{Type: "original", APIKey: makeSecretString("original-key")},
			},
			source: &AIConfig{
				Gemini: &GeminiConfig{Type: "new-type", APIKey: makeSecretString("new-key")},
			},
			validate: func(t *testing.T, result *AIConfig) {
				assert.Equal(t, "new-type", result.Gemini.Type)
			},
		},
		{
			name: "正常系_Mockをマージ",
			target: &AIConfig{
				Gemini: &GeminiConfig{Type: "original", APIKey: makeSecretString("original-key")},
			},
			source: &AIConfig{
				Mock: &MockConfig{Enabled: testutil.BoolPtr(true), SelectorMode: "first"},
			},
			validate: func(t *testing.T, result *AIConfig) {
				assert.NotNil(t, result.Mock)
				assert.NotNil(t, result.Mock.Enabled)
				assert.True(t, *result.Mock.Enabled)
				assert.Equal(t, "first", result.Mock.SelectorMode)
			},
		},
		{
			name:   "正常系_nilのtargetにマージ",
			target: &AIConfig{},
			source: &AIConfig{
				Gemini: &GeminiConfig{Type: "new-type", APIKey: makeSecretString("new-key")},
			},
			validate: func(t *testing.T, result *AIConfig) {
				assert.NotNil(t, result.Gemini)
				assert.Equal(t, "new-type", result.Gemini.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			tt.validate(t, tt.target)
		})
	}
}

func TestGeminiConfig_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	tests := []struct {
		name     string
		target   *GeminiConfig
		source   *GeminiConfig
		expected *GeminiConfig
	}{
		{
			name: "正常系_nilをマージ",
			target: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
			source: nil,
			expected: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
			source: &GeminiConfig{
				Type:   "new-type",
				APIKey: makeSecretString("new-key"),
			},
			expected: &GeminiConfig{
				Type:   "new-type",
				APIKey: makeSecretString("new-key"),
			},
		},
		{
			name: "正常系_空文字列はマージしない",
			target: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
			source: &GeminiConfig{
				Type:   "",
				APIKey: SecretString{},
			},
			expected: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
		},
		{
			name: "正常系_部分的な上書き",
			target: &GeminiConfig{
				Type:   "original",
				APIKey: makeSecretString("original-key"),
			},
			source: &GeminiConfig{
				Type:   "new-type",
				APIKey: SecretString{},
			},
			expected: &GeminiConfig{
				Type:   "new-type",
				APIKey: makeSecretString("original-key"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestPromptConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		target   *PromptConfig
		source   *PromptConfig
		expected *PromptConfig
	}{
		{
			name: "正常系_nilをマージ",
			target: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
			source: nil,
			expected: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
			source: &PromptConfig{
				SystemPrompt:          "new system",
				CommentPromptTemplate: "new comment",
				SelectorPrompt:        "new selector",
				FixedMessage:          "new fixed",
			},
			expected: &PromptConfig{
				SystemPrompt:          "new system",
				CommentPromptTemplate: "new comment",
				SelectorPrompt:        "new selector",
				FixedMessage:          "new fixed",
			},
		},
		{
			name: "正常系_空文字列はマージしない",
			target: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
			source: &PromptConfig{
				SystemPrompt:          "",
				CommentPromptTemplate: "",
				SelectorPrompt:        "",
				FixedMessage:          "",
			},
			expected: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
		},
		{
			name: "正常系_部分的な上書き",
			target: &PromptConfig{
				SystemPrompt:          "original system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "original selector",
				FixedMessage:          "original fixed",
			},
			source: &PromptConfig{
				SystemPrompt:          "new system",
				CommentPromptTemplate: "",
				SelectorPrompt:        "new selector",
				FixedMessage:          "",
			},
			expected: &PromptConfig{
				SystemPrompt:          "new system",
				CommentPromptTemplate: "original comment",
				SelectorPrompt:        "new selector",
				FixedMessage:          "original fixed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestPromptConfig_BuildCommentPrompt(t *testing.T) {
	tests := []struct {
		name     string
		config   *PromptConfig
		article  *Article
		expected string
		wantErr  bool
	}{
		{
			name: "正常系_新形式テンプレート",
			config: &PromptConfig{
				CommentPromptTemplate: "Title: {{.Title}}, Link: {{.Link}}",
			},
			article: &Article{
				Title: "Test Article",
				Link:  "https://example.com",
			},
			expected: "Title: Test Article, Link: https://example.com",
			wantErr:  false,
		},
		{
			name: "正常系_旧形式テンプレート_title",
			config: &PromptConfig{
				CommentPromptTemplate: "Title: {{title}}",
			},
			article: &Article{
				Title: "Test Article",
			},
			expected: "Title: Test Article",
			wantErr:  false,
		},
		{
			name: "正常系_旧形式テンプレート_url",
			config: &PromptConfig{
				CommentPromptTemplate: "URL: {{url}}",
			},
			article: &Article{
				Link: "https://example.com",
			},
			expected: "URL: https://example.com",
			wantErr:  false,
		},
		{
			name: "正常系_旧形式テンプレート_content",
			config: &PromptConfig{
				CommentPromptTemplate: "Content: {{content}}",
			},
			article: &Article{
				Content: "Article content here",
			},
			expected: "Content: Article content here",
			wantErr:  false,
		},
		{
			name: "正常系_複合テンプレート",
			config: &PromptConfig{
				CommentPromptTemplate: "{{.Title}} - {{.Link}} - {{.Content}}",
			},
			article: &Article{
				Title:   "Test",
				Link:    "https://test.com",
				Content: "Some content",
			},
			expected: "Test - https://test.com - Some content",
			wantErr:  false,
		},
		{
			name: "異常系_不正なテンプレート構文",
			config: &PromptConfig{
				CommentPromptTemplate: "{{.Title",
			},
			article: &Article{
				Title: "Test",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "正常系_空のフィールド",
			config: &PromptConfig{
				CommentPromptTemplate: "Title: {{.Title}}, Link: {{.Link}}",
			},
			article: &Article{
				Title: "",
				Link:  "",
			},
			expected: "Title: , Link: ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.BuildCommentPrompt(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSlackAPIConfig_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	template1 := "template1"
	template2 := "template2"
	username := "bot"
	iconURL := "https://example.com/icon.png"
	iconEmoji := ":robot:"

	tests := []struct {
		name     string
		target   *SlackAPIConfig
		source   *SlackAPIConfig
		validate func(t *testing.T, result *SlackAPIConfig)
	}{
		{
			name: "正常系_nilをマージ",
			target: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("original-token"),
				Channel:         "#original",
				MessageTemplate: &template1,
			},
			source: nil,
			validate: func(t *testing.T, result *SlackAPIConfig) {
				assert.NotNil(t, result.Enabled)
				assert.True(t, *result.Enabled)
				assert.Equal(t, "#original", result.Channel)
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(false),
				APIToken:        makeSecretString("original-token"),
				Channel:         "#original",
				MessageTemplate: &template1,
			},
			source: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("new-token"),
				Channel:         "#new",
				MessageTemplate: &template2,
				Username:        &username,
				IconURL:         &iconURL,
				IconEmoji:       &iconEmoji,
			},
			validate: func(t *testing.T, result *SlackAPIConfig) {
				assert.NotNil(t, result.Enabled)
				assert.True(t, *result.Enabled)
				assert.Equal(t, "#new", result.Channel)
				assert.Equal(t, "template2", *result.MessageTemplate)
				assert.Equal(t, "bot", *result.Username)
				assert.Equal(t, "https://example.com/icon.png", *result.IconURL)
				assert.Equal(t, ":robot:", *result.IconEmoji)
			},
		},
		{
			name: "正常系_空文字列はマージしない",
			target: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("original-token"),
				Channel:         "#original",
				MessageTemplate: &template1,
			},
			source: &SlackAPIConfig{
				Enabled:  testutil.BoolPtr(false),
				APIToken: SecretString{},
				Channel:  "",
			},
			validate: func(t *testing.T, result *SlackAPIConfig) {
				assert.NotNil(t, result.Enabled)
				assert.False(t, *result.Enabled)
				assert.Equal(t, "#original", result.Channel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			tt.validate(t, tt.target)
		})
	}
}

func TestMisskeyConfig_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	template1 := "template1"
	template2 := "template2"

	tests := []struct {
		name     string
		target   *MisskeyConfig
		source   *MisskeyConfig
		validate func(t *testing.T, result *MisskeyConfig)
	}{
		{
			name: "正常系_nilをマージ",
			target: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("original-token"),
				APIURL:          "https://original.example.com",
				MessageTemplate: &template1,
			},
			source: nil,
			validate: func(t *testing.T, result *MisskeyConfig) {
				assert.NotNil(t, result.Enabled)
				assert.True(t, *result.Enabled)
				assert.Equal(t, "https://original.example.com", result.APIURL)
			},
		},
		{
			name: "正常系_全フィールドを上書き",
			target: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(false),
				APIToken:        makeSecretString("original-token"),
				APIURL:          "https://original.example.com",
				MessageTemplate: &template1,
			},
			source: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("new-token"),
				APIURL:          "https://new.example.com",
				MessageTemplate: &template2,
			},
			validate: func(t *testing.T, result *MisskeyConfig) {
				assert.NotNil(t, result.Enabled)
				assert.True(t, *result.Enabled)
				assert.Equal(t, "https://new.example.com", result.APIURL)
				assert.Equal(t, "template2", *result.MessageTemplate)
			},
		},
		{
			name: "正常系_空文字列はマージしない",
			target: &MisskeyConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("original-token"),
				APIURL:          "https://original.example.com",
				MessageTemplate: &template1,
			},
			source: &MisskeyConfig{
				Enabled:  testutil.BoolPtr(false),
				APIToken: SecretString{},
				APIURL:   "",
			},
			validate: func(t *testing.T, result *MisskeyConfig) {
				assert.NotNil(t, result.Enabled)
				assert.False(t, *result.Enabled)
				assert.Equal(t, "https://original.example.com", result.APIURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			tt.validate(t, tt.target)
		})
	}
}

func TestOutputConfig_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	template := "{{.Article.Title}}"

	tests := []struct {
		name     string
		target   *OutputConfig
		source   *OutputConfig
		validate func(t *testing.T, result *OutputConfig)
	}{
		{
			name: "正常系_nilをマージ",
			target: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        makeSecretString("token"),
					Channel:         "#test",
					MessageTemplate: &template,
				},
			},
			source: nil,
			validate: func(t *testing.T, result *OutputConfig) {
				assert.NotNil(t, result.SlackAPI)
				assert.Equal(t, "#test", result.SlackAPI.Channel)
			},
		},
		{
			name:   "正常系_SlackAPIをマージ",
			target: &OutputConfig{},
			source: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        makeSecretString("token"),
					Channel:         "#new",
					MessageTemplate: &template,
				},
			},
			validate: func(t *testing.T, result *OutputConfig) {
				assert.NotNil(t, result.SlackAPI)
				assert.Equal(t, "#new", result.SlackAPI.Channel)
			},
		},
		{
			name:   "正常系_Misskeyをマージ",
			target: &OutputConfig{},
			source: &OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        makeSecretString("token"),
					APIURL:          "https://misskey.example.com",
					MessageTemplate: &template,
				},
			},
			validate: func(t *testing.T, result *OutputConfig) {
				assert.NotNil(t, result.Misskey)
				assert.Equal(t, "https://misskey.example.com", result.Misskey.APIURL)
			},
		},
		{
			name:   "正常系_両方をマージ",
			target: &OutputConfig{},
			source: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        makeSecretString("slack-token"),
					Channel:         "#slack",
					MessageTemplate: &template,
				},
				Misskey: &MisskeyConfig{
					Enabled:         testutil.BoolPtr(true),
					APIToken:        makeSecretString("misskey-token"),
					APIURL:          "https://misskey.example.com",
					MessageTemplate: &template,
				},
			},
			validate: func(t *testing.T, result *OutputConfig) {
				assert.NotNil(t, result.SlackAPI)
				assert.NotNil(t, result.Misskey)
				assert.Equal(t, "#slack", result.SlackAPI.Channel)
				assert.Equal(t, "https://misskey.example.com", result.Misskey.APIURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			tt.validate(t, tt.target)
		})
	}
}

func TestProfile_Merge(t *testing.T) {
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	template := "{{.Article.Title}}"

	tests := []struct {
		name     string
		target   *Profile
		source   *Profile
		validate func(t *testing.T, result *Profile)
	}{
		{
			name: "正常系_nilをマージ",
			target: &Profile{
				AI: &AIConfig{
					Gemini: &GeminiConfig{Type: "original", APIKey: makeSecretString("key")},
				},
				Prompt: &PromptConfig{SystemPrompt: "original"},
				Output: &OutputConfig{},
			},
			source: nil,
			validate: func(t *testing.T, result *Profile) {
				assert.Equal(t, "original", result.AI.Gemini.Type)
				assert.Equal(t, "original", result.Prompt.SystemPrompt)
			},
		},
		{
			name:   "正常系_AIをマージ",
			target: &Profile{},
			source: &Profile{
				AI: &AIConfig{
					Gemini: &GeminiConfig{Type: "new-type", APIKey: makeSecretString("new-key")},
				},
			},
			validate: func(t *testing.T, result *Profile) {
				assert.NotNil(t, result.AI)
				assert.Equal(t, "new-type", result.AI.Gemini.Type)
			},
		},
		{
			name: "正常系_全フィールドをマージ",
			target: &Profile{
				AI: &AIConfig{
					Gemini: &GeminiConfig{Type: "original", APIKey: makeSecretString("original-key")},
				},
				Prompt: &PromptConfig{SystemPrompt: "original system"},
				Output: &OutputConfig{},
			},
			source: &Profile{
				AI: &AIConfig{
					Gemini: &GeminiConfig{Type: "new-type", APIKey: makeSecretString("new-key")},
				},
				Prompt: &PromptConfig{SystemPrompt: "new system"},
				Output: &OutputConfig{
					SlackAPI: &SlackAPIConfig{
						Enabled:         testutil.BoolPtr(true),
						APIToken:        makeSecretString("token"),
						Channel:         "#test",
						MessageTemplate: &template,
					},
				},
			},
			validate: func(t *testing.T, result *Profile) {
				assert.Equal(t, "new-type", result.AI.Gemini.Type)
				assert.Equal(t, "new system", result.Prompt.SystemPrompt)
				assert.NotNil(t, result.Output.SlackAPI)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Merge(tt.source)
			tt.validate(t, tt.target)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
