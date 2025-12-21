package entity

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestSlackAPIConfig_Validate はSlackAPIConfigのValidateメソッドをテストする
func TestSlackAPIConfig_Validate(t *testing.T) {
	// ヘルパー関数: SecretStringを作成
	makeSecretString := func(value string) SecretString {
		return NewSecretString(value)
	}

	validTemplate := "{{.Article.Title}} {{.Article.Link}}"
	invalidTemplate := "{{.Article.Title" // 不正な構文
	emptyTemplate := ""

	tests := []struct {
		name    string
		config  *SlackAPIConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_必須項目すべて",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &validTemplate,
			},
			wantErr: false,
		},
		{
			name: "異常系_MessageTemplateが未設定",
			config: &SlackAPIConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: makeSecretString("xoxb-valid-token"),
				Channel:  "#test",
			},
			wantErr: true,
			errors:  []string{"Slackメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nslack_api:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}"},
		},
		{
			name: "異常系_MessageTemplateが空文字列",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &emptyTemplate,
			},
			wantErr: true,
			errors:  []string{"Slackメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nslack_api:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}"},
		},
		{
			name: "異常系_APITokenが空",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        SecretString{}, // ゼロ値 (空)
				Channel:         "#test",
				MessageTemplate: &validTemplate,
			},
			wantErr: true,
			errors:  []string{"Slack APIトークンが設定されていません"},
		},
		{
			name: "異常系_Channelが空",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "",
				MessageTemplate: &validTemplate,
			},
			wantErr: true,
			errors:  []string{"Slackチャンネルが設定されていません"},
		},
		{
			name: "異常系_不正なテンプレート構文",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &invalidTemplate,
			},
			wantErr: true,
			errors:  []string{"Slackメッセージテンプレートが無効です: テンプレート構文エラー: template: slack_message:1: unclosed action"},
		},
		{
			name: "異常系_複数のエラー",
			config: &SlackAPIConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: SecretString{}, // ゼロ値 (空)
				Channel:  "",
			},
			wantErr: true,
			errors: []string{
				"Slack APIトークンが設定されていません",
				"Slackチャンネルが設定されていません",
				"Slackメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nslack_api:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}",
			},
		},
		{
			name: "正常系_Enabled_false",
			config: &SlackAPIConfig{
				Enabled: testutil.BoolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "異常系_IconURLとIconEmojiが両方設定",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &validTemplate,
				IconURL:         stringPtr("https://example.com/icon.png"),
				IconEmoji:       stringPtr(":robot:"),
			},
			wantErr: true,
			errors:  []string{"Slack設定エラー: icon_urlとicon_emojiを同時に指定することはできません。"},
		},
		{
			name: "正常系_IconURLのみ設定",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &validTemplate,
				IconURL:         stringPtr("https://example.com/icon.png"),
			},
			wantErr: false,
		},
		{
			name: "正常系_IconEmojiのみ設定",
			config: &SlackAPIConfig{
				Enabled:         testutil.BoolPtr(true),
				APIToken:        makeSecretString("xoxb-valid-token"),
				Channel:         "#test",
				MessageTemplate: &validTemplate,
				IconEmoji:       stringPtr(":robot:"),
			},
			wantErr: false,
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
