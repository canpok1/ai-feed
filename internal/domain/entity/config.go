package entity

// Config is the root of the configuration structure.
type Config struct {
	General           GeneralConfig               `yaml:"general"`
	Cache             CacheConfig                 `yaml:"cache"`
	AIModels          map[string]AIModelConfig    `yaml:"ai_models"`
	Prompts           map[string]PromptConfig     `yaml:"prompts"`
	Outputs           map[string]OutputConfig     `yaml:"outputs"`
	ExecutionProfiles map[string]ExecutionProfile `yaml:"execution_profiles"`
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
	SystemMessage         string `yaml:"system_message"`
	CommentPromptTemplate string `yaml:"comment_prompt_template"`
}

// OutputConfig holds configuration for a specific output destination.
type OutputConfig struct {
	Type       string `yaml:"type"`
	WebhookURL string `yaml:"webhook_url,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
	APIURL     string `yaml:"api_url,omitempty"`
	APIToken   string `yaml:"api_token,omitempty"`
	Visibility string `yaml:"visibility,omitempty"`
}

// ExecutionProfile defines a combination of AI model, prompt, and output.
type ExecutionProfile struct {
	AIModel string `yaml:"ai_model,omitempty"`
	Prompt  string `yaml:"prompt,omitempty"`
	Output  string `yaml:"output"`
}

func MakeDefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			DefaultExecutionProfile: "プロファイル名",
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
				AIModel: "AIモデル名",
				Prompt:  "プロンプト名",
				Output:  "出力名",
			},
		},
	}
}
