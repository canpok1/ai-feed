package entity

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/template"
)

// テンプレートキャッシュ（スレッドセーフ）
var templateCache sync.Map

// デフォルト値の定数
const (
	DefaultGeminiAPIKey    = "YOUR_GEMINI_API_KEY_HERE"
	DefaultSlackAPIToken   = "xoxb-YOUR_SLACK_API_TOKEN_HERE"
	DefaultMisskeyAPIToken = "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE"
)

type AIConfig struct {
	Gemini *GeminiConfig
}

// Validate はAIConfigの内容をバリデーションする
func (a *AIConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Gemini: 必須項目（nilでない）
	if a.Gemini == nil {
		builder.AddError("Gemini設定が設定されていません")
	} else {
		// GeminiConfig.Validate()を呼び出して、結果を集約
		builder.MergeResult(a.Gemini.Validate())
	}

	return builder.Build()
}

type GeminiConfig struct {
	Type   string
	APIKey string
}

// Validate はGeminiConfigの内容をバリデーションする
func (g *GeminiConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Type: 必須項目（空文字列でない）
	if err := ValidateRequired(g.Type, "Gemini設定のType"); err != nil {
		builder.AddError(err.Error())
	}

	// APIKey: 必須項目（空文字列でない）
	if err := ValidateRequiredWithDefault(g.APIKey, DefaultGeminiAPIKey, "Gemini API key"); err != nil {
		builder.AddError(err.Error())
	}

	return builder.Build()
}

type PromptConfig struct {
	SystemPrompt          string
	CommentPromptTemplate string
	FixedMessage          string
}

// Validate はPromptConfigの内容をバリデーションする
func (p *PromptConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// SystemPrompt: 必須項目（空文字列でない）
	if err := ValidateRequired(p.SystemPrompt, "システムプロンプト"); err != nil {
		builder.AddError(err.Error())
	}

	// CommentPromptTemplate: 必須項目（空文字列でない）
	if err := ValidateRequired(p.CommentPromptTemplate, "コメントプロンプトテンプレート"); err != nil {
		builder.AddError(err.Error())
	}

	// FixedMessage: 任意項目（空文字列でも可）

	return builder.Build()
}

// BuildCommentPrompt はtext/templateを使用してコメントプロンプトを生成する
func (c *PromptConfig) BuildCommentPrompt(article *Article) string {
	// 後方互換性のため、古い形式のプレースホルダーを新形式に変換
	templateStr := c.CommentPromptTemplate
	templateStr = strings.ReplaceAll(templateStr, "{{title}}", "{{.Title}}")
	templateStr = strings.ReplaceAll(templateStr, "{{url}}", "{{.Link}}")
	templateStr = strings.ReplaceAll(templateStr, "{{content}}", "{{.Content}}")

	// キャッシュからテンプレートを取得
	var tmpl *template.Template
	if cached, ok := templateCache.Load(templateStr); ok {
		tmpl = cached.(*template.Template)
	} else {
		// キャッシュにない場合はパースして保存
		var err error
		tmpl, err = template.New("comment").Parse(templateStr)
		if err != nil {
			// テンプレートの解析に失敗した場合は、元のテンプレートを返す
			return c.CommentPromptTemplate
		}
		// パース成功したテンプレートをキャッシュに保存
		templateCache.Store(templateStr, tmpl)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, article)
	if err != nil {
		// テンプレートの実行に失敗した場合も、元のテンプレートを返す
		return c.CommentPromptTemplate
	}

	return buf.String()
}

type MisskeyConfig struct {
	APIToken string
	APIURL   string
}

// Validate はMisskeyConfigの内容をバリデーションする
func (m *MisskeyConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// APIToken: 必須項目（空文字列でない）
	if err := ValidateRequiredWithDefault(m.APIToken, DefaultMisskeyAPIToken, "Misskey APIトークン"); err != nil {
		builder.AddError(err.Error())
	}

	// APIURL: 必須項目（空文字列でない）、URL形式であること
	if err := ValidateURL(m.APIURL, "Misskey API URL"); err != nil {
		builder.AddError(err.Error())
	}

	return builder.Build()
}

type SlackAPIConfig struct {
	APIToken        string
	Channel         string
	MessageTemplate *string
}

// Validate はSlackAPIConfigの内容をバリデーションする
func (s *SlackAPIConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// APIToken: 必須項目（空文字列でない）
	if err := ValidateRequiredWithDefault(s.APIToken, DefaultSlackAPIToken, "Slack APIトークン"); err != nil {
		builder.AddError(err.Error())
	}

	// Channel: 必須項目（空文字列でない）
	if err := ValidateRequired(s.Channel, "Slackチャンネル"); err != nil {
		builder.AddError(err.Error())
	}

	// MessageTemplate: 存在する場合はtext/templateとして妥当であること
	if s.MessageTemplate != nil {
		if err := s.validateSlackMessageTemplate(*s.MessageTemplate); err != nil {
			builder.AddError(fmt.Sprintf("Slackメッセージテンプレートが無効です: %v", err))
		}
	}

	return builder.Build()
}

// validateSlackMessageTemplate はSlackメッセージテンプレートの構文を検証する
func (s *SlackAPIConfig) validateSlackMessageTemplate(templateStr string) error {
	// 空文字列や空白のみの場合はエラーとしない（デフォルトテンプレートが使用される）
	if strings.TrimSpace(templateStr) == "" {
		return nil
	}

	// text/templateでパースして構文チェック
	_, err := template.New("slack_message").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("テンプレート構文エラー: %w", err)
	}

	return nil
}

type Profile struct {
	AI     *AIConfig
	Prompt *PromptConfig
	Output *OutputConfig
}

// Validate はProfileの内容をバリデーションする
func (p *Profile) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// AI: 必須項目（nilでない）
	if p.AI == nil {
		builder.AddError("AI設定が設定されていません")
	} else {
		builder.MergeResult(p.AI.Validate())
	}

	// Prompt: 必須項目（nilでない）
	if p.Prompt == nil {
		builder.AddError("プロンプト設定が設定されていません")
	} else {
		builder.MergeResult(p.Prompt.Validate())
	}

	// Output: 必須項目（nilでない）
	if p.Output == nil {
		builder.AddError("出力設定が設定されていません")
	} else {
		builder.MergeResult(p.Output.Validate())
	}

	return builder.Build()
}

type OutputConfig struct {
	SlackAPI *SlackAPIConfig
	Misskey  *MisskeyConfig
}

// Validate はOutputConfigの内容をバリデーションする
func (o *OutputConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// SlackAPIとMisskeyの少なくとも一方は設定されている必要がある
	if o.SlackAPI == nil && o.Misskey == nil {
		builder.AddError("SlackAPI設定またはMisskey設定の少なくとも一方が必要です")
	}

	// 設定されているConfigオブジェクトに対してそれぞれのValidate()メソッドを呼び出す
	if o.SlackAPI != nil {
		builder.MergeResult(o.SlackAPI.Validate())
	}

	if o.Misskey != nil {
		builder.MergeResult(o.Misskey.Validate())
	}

	return builder.Build()
}

// ValidationResult はバリデーション結果を表現する
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
