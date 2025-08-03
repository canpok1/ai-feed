package entity

import (
	"strings"
)

// デフォルト値の定数
const (
	DefaultGeminiAPIKey       = "YOUR_GEMINI_API_KEY_HERE"
	DefaultSlackAPIToken      = "xoxb-YOUR_SLACK_API_TOKEN_HERE"
	DefaultMisskeyAPIToken    = "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE"
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

func (c *PromptConfig) BuildCommentPrompt(article *Article) string {
	prompt := c.CommentPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{title}}", article.Title)
	prompt = strings.ReplaceAll(prompt, "{{url}}", article.Link)
	prompt = strings.ReplaceAll(prompt, "{{content}}", article.Content)
	return prompt
}

type MisskeyConfig struct {
	APIToken string
	APIURL   string
}

type SlackAPIConfig struct {
	APIToken string
	Channel  string
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
