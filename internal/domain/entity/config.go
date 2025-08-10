package entity

import (
	"bytes"
	"strings"
	"text/template"
)

// デフォルト値の定数
const (
	DefaultGeminiAPIKey    = "YOUR_GEMINI_API_KEY_HERE"
	DefaultSlackAPIToken   = "xoxb-YOUR_SLACK_API_TOKEN_HERE"
	DefaultMisskeyAPIToken = "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE"
)

type AIConfig struct {
	Gemini *GeminiConfig
}

type GeminiConfig struct {
	Type   string
	APIKey string
}

type PromptConfig struct {
	SystemPrompt          string
	CommentPromptTemplate string
	FixedMessage          string
}

// BuildCommentPrompt はtext/templateを使用してコメントプロンプトを生成する
func (c *PromptConfig) BuildCommentPrompt(article *Article) string {
	// 互換性のため、古い形式のプレースホルダーを新形式に変換
	templateStr := c.CommentPromptTemplate
	templateStr = strings.ReplaceAll(templateStr, "{{title}}", "{{.Title}}")
	templateStr = strings.ReplaceAll(templateStr, "{{url}}", "{{.Link}}")
	templateStr = strings.ReplaceAll(templateStr, "{{content}}", "{{.Content}}")

	// text/templateを使用してテンプレートを解析・実行
	tmpl, err := template.New("comment").Parse(templateStr)
	if err != nil {
		// テンプレート解析エラーの場合は、フォールバック処理
		// 元の文字列置換方式で処理
		prompt := c.CommentPromptTemplate
		prompt = strings.ReplaceAll(prompt, "{{title}}", article.Title)
		prompt = strings.ReplaceAll(prompt, "{{url}}", article.Link)
		prompt = strings.ReplaceAll(prompt, "{{content}}", article.Content)
		return prompt
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, article)
	if err != nil {
		// テンプレート実行エラーの場合も、フォールバック処理
		prompt := c.CommentPromptTemplate
		prompt = strings.ReplaceAll(prompt, "{{title}}", article.Title)
		prompt = strings.ReplaceAll(prompt, "{{url}}", article.Link)
		prompt = strings.ReplaceAll(prompt, "{{content}}", article.Content)
		return prompt
	}

	return buf.String()
}

type MisskeyConfig struct {
	APIToken string
	APIURL   string
}

type SlackAPIConfig struct {
	APIToken        string
	Channel         string
	MessageTemplate *string
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
