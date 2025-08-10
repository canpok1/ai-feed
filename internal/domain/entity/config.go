package entity

import (
	"bytes"
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
	tmpl, err := template.New("comment").Parse(c.CommentPromptTemplate)
	if err != nil {
		// テンプレート解析エラーの場合は空文字列を返す
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, article)
	if err != nil {
		// テンプレート実行エラーの場合も空文字列を返す
		return ""
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
