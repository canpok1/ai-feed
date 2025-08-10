package entity

import (
	"bytes"
	"fmt"
	"net/url"
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
	var errors []string
	var warnings []string

	// Gemini: 必須項目（nilでない）
	if a.Gemini == nil {
		errors = append(errors, "Gemini設定が設定されていません")
	} else {
		// GeminiConfig.Validate()を呼び出して、結果を集約
		geminiResult := a.Gemini.Validate()
		if !geminiResult.IsValid {
			errors = append(errors, geminiResult.Errors...)
		}
		warnings = append(warnings, geminiResult.Warnings...)
	}

	return &ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

type GeminiConfig struct {
	Type   string
	APIKey string
}

// Validate はGeminiConfigの内容をバリデーションする
func (g *GeminiConfig) Validate() *ValidationResult {
	var errors []string

	// Type: 必須項目（空文字列でない）
	if g.Type == "" {
		errors = append(errors, "Gemini設定のTypeが設定されていません")
	}

	// APIKey: 必須項目（空文字列でない）
	if g.APIKey == "" || g.APIKey == DefaultGeminiAPIKey {
		errors = append(errors, "Gemini API keyが設定されていません")
	}

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

type PromptConfig struct {
	SystemPrompt          string
	CommentPromptTemplate string
	FixedMessage          string
}

// Validate はPromptConfigの内容をバリデーションする
func (p *PromptConfig) Validate() *ValidationResult {
	var errors []string

	// SystemPrompt: 必須項目（空文字列でない）
	if p.SystemPrompt == "" {
		errors = append(errors, "システムプロンプトが設定されていません")
	}

	// CommentPromptTemplate: 必須項目（空文字列でない）
	if p.CommentPromptTemplate == "" {
		errors = append(errors, "コメントプロンプトテンプレートが設定されていません")
	}

	// FixedMessage: 任意項目（空文字列でも可）

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
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
	var errors []string

	// APIToken: 必須項目（空文字列でない）
	if m.APIToken == "" || m.APIToken == DefaultMisskeyAPIToken {
		errors = append(errors, "Misskey APIトークンが設定されていません")
	}

	// APIURL: 必須項目（空文字列でない）、URL形式であること
	if m.APIURL == "" {
		errors = append(errors, "Misskey API URLが設定されていません")
	} else {
		// URL形式チェック
		if _, err := url.Parse(m.APIURL); err != nil {
			errors = append(errors, "Misskey API URLが正しいURL形式ではありません")
		}
	}

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

type SlackAPIConfig struct {
	APIToken        string
	Channel         string
	MessageTemplate *string
}

// Validate はSlackAPIConfigの内容をバリデーションする
func (s *SlackAPIConfig) Validate() *ValidationResult {
	var errors []string

	// APIToken: 必須項目（空文字列でない）
	if s.APIToken == "" || s.APIToken == DefaultSlackAPIToken {
		errors = append(errors, "Slack APIトークンが設定されていません")
	}

	// Channel: 必須項目（空文字列でない）
	if s.Channel == "" {
		errors = append(errors, "Slackチャンネルが設定されていません")
	}

	// MessageTemplate: 存在する場合はtext/templateとして妥当であること
	if s.MessageTemplate != nil {
		if err := s.validateSlackMessageTemplate(*s.MessageTemplate); err != nil {
			errors = append(errors, fmt.Sprintf("Slackメッセージテンプレートが無効です: %v", err))
		}
	}

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
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

type OutputConfig struct {
	SlackAPI *SlackAPIConfig
	Misskey  *MisskeyConfig
}

// ValidationResult はバリデーション結果を表現する
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
