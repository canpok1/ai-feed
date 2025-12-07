//go:build integration

// Package config はプロンプト設定の統合テストを提供する
package config

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptConfig_RequiredFields はプロンプト設定の必須フィールドに対するバリデーションを検証する
func TestPromptConfig_RequiredFields(t *testing.T) {
	basePrompt := &infra.PromptConfig{
		SystemPrompt:          "あなたはテスト用のアシスタントです。",
		CommentPromptTemplate: "以下の記事の紹介文を作成してください。\n記事タイトル: {{TITLE}}",
		SelectorPrompt:        "以下の記事一覧から、最も興味深い記事を1つ選択してください。",
		FixedMessage:          "",
	}

	tests := []struct {
		name          string
		modifier      func(p *infra.PromptConfig)
		expectedError string
	}{
		{
			name: "system_prompt is required",
			modifier: func(p *infra.PromptConfig) {
				p.SystemPrompt = ""
			},
			expectedError: "システムプロンプトが設定されていません",
		},
		{
			name: "comment_prompt_template is required",
			modifier: func(p *infra.PromptConfig) {
				p.CommentPromptTemplate = ""
			},
			expectedError: "コメントプロンプトテンプレートが設定されていません",
		},
		{
			name: "selector_prompt is required",
			modifier: func(p *infra.PromptConfig) {
				p.SelectorPrompt = ""
			},
			expectedError: "記事選択プロンプトが設定されていません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ベースとなる設定をコピーして、テストケースごとに変更を加える
			prompt := *basePrompt
			tt.modifier(&prompt)

			profile := &infra.Profile{
				AI:     NewAIConfig(),
				Prompt: &prompt,
				Output: NewOutputConfig(),
			}

			entityProfile, err := profile.ToEntity()
			require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

			result := entityProfile.Validate()

			assert.False(t, result.IsValid, "必須フィールドが空の場合、バリデーションは失敗するはずです")
			assert.Contains(t, result.Errors, tt.expectedError, "必須フィールドに関するエラーメッセージが含まれているはずです")
		})
	}
}

// TestPromptConfig_FixedMessageOptional はfixed_messageがオプションであることを検証する
// prompt.fixed_message が空文字列でもバリデーションエラーにならないこと
func TestPromptConfig_FixedMessageOptional(t *testing.T) {
	// FixedMessageが空（デフォルト）のプロンプト設定を作成
	profile := &infra.Profile{
		AI: NewAIConfig(),
		Prompt: &infra.PromptConfig{
			SystemPrompt:          "あなたはテスト用のアシスタントです。",
			CommentPromptTemplate: "以下の記事の紹介文を作成してください。\n記事タイトル: {{TITLE}}",
			SelectorPrompt:        "以下の記事一覧から、最も興味深い記事を1つ選択してください。",
			FixedMessage:          "", // 空文字列（オプショナル）
		},
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認（FixedMessageは任意項目のため）
	assert.True(t, result.IsValid, "FixedMessageが空の場合でも、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestPromptConfig_ValidConversion は正しいプロンプト設定がentity.Profileに変換できることを検証する
// すべての必須フィールドが正しく設定されている場合、正常に変換・バリデーションが完了すること
func TestPromptConfig_ValidConversion(t *testing.T) {
	// 正しいプロンプト設定を作成
	profile := ValidInfraProfile()

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 変換されたProfileの値を検証
	require.NotNil(t, entityProfile.Prompt, "Prompt設定が存在するはずです")
	assert.Equal(t, "あなたはテスト用のアシスタントです。", entityProfile.Prompt.SystemPrompt,
		"SystemPromptが正しく変換されるはずです")
	assert.NotEmpty(t, entityProfile.Prompt.CommentPromptTemplate,
		"CommentPromptTemplateが正しく変換されるはずです")
	assert.Equal(t, "以下の記事一覧から、最も興味深い記事を1つ選択してください。", entityProfile.Prompt.SelectorPrompt,
		"SelectorPromptが正しく変換されるはずです")
	assert.Empty(t, entityProfile.Prompt.FixedMessage,
		"FixedMessageは空文字列のはずです（デフォルト値）")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが成功することを確認
	assert.True(t, result.IsValid, "正しい設定の場合、バリデーションは成功するはずです")
	assert.Empty(t, result.Errors, "エラーメッセージがないはずです")
}

// TestPromptConfig_NilPromptConfig はPrompt設定がnilの場合のバリデーションエラーを検証する
// prompt設定が省略された場合、バリデーションエラーになること
func TestPromptConfig_NilPromptConfig(t *testing.T) {
	// Prompt設定がnilのProfile
	profile := &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: nil,
		Output: NewOutputConfig(),
	}

	// infra.Profile から entity.Profile に変換
	entityProfile, err := profile.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// entity.Profile のバリデーションを実行
	result := entityProfile.Validate()

	// バリデーションが失敗することを確認
	assert.False(t, result.IsValid, "Prompt設定がnilの場合、バリデーションは失敗するはずです")
	assert.Contains(t, result.Errors, "プロンプト設定が設定されていません",
		"Prompt設定に関するエラーメッセージが含まれているはずです")
}
