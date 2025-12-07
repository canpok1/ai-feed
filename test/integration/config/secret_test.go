//go:build integration

// Package config はAPIキー/トークンの優先順位と環境変数解決の統合テストを提供する
package config

import (
	"os"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Gemini APIキー解決テスト
// ============================================================================

// TestGeminiSecret_DirectAPIKeyPrecedence はapi_keyとapi_key_env両方指定時、api_keyが優先されることを検証する
func TestGeminiSecret_DirectAPIKeyPrecedence(t *testing.T) {
	const envVarName = "TEST_GEMINI_API_KEY_SECRET"
	const envAPIKey = "api-key-from-env"
	const directAPIKey = "direct-api-key"

	err := os.Setenv(envVarName, envAPIKey)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:      "gemini-2.5-flash",
				APIKey:    directAPIKey,
				APIKeyEnv: envVarName,
			},
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	assert.Equal(t, directAPIKey, entityProfile.AI.Gemini.APIKey.Value(),
		"api_keyがapi_key_envより優先されるはずです")
}

// TestGeminiSecret_EnvVarNotSet はapi_key_envで指定した環境変数が未定義の場合のエラーを検証する
func TestGeminiSecret_EnvVarNotSet(t *testing.T) {
	const envVarName = "NONEXISTENT_GEMINI_API_KEY_SECRET"

	_ = os.Unsetenv(envVarName)

	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:      "gemini-2.5-flash",
				APIKey:    "",
				APIKeyEnv: envVarName,
			},
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	_, err := profile.ToEntity()

	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestGeminiSecret_EnvVarEmpty は環境変数が空文字の場合のエラーを検証する
func TestGeminiSecret_EnvVarEmpty(t *testing.T) {
	const envVarName = "TEST_GEMINI_API_KEY_EMPTY"

	err := os.Setenv(envVarName, "")
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:      "gemini-2.5-flash",
				APIKey:    "",
				APIKeyEnv: envVarName,
			},
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	_, err = profile.ToEntity()

	assert.Error(t, err, "環境変数が空文字の場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// ============================================================================
// Slack APIトークン解決テスト
// ============================================================================

// TestSlackSecret_DirectAPITokenPrecedence はapi_tokenとapi_token_env両方指定時、api_tokenが優先されることを検証する
func TestSlackSecret_DirectAPITokenPrecedence(t *testing.T) {
	const envVarName = "TEST_SLACK_API_TOKEN_SECRET"
	const envToken = "token-from-env"
	const directToken = "direct-token"

	err := os.Setenv(envVarName, envToken)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:     &enabled,
				APIToken:    directToken,
				APITokenEnv: envVarName,
				Channel:     "#test-channel",
			},
		},
	}

	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	assert.Equal(t, directToken, entityProfile.Output.SlackAPI.APIToken.Value(),
		"api_tokenがapi_token_envより優先されるはずです")
}

// TestSlackSecret_EnvVarNotSet はapi_token_envで指定した環境変数が未定義の場合のエラーを検証する
func TestSlackSecret_EnvVarNotSet(t *testing.T) {
	const envVarName = "NONEXISTENT_SLACK_API_TOKEN_SECRET"

	_ = os.Unsetenv(envVarName)

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:     &enabled,
				APIToken:    "",
				APITokenEnv: envVarName,
				Channel:     "#test-channel",
			},
		},
	}

	_, err := profile.ToEntity()

	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestSlackSecret_EnvVarEmpty は環境変数が空文字の場合のエラーを検証する
func TestSlackSecret_EnvVarEmpty(t *testing.T) {
	const envVarName = "TEST_SLACK_API_TOKEN_EMPTY"

	err := os.Setenv(envVarName, "")
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				Enabled:     &enabled,
				APIToken:    "",
				APITokenEnv: envVarName,
				Channel:     "#test-channel",
			},
		},
	}

	_, err = profile.ToEntity()

	assert.Error(t, err, "環境変数が空文字の場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// ============================================================================
// Misskey APIトークン解決テスト
// ============================================================================

// TestMisskeySecret_DirectAPITokenPrecedence はapi_tokenとapi_token_env両方指定時、api_tokenが優先されることを検証する
func TestMisskeySecret_DirectAPITokenPrecedence(t *testing.T) {
	const envVarName = "TEST_MISSKEY_API_TOKEN_SECRET"
	const envToken = "token-from-env"
	const directToken = "direct-token"

	err := os.Setenv(envVarName, envToken)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			Misskey: &infra.MisskeyConfig{
				Enabled:     &enabled,
				APIToken:    directToken,
				APITokenEnv: envVarName,
				APIURL:      "http://localhost:3000",
			},
		},
	}

	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	assert.Equal(t, directToken, entityProfile.Output.Misskey.APIToken.Value(),
		"api_tokenがapi_token_envより優先されるはずです")
}

// TestMisskeySecret_EnvVarNotSet はapi_token_envで指定した環境変数が未定義の場合のエラーを検証する
func TestMisskeySecret_EnvVarNotSet(t *testing.T) {
	const envVarName = "NONEXISTENT_MISSKEY_API_TOKEN_SECRET"

	_ = os.Unsetenv(envVarName)

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			Misskey: &infra.MisskeyConfig{
				Enabled:     &enabled,
				APIToken:    "",
				APITokenEnv: envVarName,
				APIURL:      "http://localhost:3000",
			},
		},
	}

	_, err := profile.ToEntity()

	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestMisskeySecret_EnvVarEmpty は環境変数が空文字の場合のエラーを検証する
func TestMisskeySecret_EnvVarEmpty(t *testing.T) {
	const envVarName = "TEST_MISSKEY_API_TOKEN_EMPTY"

	err := os.Setenv(envVarName, "")
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	enabled := true
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			Misskey: &infra.MisskeyConfig{
				Enabled:     &enabled,
				APIToken:    "",
				APITokenEnv: envVarName,
				APIURL:      "http://localhost:3000",
			},
		},
	}

	_, err = profile.ToEntity()

	assert.Error(t, err, "環境変数が空文字の場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}
