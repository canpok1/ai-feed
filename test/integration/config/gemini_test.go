//go:build integration

// Package config はGemini設定の統合テストを提供する
package config

import (
	"os"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeminiConfig_TypeRequired はGemini設定でTypeフィールドが必須であることを検証する
// ai.gemini.type が省略された場合、バリデーションエラーになること
func TestGeminiConfig_TypeRequired(t *testing.T) {
	// Typeが空のGemini設定を作成
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:   "", // 空文字列
				APIKey: "test-api-key",
			},
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、Typeに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "Typeが空の場合、バリデーションは失敗するはずです")
	assert.NotEmpty(t, result.Errors, "エラーメッセージが含まれているはずです")

	// エラーメッセージにTypeに関する記述があることを確認
	hasTypeError := false
	for _, errMsg := range result.Errors {
		if assert.Regexp(t, `(?i)(type|Type)`, errMsg) {
			hasTypeError = true
			break
		}
	}
	assert.True(t, hasTypeError, "Typeに関するエラーメッセージが含まれているはずです")
}

// TestGeminiConfig_APIKeyRequired はapi_keyまたはapi_key_envのどちらかが必須であることを検証する
// 両方とも省略された場合、バリデーションエラーになること
func TestGeminiConfig_APIKeyRequired(t *testing.T) {
	// APIKeyもAPIKeyEnvも設定されていないGemini設定を作成
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:      "gemini-2.5-flash",
				APIKey:    "",
				APIKeyEnv: "",
			},
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、APIKeyに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "APIKeyが空の場合、バリデーションは失敗するはずです")
	assert.NotEmpty(t, result.Errors, "エラーメッセージが含まれているはずです")

	// エラーメッセージにAPIKeyに関する記述があることを確認
	hasAPIKeyError := false
	for _, errMsg := range result.Errors {
		if assert.Regexp(t, `(?i)(api.*key|APIキー)`, errMsg) {
			hasAPIKeyError = true
			break
		}
	}
	assert.True(t, hasAPIKeyError, "APIKeyに関するエラーメッセージが含まれているはずです")
}

// TestGeminiConfig_APIKeyPrecedence はapi_keyとapi_key_env両方指定時の優先度を検証する
// api_keyが優先され、api_key_envの環境変数は使用されないこと
func TestGeminiConfig_APIKeyPrecedence(t *testing.T) {
	// 環境変数にAPIキーを設定
	const envVarName = "TEST_GEMINI_API_KEY"
	const envAPIKey = "api-key-from-env"
	const directAPIKey = "direct-api-key"

	err := os.Setenv(envVarName, envAPIKey)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	// api_keyとapi_key_envの両方を設定
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

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// api_keyが優先されることを確認
	assert.Equal(t, directAPIKey, entityProfile.AI.Gemini.APIKey.Value(),
		"api_keyがapi_key_envより優先されるはずです")
}

// TestGeminiConfig_APIKeyFromEnv は環境変数からAPIキーを取得できることを検証する
// api_key_envで指定した環境変数の値がAPIキーとして使用されること
func TestGeminiConfig_APIKeyFromEnv(t *testing.T) {
	// 環境変数にAPIキーを設定
	const envVarName = "TEST_GEMINI_API_KEY"
	const envAPIKey = "api-key-from-env"

	err := os.Setenv(envVarName, envAPIKey)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	// api_key_envのみを設定
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

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 環境変数の値が使用されることを確認
	assert.Equal(t, envAPIKey, entityProfile.AI.Gemini.APIKey.Value(),
		"api_key_envで指定した環境変数の値が使用されるはずです")
}

// TestGeminiConfig_EnvVarNotSet は環境変数が設定されていない場合のエラーを検証する
// api_key_envで指定した環境変数が存在しない場合、ToEntity()でエラーになること
func TestGeminiConfig_EnvVarNotSet(t *testing.T) {
	// 存在しない環境変数を指定
	const envVarName = "NONEXISTENT_GEMINI_API_KEY"

	// 環境変数が設定されていないことを確認
	_ = os.Unsetenv(envVarName)

	// api_key_envのみを設定（api_keyは空）
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

	// infra.Profile から entity.Profile に変換
	_, err := profile.ToEntity()

	// 環境変数が設定されていない場合、エラーが返されることを確認
	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestGeminiConfig_NilAIConfig はAI設定がnilの場合のバリデーションエラーを検証する
// ai設定が省略された場合、バリデーションエラーになること
func TestGeminiConfig_NilAIConfig(t *testing.T) {
	// AI設定がnilのProfile
	profile := &infra.Profile{
		AI:     nil,
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗することを確認
	assert.False(t, result.IsValid, "AI設定がnilの場合、バリデーションは失敗するはずです")
	assert.NotEmpty(t, result.Errors, "エラーメッセージが含まれているはずです")

	// エラーメッセージにAI設定に関する記述があることを確認
	hasAIError := false
	for _, errMsg := range result.Errors {
		if assert.Regexp(t, `(?i)(ai|AI|Gemini)`, errMsg) {
			hasAIError = true
			break
		}
	}
	assert.True(t, hasAIError, "AI設定に関するエラーメッセージが含まれているはずです")
}

// TestGeminiConfig_ValidConversion は正しい設定がentity.Profileに変換できることを検証する
// すべての必須フィールドが正しく設定されている場合、正常に変換・バリデーションが完了すること
func TestGeminiConfig_ValidConversion(t *testing.T) {
	// 正しいGemini設定を作成
	profile := ValidInfraProfile()

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 変換されたProfileの値を検証
	require.NotNil(t, entityProfile.AI, "AI設定が存在するはずです")
	require.NotNil(t, entityProfile.AI.Gemini, "Gemini設定が存在するはずです")
	assert.Equal(t, "gemini-2.5-flash", entityProfile.AI.Gemini.Type, "Typeが正しく変換されるはずです")
	assert.Equal(t, "test-api-key", entityProfile.AI.Gemini.APIKey.Value(), "APIKeyが正しく変換されるはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認
	assert.True(t, result.IsValid, "正しい設定の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestGeminiConfig_NilGeminiConfig はGemini設定がnilの場合のバリデーションエラーを検証する
// ai.gemini設定が省略された場合、バリデーションエラーになること
func TestGeminiConfig_NilGeminiConfig(t *testing.T) {
	// Gemini設定がnilのProfile
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: nil,
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗することを確認
	assert.False(t, result.IsValid, "Gemini設定がnilの場合、バリデーションは失敗するはずです")
	assert.NotEmpty(t, result.Errors, "エラーメッセージが含まれているはずです")
}
