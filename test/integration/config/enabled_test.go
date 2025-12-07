//go:build integration

// Package config はenabled=false時のバリデーションスキップ統合テストを提供する
package config

import (
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// TestSlackAPI_DisabledSkipsAllValidation はSlack API無効化時に全バリデーションがスキップされることを検証する
// enabled=falseの場合、すべての必須項目を省略しても、また排他チェックに違反してもエラーにならないこと
func TestSlackAPI_DisabledSkipsAllValidation(t *testing.T) {
	iconURL := "https://example.com/icon.png"
	iconEmoji := ":smile:"

	// すべてのバリデーション違反を含む設定
	config := &entity.SlackAPIConfig{
		Enabled:         false,
		APIToken:        entity.SecretString{}, // 空のAPIToken（通常はエラー）
		Channel:         "",                    // 空のChannel（通常はエラー）
		MessageTemplate: nil,                   // nilのMessageTemplate（通常はエラー）
		IconURL:         &iconURL,
		IconEmoji:       &iconEmoji, // icon_urlとicon_emojiの両方設定（通常はエラー）
	}

	result := config.Validate()

	assert.True(t, result.IsValid, "enabled=falseの場合、すべてのバリデーションがスキップされるはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestMisskey_DisabledSkipsAllValidation はMisskey無効化時に全バリデーションがスキップされることを検証する
// enabled=falseの場合、すべての必須項目を省略しても、また不正なURLを指定してもエラーにならないこと
func TestMisskey_DisabledSkipsAllValidation(t *testing.T) {
	// すべてのバリデーション違反を含む設定
	config := &entity.MisskeyConfig{
		Enabled:         false,
		APIToken:        entity.SecretString{}, // 空のAPIToken（通常はエラー）
		APIURL:          "invalid-url-format",  // 不正なURL形式（通常はエラー）
		MessageTemplate: nil,                   // nilのMessageTemplate（通常はエラー）
	}

	result := config.Validate()

	assert.True(t, result.IsValid, "enabled=falseの場合、すべてのバリデーションがスキップされるはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestBothOutputsDisabled_SkipsAllValidation は両出力無効化時に全出力バリデーションがスキップされることを検証する
// SlackAPIとMisskeyの両方がenabled=falseの場合、すべての設定項目を省略してもエラーにならないこと
func TestBothOutputsDisabled_SkipsAllValidation(t *testing.T) {
	iconURL := "https://example.com/icon.png"
	iconEmoji := ":smile:"

	profile := &entity.Profile{
		AI:     NewEntityAIConfig(),
		Prompt: NewEntityPromptConfig(),
		Output: &entity.OutputConfig{
			SlackAPI: &entity.SlackAPIConfig{
				Enabled:         false,
				APIToken:        entity.SecretString{},
				Channel:         "",
				MessageTemplate: nil,
				IconURL:         &iconURL,
				IconEmoji:       &iconEmoji,
			},
			Misskey: &entity.MisskeyConfig{
				Enabled:         false,
				APIToken:        entity.SecretString{},
				APIURL:          "invalid-url-format",
				MessageTemplate: nil,
			},
		},
	}

	result := profile.Validate()

	assert.True(t, result.IsValid, "両方の出力がenabled=falseの場合、すべての設定を省略してもバリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestSlackAPI_EnabledExecutesValidation はSlack API有効化時にバリデーションが正常に実行されることを検証する
// enabled=trueの場合、必須項目が欠けているとバリデーションエラーになること
func TestSlackAPI_EnabledExecutesValidation(t *testing.T) {
	messageTemplate := "{{.Article.Title}}"
	iconURL := "https://example.com/icon.png"
	iconEmoji := ":smile:"

	tests := []struct {
		name           string
		config         *entity.SlackAPIConfig
		expectedErrors []string
	}{
		{
			name: "api_tokenが空の場合エラー",
			config: &entity.SlackAPIConfig{
				Enabled:         true,
				APIToken:        entity.SecretString{}, // 空
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
			expectedErrors: []string{"Slack APIトークンが設定されていません"},
		},
		{
			name: "channelが空の場合エラー",
			config: &entity.SlackAPIConfig{
				Enabled:         true,
				APIToken:        entity.NewSecretString("test-token"),
				Channel:         "", // 空
				MessageTemplate: &messageTemplate,
			},
			expectedErrors: []string{"Slackチャンネルが設定されていません"},
		},
		{
			name: "message_templateがnilの場合エラー",
			config: &entity.SlackAPIConfig{
				Enabled:         true,
				APIToken:        entity.NewSecretString("test-token"),
				Channel:         "#test-channel",
				MessageTemplate: nil, // nil
			},
			expectedErrors: []string{"Slackメッセージテンプレートが設定されていません"},
		},
		{
			name: "icon_urlとicon_emojiが同時に設定された場合エラー",
			config: &entity.SlackAPIConfig{
				Enabled:         true,
				APIToken:        entity.NewSecretString("test-token"),
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
				IconURL:         &iconURL,
				IconEmoji:       &iconEmoji,
			},
			expectedErrors: []string{"icon_urlとicon_emojiを同時に指定することはできません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.False(t, result.IsValid, "enabled=trueの場合、不正な設定ではバリデーションが失敗するはずです")
			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, expectedError) {
						found = true
						break
					}
				}
				assert.True(t, found, "期待されるエラーメッセージ '%s' が含まれているはずです。実際のエラー: %v", expectedError, result.Errors)
			}
		})
	}
}

// TestMisskey_EnabledExecutesValidation はMisskey有効化時にバリデーションが正常に実行されることを検証する
// enabled=trueの場合、必須項目が欠けているとバリデーションエラーになること
func TestMisskey_EnabledExecutesValidation(t *testing.T) {
	messageTemplate := "{{.Article.Title}}"

	tests := []struct {
		name           string
		config         *entity.MisskeyConfig
		expectedErrors []string
	}{
		{
			name: "api_tokenが空の場合エラー",
			config: &entity.MisskeyConfig{
				Enabled:         true,
				APIToken:        entity.SecretString{}, // 空
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &messageTemplate,
			},
			expectedErrors: []string{"Misskey APIトークンが設定されていません"},
		},
		{
			name: "api_urlが不正な形式の場合エラー",
			config: &entity.MisskeyConfig{
				Enabled:         true,
				APIToken:        entity.NewSecretString("test-token"),
				APIURL:          "invalid-url", // 不正なURL
				MessageTemplate: &messageTemplate,
			},
			expectedErrors: []string{"Misskey API URL"},
		},
		{
			name: "message_templateがnilの場合エラー",
			config: &entity.MisskeyConfig{
				Enabled:         true,
				APIToken:        entity.NewSecretString("test-token"),
				APIURL:          "https://misskey.example.com",
				MessageTemplate: nil, // nil
			},
			expectedErrors: []string{"Misskeyメッセージテンプレートが設定されていません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.False(t, result.IsValid, "enabled=trueの場合、不正な設定ではバリデーションが失敗するはずです")
			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, expectedError) {
						found = true
						break
					}
				}
				assert.True(t, found, "期待されるエラーメッセージ '%s' が含まれているはずです。実際のエラー: %v", expectedError, result.Errors)
			}
		})
	}
}
