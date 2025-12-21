//go:build integration

package config

import (
	"os"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMisskeyTestProfile は有効なMisskey設定を含むテスト用プロファイルを生成する
// 各テストでこのベースプロファイルを変更して条件を作り出す
func newMisskeyTestProfile() *infra.Profile {
	enabled := true
	messageTemplate := "{{.Comment}}\n{{.Article.Title}}"
	return &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: &infra.OutputConfig{
			Misskey: &infra.MisskeyConfig{
				Enabled:         &enabled,
				APIToken:        "test-api-token",
				APIURL:          "https://misskey.example.com",
				MessageTemplate: &messageTemplate,
			},
		},
	}
}

// containsErrorWithSubstring はエラーリスト内に指定した部分文字列を含むエラーが存在するかを確認する
func containsErrorWithSubstring(errors []string, substring string) bool {
	for _, e := range errors {
		if strings.Contains(e, substring) {
			return true
		}
	}
	return false
}

// TestMisskeyConfig_APITokenRequired はenabled=true時にapi_tokenまたはapi_token_envのどちらかが必須であることを検証する
// 両方とも省略された場合、バリデーションエラーになること
func TestMisskeyConfig_APITokenRequired(t *testing.T) {
	// ベースプロファイルからAPIトークンを削除
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.APIToken = ""
	profile.Output.Misskey.APITokenEnv = ""

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、APITokenに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "APITokenが空の場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Misskey APIトークンが設定されていません",
		"APITokenに関するエラーメッセージが含まれているはずです")
}

// TestMisskeyConfig_APIURLRequired はenabled=true時にapi_urlが必須であることを検証する
// api_urlが省略された場合、バリデーションエラーになること
func TestMisskeyConfig_APIURLRequired(t *testing.T) {
	// ベースプロファイルからAPIURLを削除
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.APIURL = ""

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、APIURLに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "APIURLが空の場合、バリデーションは失敗するはずです")
	assert.True(t, containsErrorWithSubstring(result.Errors, "Misskey API URL"),
		"APIURLに関するエラーメッセージが含まれているはずです: %v", result.Errors)
}

// TestMisskeyConfig_URLFormatValidation はURL形式の妥当性を検証する
// 不正なURL形式の場合、バリデーションエラーになること
func TestMisskeyConfig_URLFormatValidation(t *testing.T) {
	tests := []struct {
		name    string
		apiURL  string
		wantErr bool
	}{
		{
			name:    "正常系: https URL",
			apiURL:  "https://misskey.example.com",
			wantErr: false,
		},
		{
			name:    "正常系: http URL",
			apiURL:  "http://localhost:3000",
			wantErr: false,
		},
		{
			name:    "異常系: スキームなし",
			apiURL:  "misskey.example.com",
			wantErr: true,
		},
		{
			name:    "異常系: 無効なURL",
			apiURL:  "not-a-valid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ベースプロファイルのAPIURLを変更
			profile := newMisskeyTestProfile()
			profile.Output.Misskey.APIURL = tt.apiURL

			// infra.Profile から entity.Profile に変換
			entityProfile, err := profile.ToEntity()
			require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

			// entity.Profile のバリデーションを実行
			result := entityProfile.Validate()

			if tt.wantErr {
				assert.False(t, result.IsValid, "無効なURLの場合、バリデーションは失敗するはずです")
				assert.True(t, containsErrorWithSubstring(result.Errors, "Misskey API URL"),
					"URLに関するエラーメッセージが含まれているはずです")
			} else {
				// Misskey設定のバリデーションエラーのみをチェック
				assert.False(t, containsErrorWithSubstring(result.Errors, "Misskey API URL"),
					"有効なURLの場合、URL関連のエラーは発生しないはずです")
			}
		})
	}
}

// TestMisskeyConfig_MessageTemplateRequired はenabled=true時にmessage_templateが必須であることを検証する
// message_templateが省略された場合、バリデーションエラーになること
func TestMisskeyConfig_MessageTemplateRequired(t *testing.T) {
	// ベースプロファイルからMessageTemplateを削除
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.MessageTemplate = nil

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、MessageTemplateに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "MessageTemplateがnilの場合、バリデーションは失敗するはずです")
	assert.True(t, containsErrorWithSubstring(result.Errors, "Misskeyメッセージテンプレートが設定されていません"),
		"MessageTemplateに関するエラーメッセージが含まれているはずです")
}

// TestMisskeyConfig_APITokenPrecedence はapi_tokenとapi_token_env両方指定時の優先度を検証する
// api_tokenが優先され、api_token_envの環境変数は使用されないこと
func TestMisskeyConfig_APITokenPrecedence(t *testing.T) {
	// 環境変数にAPIトークンを設定
	const envVarName = "TEST_MISSKEY_API_TOKEN"
	const envAPIToken = "api-token-from-env"
	const directAPIToken = "direct-api-token"

	err := os.Setenv(envVarName, envAPIToken)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	// ベースプロファイルにapi_tokenとapi_token_envの両方を設定
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.APIToken = directAPIToken
	profile.Output.Misskey.APITokenEnv = envVarName

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// api_tokenが優先されることを確認
	assert.Equal(t, directAPIToken, entityProfile.Output.Misskey.APIToken.Value(),
		"api_tokenがapi_token_envより優先されるはずです")
}

// TestMisskeyConfig_TemplateSyntaxError はテンプレート構文エラーの検出を検証する
// 無効なテンプレート構文の場合、バリデーションエラーになること
func TestMisskeyConfig_TemplateSyntaxError(t *testing.T) {
	// ベースプロファイルに無効なテンプレートを設定
	profile := newMisskeyTestProfile()
	invalidTemplate := "{{.Comment" // 閉じタグがない
	profile.Output.Misskey.MessageTemplate = &invalidTemplate

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗し、テンプレート構文に関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "無効なテンプレート構文の場合、バリデーションは失敗するはずです")
	assert.True(t, containsErrorWithSubstring(result.Errors, "Misskeyメッセージテンプレートが無効です"),
		"テンプレート構文エラーに関するメッセージが含まれているはずです")
}

// TestMisskeyConfig_EnabledDefaultValue はenabled省略時のデフォルト値（後方互換性）を検証する
// enabledが省略された場合、trueとして扱われること
func TestMisskeyConfig_EnabledDefaultValue(t *testing.T) {
	// ベースプロファイルのEnabledをnilに設定
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.Enabled = nil

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// Enabledがtrueとして扱われることを確認
	assert.True(t, *entityProfile.Output.Misskey.Enabled,
		"Enabledが省略された場合、trueとして扱われるはずです（後方互換性）")
}

// TestMisskeyConfig_ValidationSkippedWhenDisabled はenabled=false時のバリデーションスキップを検証する
// enabled=falseの場合、他の必須フィールドが空でもバリデーションが成功すること
func TestMisskeyConfig_ValidationSkippedWhenDisabled(t *testing.T) {
	// ベースプロファイルをenabled=falseに変更し、他のフィールドを空に
	profile := newMisskeyTestProfile()
	enabled := false
	profile.Output.Misskey.Enabled = &enabled
	profile.Output.Misskey.APIToken = ""
	profile.Output.Misskey.APIURL = ""
	profile.Output.Misskey.MessageTemplate = nil

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// Misskey関連のエラーがないことを確認
	assert.False(t, containsErrorWithSubstring(result.Errors, "Misskey"),
		"enabled=falseの場合、Misskey関連のバリデーションエラーは発生しないはずです")
}

// TestMisskeyConfig_APITokenFromEnv は環境変数からAPIトークンを取得できることを検証する
// api_token_envで指定した環境変数の値がAPIトークンとして使用されること
func TestMisskeyConfig_APITokenFromEnv(t *testing.T) {
	// 環境変数にAPIトークンを設定
	const envVarName = "TEST_MISSKEY_API_TOKEN"
	const envAPIToken = "api-token-from-env"

	err := os.Setenv(envVarName, envAPIToken)
	require.NoError(t, err, "環境変数の設定に成功するはずです")
	defer func() { _ = os.Unsetenv(envVarName) }()

	// ベースプロファイルでapi_token_envのみを設定
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.APIToken = ""
	profile.Output.Misskey.APITokenEnv = envVarName

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 環境変数の値が使用されることを確認
	assert.Equal(t, envAPIToken, entityProfile.Output.Misskey.APIToken.Value(),
		"api_token_envで指定した環境変数の値が使用されるはずです")
}

// TestMisskeyConfig_EnvVarNotSet は環境変数が設定されていない場合のエラーを検証する
// api_token_envで指定した環境変数が存在しない場合、ToEntity()でエラーになること
func TestMisskeyConfig_EnvVarNotSet(t *testing.T) {
	// 存在しない環境変数を指定
	const envVarName = "NONEXISTENT_MISSKEY_API_TOKEN"

	// 環境変数が設定されていないことを確認
	_ = os.Unsetenv(envVarName)

	// ベースプロファイルでapi_token_envのみを設定（api_tokenは空）
	profile := newMisskeyTestProfile()
	profile.Output.Misskey.APIToken = ""
	profile.Output.Misskey.APITokenEnv = envVarName

	// infra.Profile から entity.Profile に変換
	_, err := profile.ToEntity()

	// 環境変数が設定されていない場合、エラーが返されることを確認
	assert.Error(t, err, "指定された環境変数が設定されていない場合、エラーが返されるはずです")
	assert.Contains(t, err.Error(), envVarName, "エラーメッセージに環境変数名が含まれるはずです")
}

// TestMisskeyConfig_ValidConfig は正しい設定がentity.Profileに変換できることを検証する
// すべての必須フィールドが正しく設定されている場合、正常に変換・バリデーションが完了すること
func TestMisskeyConfig_ValidConfig(t *testing.T) {
	// ベースプロファイルをそのまま使用（有効な設定）
	profile := newMisskeyTestProfile()

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 変換されたProfileの値を検証
	require.NotNil(t, entityProfile.Output, "Output設定が存在するはずです")
	require.NotNil(t, entityProfile.Output.Misskey, "Misskey設定が存在するはずです")
	assert.True(t, *entityProfile.Output.Misskey.Enabled, "Enabledが正しく変換されるはずです")
	assert.Equal(t, "test-api-token", entityProfile.Output.Misskey.APIToken.Value(), "APITokenが正しく変換されるはずです")
	assert.Equal(t, "https://misskey.example.com", entityProfile.Output.Misskey.APIURL, "APIURLが正しく変換されるはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// Misskey関連のエラーがないことを確認
	assert.False(t, containsErrorWithSubstring(result.Errors, "Misskey"),
		"正しい設定の場合、Misskey関連のエラーは発生しないはずです")
}
