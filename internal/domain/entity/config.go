package entity

import (
	"fmt"
	"strings"
)

// Config is the root of the configuration structure.
type Config struct {
	General           GeneralConfig               `yaml:"general"`
	Cache             CacheConfig                 `yaml:"cache"`
	AIModels          map[string]AIModelConfig    `yaml:"ai_models"`
	SystemPrompts     map[string]string           `yaml:"system_prompts"`
	Prompts           map[string]PromptConfig     `yaml:"prompts"`
	Outputs           map[string]OutputConfig     `yaml:"outputs"`
	ExecutionProfiles map[string]ExecutionProfile `yaml:"execution_profiles"`
}

func (c *Config) getDefaultExecutionProfile() (*ExecutionProfile, error) {
	profile, ok := c.ExecutionProfiles[c.General.DefaultExecutionProfile]
	if !ok {
		return nil, fmt.Errorf("default execution profile not found: %s", c.General.DefaultExecutionProfile)
	}
	return &profile, nil
}

func (c *Config) GetDefaultAIModel() (*AIModelConfig, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return nil, err
	}
	if profile.AIModel == "" {
		return nil, nil
	}

	model, ok := c.AIModels[profile.AIModel]
	if !ok {
		return nil, fmt.Errorf("AI model not found: %s", profile.AIModel)
	}

	return &model, nil
}

func (c *Config) GetDefaultPrompt() (*PromptConfig, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return nil, err
	}
	if profile.Prompt == "" {
		return nil, nil
	}

	prompt, ok := c.Prompts[profile.Prompt]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", profile.Prompt)
	}

	return &prompt, nil
}

func (c *Config) GetDefaultSystemPrompt() (string, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return "", err
	}
	if profile.SystemPrompt == "" {
		return "", nil
	}

	systemPrompt, ok := c.SystemPrompts[profile.SystemPrompt]
	if !ok {
		return "", fmt.Errorf("system prompt not found: %s", profile.SystemPrompt)
	}

	return systemPrompt, nil
}

func (c *Config) GetDefaultOutputs() ([]OutputConfig, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return nil, err
	}

	outputs := make([]OutputConfig, 0, len(profile.Outputs))
	for _, outputName := range profile.Outputs {
		output, outputFound := c.Outputs[outputName]
		if !outputFound {
			return nil, fmt.Errorf("output not found: %s", outputName)
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

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

// MarshalYAML implements the yaml.Marshaler interface.
func (o OutputConfig) MarshalYAML() (interface{}, error) {
	// Create a map to hold the marshaled data
	m := make(map[string]interface{})
	m["type"] = o.Type

	switch o.Type {
	case "misskey":
		if o.MisskeyConfig != nil {
			// Marshal MisskeyConfig into the map
			misskeyMap, err := yaml.Marshal(o.MisskeyConfig)
			if err != nil {
				return nil, err
			}
			var temp map[string]interface{}
			if err := yaml.Unmarshal(misskeyMap, &temp); err != nil {
				return nil, err
			}
			for k, v := range temp {
				m[k] = v
			}
		}
	case "slack-api":
		if o.SlackAPIConfig != nil {
			// Marshal SlackAPIConfig into the map
			slackAPIMap, err := yaml.Marshal(o.SlackAPIConfig)
			if err != nil {
				return nil, err
			}
			var temp map[string]interface{}
			if err := yaml.Unmarshal(slackAPIMap, &temp); err != nil {
				return nil, err
			}
			for k, v := range temp {
				m[k] = v
			}
		}
	}
	return m, nil
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

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (o *OutputConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First, unmarshal into a temporary struct to get the 'type' field.
	var raw struct {
		Type string `yaml:"type"`
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	o.Type = raw.Type

	// Then, unmarshal into the specific config type based on 'type' field.
	switch o.Type {
	case "misskey":
		var misskey MisskeyConfig
		if err := unmarshal(&misskey); err != nil {
			return err
		}
		o.MisskeyConfig = &misskey
	case "slack-api":
		var slackAPI SlackAPIConfig
		if err := unmarshal(&slackAPI); err != nil {
			return err
		}
		o.SlackAPIConfig = &slackAPI
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

func MakeDefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			DefaultExecutionProfile: "任意のプロファイル名",
		},
		Cache: CacheConfig{
			RetentionDays: 7,
		},
		AIModels: map[string]AIModelConfig{
			"任意のAIモデル名": {
				Type:   "gemini-2.5-flash または gemini-2.5-pro",
				APIKey: "xxxxxx",
			},
		},
		SystemPrompts: map[string]string{
			"任意のシステムプロンプト名": "あなたはXXXXなAIアシスタントです。",
		},
		Prompts: map[string]PromptConfig{
			"任意のプロンプト名": {
				CommentPromptTemplate: `以下の記事の紹介文を100字以内で作成してください。
---
記事タイトル: {{title}}
記事URL: {{url}}
記事内容:
{{content}}`,
			},
		},
		Outputs: map[string]OutputConfig{
			"任意の出力名(Slack)": {
				Type:     "slack-api",
				APIToken: "xoxb-xxxxxx",
				Channel:  "#general",
			},
			"任意の出力名(Misskey)": {
				Type:     "misskey",
				APIToken: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
				APIURL:   "https://misskey.social/api",
			},
		},
		ExecutionProfiles: map[string]ExecutionProfile{
			"任意のプロファイル名": {
				AIModel:      "任意のAIモデル名",
				SystemPrompt: "任意のシステムプロンプト名",
				Prompt:       "任意のプロンプト名",
				Outputs:      []string{"任意の出力名(Slack)", "任意の出力名(Misskey)"},
			},
		},
	}
}
