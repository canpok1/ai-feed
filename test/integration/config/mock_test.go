//go:build integration

// Package config はMock設定の統合テストを提供する
package config

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockConfig_SelectorModeRequired はMock設定が有効な場合にselector_modeが必須であることを検証する
// ai.mock.enabled=true の場合、selector_modeが設定されていないとバリデーションエラーになること
func TestMockConfig_SelectorModeRequired(t *testing.T) {
	// selector_modeが空のMock設定を作成
	enabled := true
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Mock: &infra.MockConfig{
				Enabled:      &enabled,
				SelectorMode: "", // 空文字列
				Comment:      "テストコメント",
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

	// バリデーションが失敗し、selector_modeに関するエラーが含まれることを確認
	assert.False(t, result.IsValid, "selector_modeが空の場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Mockの記事選択モードが不正です。first, random, lastのいずれかを指定してください",
		"selector_modeに関するエラーメッセージが含まれているはずです")
}

// TestMockConfig_ValidSelectorModes は有効なselector_modeの値を検証する
// "first", "random", "last"のいずれかが設定されている場合、バリデーションが成功すること
func TestMockConfig_ValidSelectorModes(t *testing.T) {
	validModes := []string{"first", "random", "last"}

	for _, mode := range validModes {
		t.Run("mode="+mode, func(t *testing.T) {
			// 有効なselector_modeを設定
			profile := &infra.Profile{
				AI: &infra.AIConfig{
					Mock: NewMockConfigWithMode(mode),
				},
				Prompt: NewPromptConfig(),
				Output: NewOutputConfig(),
			}

			// infra.Profile から entity.Profile に変換
			entityProfile, err := profile.ToEntity()
			require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

			// entity.Profile のバリデーションを実行
			result := entityProfile.Validate()

			// バリデーションが成功することを確認
			assert.True(t, result.IsValid, "mode=%sの場合、バリデーションは成功するはずです", mode)
			assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
		})
	}
}

// TestMockConfig_InvalidSelectorModes は無効なselector_modeの値を検証する
// "first", "random", "last"以外の値が設定されている場合、バリデーションエラーになること
func TestMockConfig_InvalidSelectorModes(t *testing.T) {
	invalidModes := []string{"invalid", "second", "middle", "FIRST", "Random", "LAST"}

	for _, mode := range invalidModes {
		t.Run("mode="+mode, func(t *testing.T) {
			// 無効なselector_modeを設定
			profile := &infra.Profile{
				AI: &infra.AIConfig{
					Mock: NewMockConfigWithMode(mode),
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
			assert.False(t, result.IsValid, "mode=%sの場合、バリデーションは失敗するはずです", mode)
			assert.Contains(t, result.Errors, "Mockの記事選択モードが不正です。first, random, lastのいずれかを指定してください",
				"selector_modeに関するエラーメッセージが含まれているはずです")
		})
	}
}

// TestMockConfig_DisabledSkipsValidation はMock設定が無効な場合にバリデーションがスキップされることを検証する
// ai.mock.enabled=false の場合、selector_modeが設定されていなくてもエラーにならないこと
func TestMockConfig_DisabledSkipsValidation(t *testing.T) {
	// Mock設定を無効化（selector_modeは空）
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: NewGeminiConfig(), // Gemini設定があるので全体バリデーションは通る
			Mock:   NewDisabledMockConfig(),
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認（Mock設定のバリデーションがスキップされる）
	assert.True(t, result.IsValid, "Mock設定が無効の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestMockConfig_TakesPrecedenceOverGemini はMock設定がGemini設定より優先されることを検証する
// ai.mock.enabled=true の場合、Gemini設定がなくてもバリデーションが成功すること
func TestMockConfig_TakesPrecedenceOverGemini(t *testing.T) {
	// Gemini設定なし、Mock設定あり
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: nil, // Gemini設定なし
			Mock:   NewMockConfig(),
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認（Mock設定が優先され、Gemini設定は不要）
	assert.True(t, result.IsValid, "Mock設定が有効な場合、Gemini設定がなくてもバリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestMockConfig_BothMockAndGeminiWithMockEnabled はMockとGemini両方設定時のMock優先を検証する
// ai.mock.enabled=true の場合、Gemini設定が不完全でもMock設定が有効ならバリデーションが成功すること
func TestMockConfig_BothMockAndGeminiWithMockEnabled(t *testing.T) {
	// Gemini設定が不完全（APIKeyなし）、Mock設定あり
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:   "gemini-2.5-flash",
				APIKey: "", // APIKeyなし
			},
			Mock: NewMockConfig(),
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認（Mock設定が優先され、Gemini設定のエラーは無視）
	assert.True(t, result.IsValid, "Mock設定が有効な場合、Gemini設定が不完全でもバリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestMockConfig_ValidConversion は正しいMock設定がentity.Profileに変換できることを検証する
// すべての必須フィールドが正しく設定されている場合、正常に変換・バリデーションが完了すること
func TestMockConfig_ValidConversion(t *testing.T) {
	// 正しいMock設定を作成
	profile := ValidInfraProfileWithMock()

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 変換されたProfileの値を検証
	require.NotNil(t, entityProfile.AI, "AI設定が存在するはずです")
	require.NotNil(t, entityProfile.AI.Mock, "Mock設定が存在するはずです")
	assert.True(t, entityProfile.AI.Mock.Enabled, "Enabledがtrueであるはずです")
	assert.Equal(t, "first", entityProfile.AI.Mock.SelectorMode, "SelectorModeが正しく変換されるはずです")
	assert.Equal(t, "テストコメント", entityProfile.AI.Mock.Comment, "Commentが正しく変換されるはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認
	assert.True(t, result.IsValid, "正しい設定の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestMockConfig_NilMockConfig はMock設定がnilの場合の挙動を検証する
// ai.mock設定がない場合、Gemini設定が必要になること
func TestMockConfig_NilMockConfig(t *testing.T) {
	// Mock設定がnilのProfile（Gemini設定もなし）
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: nil,
			Mock:   nil,
		},
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗することを確認（Mock設定もGemini設定もない）
	assert.False(t, result.IsValid, "Mock設定もGemini設定もない場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "Gemini設定が設定されていません",
		"Gemini設定に関するエラーメッセージが含まれているはずです")
}

// TestMockConfig_CommentOptional はコメントがオプションであることを検証する
// ai.mock.comment が空でもバリデーションが成功すること
func TestMockConfig_CommentOptional(t *testing.T) {
	// コメントが空のMock設定を作成
	enabled := true
	profile := &infra.Profile{
		AI: &infra.AIConfig{
			Mock: &infra.MockConfig{
				Enabled:      &enabled,
				SelectorMode: "first",
				Comment:      "", // 空コメント
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

	// バリデーションが成功することを確認（コメントはオプション）
	assert.True(t, result.IsValid, "コメントが空でもバリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}
