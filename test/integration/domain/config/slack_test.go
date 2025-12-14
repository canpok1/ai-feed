//go:build integration

// Package config はSlack API設定の統合テストを提供する
package config

import (
	"os"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSlackAPIConfig_APITokenRequired はenabled=true時にapi_tokenまたはapi_token_envが必須であることを検証する
// api_tokenとapi_token_envの両方が省略された場合、バリデーションエラーになること
func TestSlackAPIConfig_APITokenRequired(t *testing.T) {
	enabled := true
	messageTemplate := "{{.Article.Title}}"
	// APITokenもAPITokenEnvも設定されていないSlackAPI設定を作成
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "",
				APITokenEnv:     "",
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、APITokenに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "APITokenが空の場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Slack APIトークンが設定されていません",
		"APITokenに関するエラーメッセージが含まれているはずです")
}

// TestSlackAPIConfig_ChannelRequired はenabled=true時にchannelが必須であることを検証する
// channelが省略された場合、バリデーションエラーになること
func TestSlackAPIConfig_ChannelRequired(t *testing.T) {
	enabled := true
	messageTemplate := "{{.Article.Title}}"
	// Channelが設定されていないSlackAPI設定を作成
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "test-slack-token",
				Channel:         "", // 空文字列
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、Channelに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "Channelが空の場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Slackチャンネルが設定されていません",
		"Channelに関するエラーメッセージが含まれているはずです")
}

// TestSlackAPIConfig_MessageTemplateRequired はenabled=true時にmessage_templateが必須であることを検証する
// message_templateが省略された場合、バリデーションエラーになること
func TestSlackAPIConfig_MessageTemplateRequired(t *testing.T) {
	enabled := true
	// MessageTemplateが設定されていないSlackAPI設定を作成
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "test-slack-token",
				Channel:         "#test-channel",
				MessageTemplate: nil, // nilに設定
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、MessageTemplateに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "MessageTemplateがnilの場合、バリデーションは失敗するはずです")
	// エラーメッセージには改行や設定例が含まれているため、部分一致で確認
	foundMessageTemplateError := false
	for _, errMsg := range result.Errors {
		if strings.Contains(errMsg, "Slackメッセージテンプレートが設定されていません") {
			foundMessageTemplateError = true
			break
		}
	}
	assert.True(t, foundMessageTemplateError,
		"MessageTemplateに関するエラーメッセージが含まれているはずです")
}

// TestSlackAPIConfig_APITokenPrecedence はapi_tokenとapi_token_env両方指定時の優先度を検証する
// api_tokenが優先され、api_token_envの環境変数は使用されないこと
func TestSlackAPIConfig_APITokenPrecedence(t *testing.T) {
	// 環境変数にAPIトークンを設定
	const envVarName = "TEST_SLACK_API_TOKEN"
	const envAPIToken = "api-token-from-env"
	const directAPIToken = "direct-api-token"

	t.Setenv(envVarName, envAPIToken)

	enabled := true
	messageTemplate := "{{.Article.Title}}"
	// api_tokenとapi_token_envの両方を設定
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        directAPIToken,
				APITokenEnv:     envVarName,
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// api_tokenが優先されることを確認
	assert.Equal(t, directAPIToken, entityProfile.Output.SlackAPI.APIToken.Value(),
		"api_tokenがapi_token_envより優先されるはずです")
}

// TestSlackAPIConfig_IconURLAndIconEmojiExclusive はicon_urlとicon_emojiの排他制御を検証する
// 両方同時に指定された場合、バリデーションエラーになること
func TestSlackAPIConfig_IconURLAndIconEmojiExclusive(t *testing.T) {
	enabled := true
	messageTemplate := "{{.Article.Title}}"
	iconURL := "https://example.com/icon.png"
	iconEmoji := ":robot_face:"

	// icon_urlとicon_emojiの両方を設定
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "test-slack-token",
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
				IconURL:         &iconURL,
				IconEmoji:       &iconEmoji,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、排他制御に関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "icon_urlとicon_emojiが同時に設定された場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Slack設定エラー: icon_urlとicon_emojiを同時に指定することはできません。",
		"排他制御に関するエラーメッセージが含まれているはずです")
}

// TestSlackAPIConfig_UsernameOptional はusernameがオプショナルであることを検証する
// usernameが省略されても、バリデーションが成功すること
func TestSlackAPIConfig_UsernameOptional(t *testing.T) {
	enabled := true
	messageTemplate := "{{.Article.Title}}"

	// Usernameが設定されていないSlackAPI設定を作成（他の必須項目は設定）
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "test-slack-token",
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
				Username:        nil, // オプショナルなのでnil
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認
	assert.True(t, result.IsValid, "Usernameが省略されてもバリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestSlackAPIConfig_EnabledDefaultTrue はenabled省略時にデフォルトでtrueになることを検証する
// 後方互換性のため、enabled省略時はtrueとして扱われること
func TestSlackAPIConfig_EnabledDefaultTrue(t *testing.T) {
	messageTemplate := "{{.Article.Title}}"

	// Enabledがnilの（省略された）SlackAPI設定を作成
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         nil, // 省略（nil）
				APIToken:        "test-slack-token",
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// Enabledがtrueになることを確認
	assert.True(t, entityProfile.Output.SlackAPI.Enabled,
		"enabled省略時はデフォルトでtrueになるはずです")

	// バリデーションが成功することを確認
	result := entityProfile.Validate()
	assert.True(t, result.IsValid, "正しい設定の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestSlackAPIConfig_DisabledSkipsValidation はenabled=false時に必須フィールドのバリデーションがスキップされることを検証する
// enabled=falseの場合、api_token, channel, message_templateが空でもエラーにならないこと
func TestSlackAPIConfig_DisabledSkipsValidation(t *testing.T) {
	enabled := false

	// enabled=falseで、必須項目がすべて空のSlackAPI設定を作成
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "",  // 空文字列
				Channel:         "",  // 空文字列
				MessageTemplate: nil, // nil
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認（enabled=falseなのでスキップされる）
	assert.True(t, result.IsValid, "enabled=falseの場合、必須フィールドのバリデーションはスキップされるはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestSlackAPIConfig_APITokenFromEnv は環境変数からAPIトークンを取得できることを検証する
// api_token_envで指定した環境変数の値がAPIトークンとして使用されること
func TestSlackAPIConfig_APITokenFromEnv(t *testing.T) {
	// 環境変数にAPIトークンを設定
	const envVarName = "TEST_SLACK_API_TOKEN"
	const envAPIToken = "api-token-from-env"

	t.Setenv(envVarName, envAPIToken)

	enabled := true
	messageTemplate := "{{.Article.Title}}"
	// api_token_envのみを設定
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "",
				APITokenEnv:     envVarName,
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 環境変数の値が使用されることを確認
	assert.Equal(t, envAPIToken, entityProfile.Output.SlackAPI.APIToken.Value(),
		"api_token_envで指定した環境変数の値が使用されるはずです")
}

// TestSlackAPIConfig_EnvVarNotSet は環境変数が設定されていない場合のエラーを検証する
// api_token_envで指定した環境変数が存在しない場合、ToEntity()でエラーになること
func TestSlackAPIConfig_EnvVarNotSet(t *testing.T) {
	// 存在しない環境変数を指定
	const envVarName = "NONEXISTENT_SLACK_API_TOKEN"

	// 環境変数が設定されていないことを確認
	_ = os.Unsetenv(envVarName)

	enabled := true
	messageTemplate := "{{.Article.Title}}"
	// api_token_envのみを設定（api_tokenは空）
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:         &enabled,
				APIToken:        "",
				APITokenEnv:     envVarName,
				Channel:         "#test-channel",
				MessageTemplate: &messageTemplate,
			},
		},
	}

	// infra.Profile から entity.Profile に変換
	_, err := profile.ToEntity()

	// 環境変数が設定されていない場合、エラーが返されることを確認
	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestSlackAPIConfig_ValidConfiguration は正しい設定がentity.Profileに変換・バリデーションできることを検証する
// すべての必須フィールドが正しく設定されている場合、正常に変換・バリデーションが完了すること
func TestSlackAPIConfig_ValidConfiguration(t *testing.T) {
	// 正しいSlackAPI設定を作成
	profile := ValidInfraProfile()

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 変換されたProfileの値を検証
	require.NotNil(t, entityProfile.Output, "Output設定が存在するはずです")
	require.NotNil(t, entityProfile.Output.SlackAPI, "SlackAPI設定が存在するはずです")
	assert.True(t, entityProfile.Output.SlackAPI.Enabled, "Enabledがtrueであるはずです")
	assert.Equal(t, "test-slack-token", entityProfile.Output.SlackAPI.APIToken.Value(),
		"APITokenが正しく変換されるはずです")
	assert.Equal(t, "#test-channel", entityProfile.Output.SlackAPI.Channel,
		"Channelが正しく変換されるはずです")
	require.NotNil(t, entityProfile.Output.SlackAPI.MessageTemplate,
		"MessageTemplateが存在するはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認
	assert.True(t, result.IsValid, "正しい設定の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}
