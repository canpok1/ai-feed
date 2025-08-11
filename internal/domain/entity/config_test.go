package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiConfig_Validate(t *testing.T) {
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
				APIKey: "valid-api-key",
			},
			wantErr: false,
		},
		{
			name: "異常系_Typeが空文字列",
			config: &GeminiConfig{
				Type:   "",
				APIKey: "valid-api-key",
			},
			wantErr: true,
			errors:  []string{"Gemini設定のTypeが設定されていません"},
		},
		{
			name: "異常系_APIKeyが空文字列",
			config: &GeminiConfig{
				Type:   "gemini-pro",
				APIKey: "",
			},
			wantErr: true,
			errors:  []string{"Gemini API keyが設定されていません"},
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
	validAI := &AIConfig{
		Gemini: &GeminiConfig{
			Type:   "gemini-pro",
			APIKey: "valid-api-key",
		},
	}
	validPrompt := &PromptConfig{
		SystemPrompt:          "システムプロンプト",
		CommentPromptTemplate: "コメントテンプレート",
	}
	validOutput := &OutputConfig{
		SlackAPI: &SlackAPIConfig{
			APIToken: "valid-token",
			Channel:  "#general",
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
			name: "正常系_必須項目のみ",
			config: &MisskeyConfig{
				APIToken: "valid-token",
				APIURL:   "https://misskey.example.com",
			},
			wantErr: false,
		},
		{
			name: "正常系_テンプレート付き",
			config: &MisskeyConfig{
				APIToken:        "valid-token",
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &validTemplate,
			},
			wantErr: false,
		},
		{
			name: "正常系_空のテンプレート",
			config: &MisskeyConfig{
				APIToken:        "valid-token",
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &emptyTemplate,
			},
			wantErr: false,
		},
		{
			name: "異常系_APITokenが空",
			config: &MisskeyConfig{
				APIToken: "",
				APIURL:   "https://misskey.example.com",
			},
			wantErr: true,
			errors:  []string{"Misskey APIトークンが設定されていません"},
		},
		{
			name: "異常系_APIURLが空",
			config: &MisskeyConfig{
				APIToken: "valid-token",
				APIURL:   "",
			},
			wantErr: true,
			errors:  []string{"Misskey API URLが設定されていません"},
		},
		{
			name: "異常系_APIURLが不正なURL",
			config: &MisskeyConfig{
				APIToken: "valid-token",
				APIURL:   "not-a-url",
			},
			wantErr: true,
			errors:  []string{"Misskey API URLが正しいURL形式ではありません"},
		},
		{
			name: "異常系_不正なテンプレート構文",
			config: &MisskeyConfig{
				APIToken:        "valid-token",
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &invalidTemplate,
			},
			wantErr: true,
			errors:  []string{"Misskeyメッセージテンプレートが無効です: テンプレート構文エラー: template: misskey_message:1: unclosed action"},
		},
		{
			name: "異常系_複数のエラー",
			config: &MisskeyConfig{
				APIToken:        "",
				APIURL:          "not-a-url",
				MessageTemplate: &invalidTemplate,
			},
			wantErr: true,
			errors: []string{
				"Misskey APIトークンが設定されていません",
				"Misskey API URLが正しいURL形式ではありません",
				"Misskeyメッセージテンプレートが無効です: テンプレート構文エラー: template: misskey_message:1: unclosed action",
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
