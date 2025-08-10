package entity

import (
	"bytes"
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
