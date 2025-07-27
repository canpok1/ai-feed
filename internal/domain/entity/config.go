package entity

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DefaultExecutionProfile string `yaml:"default_execution_profile"`
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	RetentionDays int `yaml:"retention_days"`
}

// AIModelConfig holds configuration for a specific AI model.
type AIModelConfig struct {
	Type   string `yaml:"type"`
	APIKey string `yaml:"api_key"`
}

// PromptConfig holds configuration for a specific prompt.
type PromptConfig struct {
	CommentPromptTemplate string `yaml:"comment_prompt_template"`
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
	Type string `yaml:"type"`

	// Specific configurations, to be populated by UnmarshalYAML
	MisskeyConfig  *MisskeyConfig
	SlackAPIConfig *SlackAPIConfig
}

// MisskeyConfig holds configuration for Misskey output.
type MisskeyConfig struct {
	APIToken string `yaml:"api_token"`
	APIURL   string `yaml:"api_url"`
}

// SlackAPIConfig holds configuration for Slack API output.
type SlackAPIConfig struct {
	APIToken string `yaml:"api_token"`
	Channel  string `yaml:"channel"`
}

// MarshalYAML implements the yaml.Marshaler interface.
func (o OutputConfig) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{})
	m["type"] = o.Type

	var configData interface{}
	switch o.Type {
	case "misskey":
		configData = o.MisskeyConfig
	case "slack-api":
		configData = o.SlackAPIConfig
	default:
		return nil, fmt.Errorf("unsupported output type: %s", o.Type)
	}

	if configData != nil {
		out, err := yaml.Marshal(configData)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(out, &m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (o *OutputConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"]
	if !ok {
		return fmt.Errorf("type field not found in OutputConfig")
	}
	o.Type, ok = typeVal.(string)
	if !ok {
		return fmt.Errorf("type field is not a string")
	}

	switch o.Type {
	case "misskey":
		misskeyConfig := MisskeyConfig{}
		if err := value.Decode(&misskeyConfig); err != nil {
			return err
		}
		o.MisskeyConfig = &misskeyConfig
	case "slack-api":
		slackAPIConfig := SlackAPIConfig{}
		if err := value.Decode(&slackAPIConfig); err != nil {
			return err
		}
		o.SlackAPIConfig = &slackAPIConfig
	default:
		return fmt.Errorf("unsupported output type: %s", o.Type)
	}

	return nil
}

// ExecutionProfile defines a combination of AI model, prompt, and output.
type ExecutionProfile struct {
	AIModel      string   `yaml:"ai_model,omitempty"`
	SystemPrompt string   `yaml:"system_prompt,omitempty"`
	Prompt       string   `yaml:"prompt,omitempty"`
	Outputs      []string `yaml:"outputs"`
}
