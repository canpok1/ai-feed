package entity

import (
	"strings"
)

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DefaultExecutionProfile string
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	RetentionDays int
}

// AIModelConfig holds configuration for a specific AI model.
type AIModelConfig struct {
	Type   string
	APIKey string
}

// PromptConfig holds configuration for a specific prompt.
type PromptConfig struct {
	CommentPromptTemplate string
}

func (c *PromptConfig) MakeCommentPromptTemplate(article *Article) string {
	prompt := c.CommentPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{title}}", article.Title)
	prompt = strings.ReplaceAll(prompt, "{{url}}", article.Link)
	prompt = strings.ReplaceAll(prompt, "{{content}}", article.Content)
	return prompt
}

// OutputConfig holds configuration for a specific output destination.
// This struct is used for initial unmarshaling to determine the type.
type OutputConfig struct {
	Type string

	// Specific configurations, to be populated by UnmarshalYAML
	MisskeyConfig  *MisskeyConfig
	SlackAPIConfig *SlackAPIConfig
}

// MisskeyConfig holds configuration for Misskey output.
type MisskeyConfig struct {
	APIToken string
	APIURL   string
}

// SlackAPIConfig holds configuration for Slack API output.
type SlackAPIConfig struct {
	APIToken string
	Channel  string
}

// ExecutionProfile defines a combination of AI model, prompt, and output.
type ExecutionProfile struct {
	AIModel      string
	SystemPrompt string
	Prompt       string
	Outputs      []string
}
