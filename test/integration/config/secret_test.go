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

// boolPtr はbool値へのポインタを返すヘルパー関数
func boolPtr(b bool) *bool {
	return &b
}

// ============================================================================
// Gemini APIキー解決テスト
// ============================================================================

// TestGeminiSecret_SecretResolution はGemini APIキーの解決ロジックを検証する
func TestGeminiSecret_SecretResolution(t *testing.T) {
	tests := []struct {
		name           string
		envVarName     string
		envVarValue    *string // nilの場合は環境変数を設定しない
		directAPIKey   string
		wantErr        bool
		wantErrContain string
		wantAPIKey     string
	}{
		{
			name:         "直接指定優先: api_keyとapi_key_env両方指定時、api_keyが優先される",
			envVarName:   "TEST_GEMINI_API_KEY_SECRET",
			envVarValue:  stringPtr("api-key-from-env"),
			directAPIKey: "direct-api-key",
			wantErr:      false,
			wantAPIKey:   "direct-api-key",
		},
		{
			name:           "環境変数未定義: api_key_envで指定した環境変数が未定義の場合エラー",
			envVarName:     "NONEXISTENT_GEMINI_API_KEY_SECRET",
			envVarValue:    nil,
			directAPIKey:   "",
			wantErr:        true,
			wantErrContain: "NONEXISTENT_GEMINI_API_KEY_SECRET",
		},
		{
			name:           "環境変数空文字: 環境変数が空文字の場合エラー",
			envVarName:     "TEST_GEMINI_API_KEY_EMPTY",
			envVarValue:    stringPtr(""),
			directAPIKey:   "",
			wantErr:        true,
			wantErrContain: "TEST_GEMINI_API_KEY_EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数のセットアップ
			if tt.envVarValue != nil {
				err := os.Setenv(tt.envVarName, *tt.envVarValue)
				require.NoError(t, err, "環境変数の設定に成功するはずです")
				defer func() { _ = os.Unsetenv(tt.envVarName) }()
			} else {
				_ = os.Unsetenv(tt.envVarName)
			}

			profile := &infra.Profile{
				AI: &infra.AIConfig{
					Gemini: &infra.GeminiConfig{
						Type:      "gemini-2.5-flash",
						APIKey:    tt.directAPIKey,
						APIKeyEnv: tt.envVarName,
					},
				},
				Prompt: NewPromptConfig(),
				Output: NewOutputConfig(),
			}

			entityProfile, err := profile.ToEntity()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain,
					"エラーメッセージに環境変数名が含まれるはずです")
			} else {
				require.NoError(t, err, "ToEntity()はエラーを返さないはずです")
				assert.Equal(t, tt.wantAPIKey, entityProfile.AI.Gemini.APIKey.Value())
			}
		})
	}
}

// ============================================================================
// Slack APIトークン解決テスト
// ============================================================================

// TestSlackSecret_SecretResolution はSlack APIトークンの解決ロジックを検証する
func TestSlackSecret_SecretResolution(t *testing.T) {
	tests := []struct {
		name           string
		envVarName     string
		envVarValue    *string // nilの場合は環境変数を設定しない
		directToken    string
		wantErr        bool
		wantErrContain string
		wantToken      string
	}{
		{
			name:        "直接指定優先: api_tokenとapi_token_env両方指定時、api_tokenが優先される",
			envVarName:  "TEST_SLACK_API_TOKEN_SECRET",
			envVarValue: stringPtr("token-from-env"),
			directToken: "direct-token",
			wantErr:     false,
			wantToken:   "direct-token",
		},
		{
			name:           "環境変数未定義: api_token_envで指定した環境変数が未定義の場合エラー",
			envVarName:     "NONEXISTENT_SLACK_API_TOKEN_SECRET",
			envVarValue:    nil,
			directToken:    "",
			wantErr:        true,
			wantErrContain: "NONEXISTENT_SLACK_API_TOKEN_SECRET",
		},
		{
			name:           "環境変数空文字: 環境変数が空文字の場合エラー",
			envVarName:     "TEST_SLACK_API_TOKEN_EMPTY",
			envVarValue:    stringPtr(""),
			directToken:    "",
			wantErr:        true,
			wantErrContain: "TEST_SLACK_API_TOKEN_EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数のセットアップ
			if tt.envVarValue != nil {
				err := os.Setenv(tt.envVarName, *tt.envVarValue)
				require.NoError(t, err, "環境変数の設定に成功するはずです")
				defer func() { _ = os.Unsetenv(tt.envVarName) }()
			} else {
				_ = os.Unsetenv(tt.envVarName)
			}

			profile := &infra.Profile{
				AI:     NewAIConfig(),
				Prompt: NewPromptConfig(),
				Output: &infra.OutputConfig{
					SlackAPI: &infra.SlackAPIConfig{
						Enabled:     boolPtr(true),
						APIToken:    tt.directToken,
						APITokenEnv: tt.envVarName,
						Channel:     "#test-channel",
					},
				},
			}

			entityProfile, err := profile.ToEntity()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain,
					"エラーメッセージに環境変数名が含まれるはずです")
			} else {
				require.NoError(t, err, "ToEntity()はエラーを返さないはずです")
				assert.Equal(t, tt.wantToken, entityProfile.Output.SlackAPI.APIToken.Value())
			}
		})
	}
}

// ============================================================================
// Misskey APIトークン解決テスト
// ============================================================================

// TestMisskeySecret_SecretResolution はMisskey APIトークンの解決ロジックを検証する
func TestMisskeySecret_SecretResolution(t *testing.T) {
	tests := []struct {
		name           string
		envVarName     string
		envVarValue    *string // nilの場合は環境変数を設定しない
		directToken    string
		wantErr        bool
		wantErrContain string
		wantToken      string
	}{
		{
			name:        "直接指定優先: api_tokenとapi_token_env両方指定時、api_tokenが優先される",
			envVarName:  "TEST_MISSKEY_API_TOKEN_SECRET",
			envVarValue: stringPtr("token-from-env"),
			directToken: "direct-token",
			wantErr:     false,
			wantToken:   "direct-token",
		},
		{
			name:           "環境変数未定義: api_token_envで指定した環境変数が未定義の場合エラー",
			envVarName:     "NONEXISTENT_MISSKEY_API_TOKEN_SECRET",
			envVarValue:    nil,
			directToken:    "",
			wantErr:        true,
			wantErrContain: "NONEXISTENT_MISSKEY_API_TOKEN_SECRET",
		},
		{
			name:           "環境変数空文字: 環境変数が空文字の場合エラー",
			envVarName:     "TEST_MISSKEY_API_TOKEN_EMPTY",
			envVarValue:    stringPtr(""),
			directToken:    "",
			wantErr:        true,
			wantErrContain: "TEST_MISSKEY_API_TOKEN_EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数のセットアップ
			if tt.envVarValue != nil {
				err := os.Setenv(tt.envVarName, *tt.envVarValue)
				require.NoError(t, err, "環境変数の設定に成功するはずです")
				defer func() { _ = os.Unsetenv(tt.envVarName) }()
			} else {
				_ = os.Unsetenv(tt.envVarName)
			}

			profile := &infra.Profile{
				AI:     NewAIConfig(),
				Prompt: NewPromptConfig(),
				Output: &infra.OutputConfig{
					Misskey: &infra.MisskeyConfig{
						Enabled:     boolPtr(true),
						APIToken:    tt.directToken,
						APITokenEnv: tt.envVarName,
						APIURL:      "http://localhost:3000",
					},
				},
			}

			entityProfile, err := profile.ToEntity()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain,
					"エラーメッセージに環境変数名が含まれるはずです")
			} else {
				require.NoError(t, err, "ToEntity()はエラーを返さないはずです")
				assert.Equal(t, tt.wantToken, entityProfile.Output.Misskey.APIToken.Value())
			}
		})
	}
}

// stringPtr はstring値へのポインタを返すヘルパー関数
func stringPtr(s string) *string {
	return &s
}
