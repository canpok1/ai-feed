package domain

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// TestProfileValidatorImpl_Validate はProfileValidatorImplのValidateメソッドをテストする
func TestProfileValidatorImpl_Validate(t *testing.T) {
	validator := NewProfileValidator()

	tests := []struct {
		name             string
		profile          *entity.Profile
		expectedIsValid  bool
		expectedErrors   []string
		expectedWarnings []string
	}{
		{
			name: "完全に有効なプロファイル",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: "valid-api-key",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "有効なシステムプロンプト",
					CommentPromptTemplate: "有効なコメントプロンプトテンプレート",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "xoxb-valid-token",
						Channel:  "#general",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "valid-misskey-token",
						APIURL:   "https://misskey.social/api",
					},
				},
			},
			expectedIsValid:  true,
			expectedErrors:   []string{},
			expectedWarnings: nil,
		},
		{
			name: "警告のみがある場合",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: "valid-api-key",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "有効なシステムプロンプト",
					CommentPromptTemplate: "有効なコメントプロンプトテンプレート",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "",
						Channel:  "",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "",
						APIURL:   "",
					},
				},
			},
			expectedIsValid: true,
			expectedErrors:  []string{},
			expectedWarnings: []string{
				"Slack API token is not configured",
				"Slack channel is not configured",
				"Misskey API token is not configured",
				"Misskey API URL is not configured",
			},
		},
		{
			name: "必須項目エラーのみの場合",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: "YOUR_GEMINI_API_KEY_HERE",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "",
					CommentPromptTemplate: "",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "xoxb-valid-token",
						Channel:  "#general",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "valid-misskey-token",
						APIURL:   "https://misskey.social/api",
					},
				},
			},
			expectedIsValid: false,
			expectedErrors: []string{
				"Gemini API key is not configured",
				"System prompt is not configured",
				"Comment prompt template is not configured",
			},
			expectedWarnings: nil,
		},
		{
			name: "エラーと警告が混在する場合",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: "YOUR_GEMINI_API_KEY_HERE",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "有効なシステムプロンプト",
					CommentPromptTemplate: "",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "xoxb-YOUR_SLACK_API_TOKEN_HERE",
						Channel:  "",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
						APIURL:   "",
					},
				},
			},
			expectedIsValid: false,
			expectedErrors: []string{
				"Gemini API key is not configured",
				"Comment prompt template is not configured",
			},
			expectedWarnings: []string{
				"Slack API token is not configured",
				"Slack channel is not configured",
				"Misskey API token is not configured",
				"Misskey API URL is not configured",
			},
		},
		{
			name: "nilフィールドがある場合",
			profile: &entity.Profile{
				AI:     nil,
				Prompt: nil,
				Output: nil,
			},
			expectedIsValid: false,
			expectedErrors: []string{
				"Gemini API key is not configured",
				"System prompt is not configured",
				"Comment prompt template is not configured",
			},
			expectedWarnings: []string{
				"Slack API token is not configured",
				"Slack channel is not configured",
				"Misskey API token is not configured",
				"Misskey API URL is not configured",
			},
		},
		{
			name: "部分的にnilフィールドがある場合",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: nil,
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "有効なシステムプロンプト",
					CommentPromptTemplate: "有効なコメントプロンプトテンプレート",
				},
				Output: &entity.OutputConfig{
					SlackAPI: nil,
					Misskey:  nil,
				},
			},
			expectedIsValid: false,
			expectedErrors: []string{
				"Gemini API key is not configured",
			},
			expectedWarnings: []string{
				"Slack API token is not configured",
				"Slack channel is not configured",
				"Misskey API token is not configured",
				"Misskey API URL is not configured",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.profile)

			// IsValidの確認
			assert.Equal(t, tt.expectedIsValid, result.IsValid, "IsValid should match expected value")

			// エラーの確認
			assert.Len(t, result.Errors, len(tt.expectedErrors), "Number of errors should match")
			for _, expectedError := range tt.expectedErrors {
				assert.Contains(t, result.Errors, expectedError, "Should contain expected error: %s", expectedError)
			}

			// 警告の確認
			assert.Len(t, result.Warnings, len(tt.expectedWarnings), "Number of warnings should match")
			for _, expectedWarning := range tt.expectedWarnings {
				assert.Contains(t, result.Warnings, expectedWarning, "Should contain expected warning: %s", expectedWarning)
			}
		})
	}
}

// TestMaskSensitiveData はMaskSensitiveData関数をテストする
func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空文字列",
			input:    "",
			expected: "",
		},
		{
			name:     "Geminiデフォルト値",
			input:    "YOUR_GEMINI_API_KEY_HERE",
			expected: "YOUR_GEMINI_API_KEY_HERE",
		},
		{
			name:     "Slackデフォルト値",
			input:    "xoxb-YOUR_SLACK_API_TOKEN_HERE",
			expected: "xoxb-YOUR_SLACK_API_TOKEN_HERE",
		},
		{
			name:     "Misskeyデフォルト値",
			input:    "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
			expected: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
		},
		{
			name:     "短い実際の値",
			input:    "abc123",
			expected: "******",
		},
		{
			name:     "通常の実際の値",
			input:    "sk-1234567890abcdef",
			expected: "sk-1***********cdef",
		},
		{
			name:     "長い実際の値",
			input:    "very-long-api-key-with-many-characters",
			expected: "very******************************ters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result, "Masked value should match expected")
		})
	}
}

// TestProfileValidatorImpl_validateRequiredFields は必須項目バリデーションをテストする
func TestProfileValidatorImpl_validateRequiredFields(t *testing.T) {
	validator := &ProfileValidatorImpl{}

	tests := []struct {
		name           string
		profile        *entity.Profile
		expectedErrors []string
	}{
		{
			name: "全ての必須項目が有効",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						APIKey: "valid-key",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
			},
			expectedErrors: nil,
		},
		{
			name: "Gemini APIキーが未設定",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						APIKey: "",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
			},
			expectedErrors: []string{"Gemini API key is not configured"},
		},
		{
			name: "Gemini APIキーがデフォルト値",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						APIKey: "YOUR_GEMINI_API_KEY_HERE",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
			},
			expectedErrors: []string{"Gemini API key is not configured"},
		},
		{
			name: "システムプロンプトが未設定",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						APIKey: "valid-key",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "",
					CommentPromptTemplate: "valid template",
				},
			},
			expectedErrors: []string{"System prompt is not configured"},
		},
		{
			name: "コメントプロンプトテンプレートが未設定",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						APIKey: "valid-key",
					},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "",
				},
			},
			expectedErrors: []string{"Comment prompt template is not configured"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateRequiredFields(tt.profile)
			assert.Equal(t, tt.expectedErrors, errors, "Errors should match expected")
		})
	}
}

// TestProfileValidatorImpl_validateWarningFields は警告項目バリデーションをテストする
func TestProfileValidatorImpl_validateWarningFields(t *testing.T) {
	validator := &ProfileValidatorImpl{}

	tests := []struct {
		name             string
		profile          *entity.Profile
		expectedWarnings []string
	}{
		{
			name: "全ての警告項目が有効",
			profile: &entity.Profile{
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "xoxb-valid-token",
						Channel:  "#general",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "valid-token",
						APIURL:   "https://example.com",
					},
				},
			},
			expectedWarnings: nil,
		},
		{
			name: "Slack APIトークンが未設定",
			profile: &entity.Profile{
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "",
						Channel:  "#general",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "valid-token",
						APIURL:   "https://example.com",
					},
				},
			},
			expectedWarnings: []string{"Slack API token is not configured"},
		},
		{
			name: "Slack APIトークンがデフォルト値",
			profile: &entity.Profile{
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "xoxb-YOUR_SLACK_API_TOKEN_HERE",
						Channel:  "#general",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "valid-token",
						APIURL:   "https://example.com",
					},
				},
			},
			expectedWarnings: []string{"Slack API token is not configured"},
		},
		{
			name: "全ての警告項目が未設定",
			profile: &entity.Profile{
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken: "",
						Channel:  "",
					},
					Misskey: &entity.MisskeyConfig{
						APIToken: "",
						APIURL:   "",
					},
				},
			},
			expectedWarnings: []string{
				"Slack API token is not configured",
				"Slack channel is not configured",
				"Misskey API token is not configured",
				"Misskey API URL is not configured",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := validator.validateWarningFields(tt.profile)
			assert.Equal(t, tt.expectedWarnings, warnings, "Warnings should match expected")
		})
	}
}

// TestValidateSlackMessageTemplate はSlackメッセージテンプレートのバリデーション機能をテストする
func TestValidateSlackMessageTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "正常なテンプレート",
			template:    "{{.Article.Title}}\n{{.Article.Link}}",
			expectError: false,
		},
		{
			name:        "デフォルトテンプレート",
			template:    `{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}`,
			expectError: false,
		},
		{
			name:        "空文字列（エラーなし）",
			template:    "",
			expectError: false,
		},
		{
			name:        "空白のみ（エラーなし）",
			template:    "   \n\t  ",
			expectError: false,
		},
		{
			name:        "無効な構文（未閉じの中括弧）",
			template:    "{{.Article.Title",
			expectError: true,
			errorSubstr: "テンプレート構文エラー",
		},
		{
			name:        "無効な構文（不正な関数）",
			template:    "{{.Article.Title | invalidFunc}}",
			expectError: true,
			errorSubstr: "テンプレート構文エラー",
		},
		{
			name:        "複雑な有効テンプレート",
			template:    `タイトル: {{.Article.Title}}{{if .Comment}}\nコメント: {{.Comment}}{{end}}\nURL: {{.Article.Link}}`,
			expectError: false,
		},
		{
			name:        "カスタムテンプレート例",
			template:    "記事: {{.Article.Title}} - {{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlackMessageTemplate(tt.template)

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid template")
				if tt.errorSubstr != "" {
					assert.Contains(t, err.Error(), tt.errorSubstr, "Error message should contain expected substring")
				}
			} else {
				assert.NoError(t, err, "Should not return error for valid template")
			}
		})
	}
}

// TestProfileValidatorImpl_validateRequiredFields_SlackMessageTemplate はSlackメッセージテンプレートの必須項目バリデーションをテストする
func TestProfileValidatorImpl_validateRequiredFields_SlackMessageTemplate(t *testing.T) {
	validator := &ProfileValidatorImpl{}
	validTemplate := "{{.Article.Title}}\n{{.Article.Link}}"
	invalidTemplate := "{{.Article.Title"

	tests := []struct {
		name           string
		profile        *entity.Profile
		expectedErrors []string
	}{
		{
			name: "有効なSlackメッセージテンプレート",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: "valid-key"},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken:        "valid-token",
						Channel:         "#general",
						MessageTemplate: &validTemplate,
					},
				},
			},
			expectedErrors: nil,
		},
		{
			name: "無効なSlackメッセージテンプレート",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: "valid-key"},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken:        "valid-token",
						Channel:         "#general",
						MessageTemplate: &invalidTemplate,
					},
				},
			},
			expectedErrors: []string{"Slack message template is invalid"},
		},
		{
			name: "Slackメッセージテンプレートがnil（エラーなし）",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: "valid-key"},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						APIToken:        "valid-token",
						Channel:         "#general",
						MessageTemplate: nil,
					},
				},
			},
			expectedErrors: nil,
		},
		{
			name: "SlackAPI設定がnil（エラーなし）",
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: "valid-key"},
				},
				Prompt: &entity.PromptConfig{
					SystemPrompt:          "valid prompt",
					CommentPromptTemplate: "valid template",
				},
				Output: &entity.OutputConfig{
					SlackAPI: nil,
				},
			},
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateRequiredFields(tt.profile)

			if tt.expectedErrors == nil {
				assert.Empty(t, errors, "Should not return errors")
			} else {
				assert.NotEmpty(t, errors, "Should return errors")
				for _, expectedError := range tt.expectedErrors {
					found := false
					for _, err := range errors {
						if assert.Contains(t, err, expectedError) {
							found = true
							break
						}
					}
					assert.True(t, found, "Should contain expected error: %s", expectedError)
				}
			}
		})
	}
}
