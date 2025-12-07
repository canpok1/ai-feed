//go:build integration

// Package config は統合テスト用の設定ヘルパー関数を提供する
package config

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
)

// ValidInfraProfile はテスト用の有効な infra.Profile を生成する
// すべての必須フィールドが設定された設定オブジェクトを返す
func ValidInfraProfile() *infra.Profile {
	return &infra.Profile{
		AI:     NewAIConfig(),
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}
}

// ValidEntityProfile はテスト用の有効な entity.Profile を生成する
// すべての必須フィールドが設定された設定オブジェクトを返す
func ValidEntityProfile() *entity.Profile {
	return &entity.Profile{
		AI:     NewEntityAIConfig(),
		Prompt: NewEntityPromptConfig(),
		Output: NewEntityOutputConfig(),
	}
}

// NewGeminiConfig はテスト用の infra.GeminiConfig を生成する
func NewGeminiConfig() *infra.GeminiConfig {
	return &infra.GeminiConfig{
		Type:   "gemini-2.5-flash",
		APIKey: "test-api-key",
	}
}

// NewAIConfig はテスト用の infra.AIConfig を生成する
func NewAIConfig() *infra.AIConfig {
	return &infra.AIConfig{
		Gemini: NewGeminiConfig(),
	}
}

// NewPromptConfig はテスト用の infra.PromptConfig を生成する
func NewPromptConfig() *infra.PromptConfig {
	return &infra.PromptConfig{
		SystemPrompt:          "あなたはテスト用のアシスタントです。",
		CommentPromptTemplate: "以下の記事の紹介文を100字以内で作成してください。\n記事タイトル: {{TITLE}}\n記事URL: {{URL}}\n記事内容:\n{{CONTENT}}",
		SelectorPrompt:        "以下の記事一覧から、最も興味深い記事を1つ選択してください。",
		FixedMessage:          "",
	}
}

// NewSlackAPIConfig はテスト用の infra.SlackAPIConfig を生成する
func NewSlackAPIConfig() *infra.SlackAPIConfig {
	enabled := true
	messageTemplate := "{{COMMENT}}\n<{{URL}}|{{TITLE}}>"
	apiURL := "http://localhost:8080/api/"
	return &infra.SlackAPIConfig{
		Enabled:         &enabled,
		APIToken:        "test-slack-token",
		Channel:         "#test-channel",
		MessageTemplate: &messageTemplate,
		APIURL:          &apiURL,
	}
}

// NewMisskeyConfig はテスト用の infra.MisskeyConfig を生成する
func NewMisskeyConfig() *infra.MisskeyConfig {
	enabled := true
	messageTemplate := "{{COMMENT}}\n[{{TITLE}}]({{URL}})"
	return &infra.MisskeyConfig{
		Enabled:         &enabled,
		APIToken:        "test-misskey-token",
		APIURL:          "http://localhost:8081",
		MessageTemplate: &messageTemplate,
	}
}

// NewOutputConfig はテスト用の infra.OutputConfig を生成する
// SlackAPI と Misskey の両方の設定を含む
func NewOutputConfig() *infra.OutputConfig {
	return &infra.OutputConfig{
		SlackAPI: NewSlackAPIConfig(),
		Misskey:  NewMisskeyConfig(),
	}
}

// NewEntityGeminiConfig はテスト用の entity.GeminiConfig を生成する
func NewEntityGeminiConfig() *entity.GeminiConfig {
	return &entity.GeminiConfig{
		Type:   "gemini-2.5-flash",
		APIKey: entity.NewSecretString("test-api-key"),
	}
}

// NewEntityAIConfig はテスト用の entity.AIConfig を生成する
func NewEntityAIConfig() *entity.AIConfig {
	return &entity.AIConfig{
		Gemini: NewEntityGeminiConfig(),
	}
}

// NewEntityPromptConfig はテスト用の entity.PromptConfig を生成する
func NewEntityPromptConfig() *entity.PromptConfig {
	return &entity.PromptConfig{
		SystemPrompt:          "あなたはテスト用のアシスタントです。",
		CommentPromptTemplate: "以下の記事の紹介文を100字以内で作成してください。\n記事タイトル: {{.Title}}\n記事URL: {{.Link}}\n記事内容:\n{{.Content}}",
		SelectorPrompt:        "以下の記事一覧から、最も興味深い記事を1つ選択してください。",
		FixedMessage:          "",
	}
}

// NewEntitySlackAPIConfig はテスト用の entity.SlackAPIConfig を生成する
func NewEntitySlackAPIConfig() *entity.SlackAPIConfig {
	messageTemplate := "{{if .Comment}}{{.Comment}}\n{{end}}<{{.Article.Link}}|{{.Article.Title}}>"
	apiURL := "http://localhost:8080/api/"
	return &entity.SlackAPIConfig{
		Enabled:         true,
		APIToken:        entity.NewSecretString("test-slack-token"),
		Channel:         "#test-channel",
		MessageTemplate: &messageTemplate,
		APIURL:          &apiURL,
	}
}

// NewEntityMisskeyConfig はテスト用の entity.MisskeyConfig を生成する
func NewEntityMisskeyConfig() *entity.MisskeyConfig {
	messageTemplate := "{{if .Comment}}{{.Comment}}\n{{end}}[{{.Article.Title}}]({{.Article.Link}})"
	return &entity.MisskeyConfig{
		Enabled:         true,
		APIToken:        entity.NewSecretString("test-misskey-token"),
		APIURL:          "http://localhost:8081",
		MessageTemplate: &messageTemplate,
	}
}

// NewEntityOutputConfig はテスト用の entity.OutputConfig を生成する
// SlackAPI と Misskey の両方の設定を含む
func NewEntityOutputConfig() *entity.OutputConfig {
	return &entity.OutputConfig{
		SlackAPI: NewEntitySlackAPIConfig(),
		Misskey:  NewEntityMisskeyConfig(),
	}
}

// WithSlackAPIOnly は SlackAPI のみを含む infra.OutputConfig を生成する
func WithSlackAPIOnly() *infra.OutputConfig {
	return &infra.OutputConfig{
		SlackAPI: NewSlackAPIConfig(),
	}
}

// WithMisskeyOnly は Misskey のみを含む infra.OutputConfig を生成する
func WithMisskeyOnly() *infra.OutputConfig {
	return &infra.OutputConfig{
		Misskey: NewMisskeyConfig(),
	}
}

// WithDisabledSlackAPI は無効化された SlackAPI 設定を含む infra.OutputConfig を生成する
func WithDisabledSlackAPI() *infra.OutputConfig {
	config := NewSlackAPIConfig()
	*config.Enabled = false
	return &infra.OutputConfig{
		SlackAPI: config,
	}
}

// WithDisabledMisskey は無効化された Misskey 設定を含む infra.OutputConfig を生成する
func WithDisabledMisskey() *infra.OutputConfig {
	config := NewMisskeyConfig()
	*config.Enabled = false
	return &infra.OutputConfig{
		Misskey: config,
	}
}

// NewMockConfig はテスト用の有効な infra.MockConfig を生成する
// デフォルトで有効化され、selector_mode="first"、comment="テストコメント"を設定
func NewMockConfig() *infra.MockConfig {
	return NewMockConfigWithMode("first")
}

// NewMockConfigWithMode は指定したselector_modeでテスト用の infra.MockConfig を生成する
func NewMockConfigWithMode(mode string) *infra.MockConfig {
	enabled := true
	return &infra.MockConfig{
		Enabled:      &enabled,
		SelectorMode: mode,
		Comment:      "テストコメント",
	}
}

// NewDisabledMockConfig は無効化された infra.MockConfig を生成する
func NewDisabledMockConfig() *infra.MockConfig {
	enabled := false
	return &infra.MockConfig{
		Enabled:      &enabled,
		SelectorMode: "",
		Comment:      "",
	}
}

// NewAIConfigWithMock はMock設定のみを含む infra.AIConfig を生成する
// Gemini設定はnilになる
func NewAIConfigWithMock() *infra.AIConfig {
	return &infra.AIConfig{
		Gemini: nil,
		Mock:   NewMockConfig(),
	}
}

// NewAIConfigWithBothMockAndGemini はMockとGemini両方の設定を含む infra.AIConfig を生成する
// Mock設定が有効な場合、Gemini設定より優先される
func NewAIConfigWithBothMockAndGemini() *infra.AIConfig {
	return &infra.AIConfig{
		Gemini: NewGeminiConfig(),
		Mock:   NewMockConfig(),
	}
}

// NewEntityMockConfig はテスト用の有効な entity.MockConfig を生成する
// Enabled=true、SelectorMode="first"、Comment="テストコメント"を設定
func NewEntityMockConfig() *entity.MockConfig {
	return NewEntityMockConfigWithMode("first")
}

// NewEntityMockConfigWithMode は指定したselector_modeでテスト用の entity.MockConfig を生成する
func NewEntityMockConfigWithMode(mode string) *entity.MockConfig {
	return &entity.MockConfig{
		Enabled:      true,
		SelectorMode: mode,
		Comment:      "テストコメント",
	}
}

// NewEntityAIConfigWithMock はMock設定のみを含む entity.AIConfig を生成する
func NewEntityAIConfigWithMock() *entity.AIConfig {
	return &entity.AIConfig{
		Gemini: nil,
		Mock:   NewEntityMockConfig(),
	}
}

// ValidInfraProfileWithMock はMock設定を使用したテスト用の有効な infra.Profile を生成する
func ValidInfraProfileWithMock() *infra.Profile {
	return &infra.Profile{
		AI:     NewAIConfigWithMock(),
		Prompt: NewPromptConfig(),
		Output: NewOutputConfig(),
	}
}

// ValidEntityProfileWithMock はMock設定を使用したテスト用の有効な entity.Profile を生成する
func ValidEntityProfileWithMock() *entity.Profile {
	return &entity.Profile{
		AI:     NewEntityAIConfigWithMock(),
		Prompt: NewEntityPromptConfig(),
		Output: NewEntityOutputConfig(),
	}
}
