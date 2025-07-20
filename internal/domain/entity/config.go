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
		AIModels:          make(map[string]AIModelConfig),
		Prompts:           make(map[string]PromptConfig),
		Outputs:           make(map[string]OutputConfig),
		ExecutionProfiles: make(map[string]ExecutionProfile),
	}
}
