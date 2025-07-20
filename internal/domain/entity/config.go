package entity

import "fmt"

// Config is the root of the configuration structure.
type Config struct {
	General           GeneralConfig               `mapstructure:"general"`
	Cache             CacheConfig                 `mapstructure:"cache"`
	AIModels          map[string]AIModelConfig    `mapstructure:"ai_models"`
	Prompts           map[string]PromptConfig     `mapstructure:"prompts"`
	Outputs           map[string]OutputConfig     `mapstructure:"outputs"`
	ExecutionProfiles map[string]ExecutionProfile `mapstructure:"execution_profiles"`
}

func (c *Config) GetDefaultAIModel() (*AIModelConfig, error) {
	prpfile, ok := c.ExecutionProfiles[c.General.DefaultExecutionProfile]
	if !ok {
		return nil, fmt.Errorf("default execution profile not found: %s", c.General.DefaultExecutionProfile)
	}

	model, ok := c.AIModels[prpfile.AIModel]
	if !ok {
		return nil, fmt.Errorf("AI model not found: %s", prpfile.AIModel)
	}

	return &model, nil
}

func (c *Config) GetDefaultPrompt() (*PromptConfig, error) {
	prpfile, ok := c.ExecutionProfiles[c.General.DefaultExecutionProfile]
	if !ok {
		return nil, fmt.Errorf("default execution profile not found: %s", c.General.DefaultExecutionProfile)
	}

	prompt, ok := c.Prompts[prpfile.Prompt]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", prpfile.Prompt)
	}

	return &prompt, nil
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DefaultExecutionProfile string `mapstructure:"default_execution_profile"`
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	RetentionDays int `mapstructure:"retention_days"`
}

// AIModelConfig holds configuration for a specific AI model.
type AIModelConfig struct {
	Type   string `mapstructure:"type"`
	APIKey string `mapstructure:"api_key"`
}

// PromptConfig holds configuration for a specific prompt.
type PromptConfig struct {
	SystemMessage         string `mapstructure:"system_message"`
	CommentPromptTemplate string `mapstructure:"comment_prompt_template"`
}

// OutputConfig holds configuration for a specific output destination.
type OutputConfig struct {
	Type       string `mapstructure:"type"`
	WebhookURL string `mapstructure:"webhook_url,omitempty"`
	Channel    string `mapstructure:"channel,omitempty"`
	Username   string `mapstructure:"username,omitempty"`
	IconEmoji  string `mapstructure:"icon_emoji,omitempty"`
	APIURL     string `mapstructure:"api_url,omitempty"`
	APIToken   string `mapstructure:"api_token,omitempty"`
	Visibility string `mapstructure:"visibility,omitempty"`
}

// ExecutionProfile defines a combination of AI model, prompt, and output.
type ExecutionProfile struct {
	AIModel string `mapstructure:"ai_model,omitempty"`
	Prompt  string `mapstructure:"prompt,omitempty"`
	Output  string `mapstructure:"output"`
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
		Prompts: map[string]PromptConfig{
			"任意のプロンプト名": {
				SystemMessage: "あなたはXXXXなAIアシスタントです。",
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
				Type:       "slack",
				WebhookURL: "https://hooks.slack.com/services/TXXXXX/BXXXXX/YYYYYYYYYYYYYYYYYYYYYYYY",
				Channel:    "#general",
				Username:   "ai-feed-bot",
				IconEmoji:  ":robot_face:",
			},
			"任意の出力名(Misskey)": {
				Type:       "misskey",
				APIURL:     "https://misskey.social/api",
				APIToken:   "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
				Visibility: "public",
			},
		},
		ExecutionProfiles: map[string]ExecutionProfile{
			"任意のプロファイル名": {
				AIModel: "任意のAIモデル名",
				Prompt:  "任意のプロンプト名",
				Output:  "任意の出力名",
			},
		},
	}
}
