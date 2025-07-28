package infra

//go:generate mockgen -source=config.go -destination=mock_infra/mock_config.go -package=mock_infra ConfigRepository

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

// ConfigRepository defines the interface for configuration operations.
type ConfigRepository interface {
	GetDefaultAIModel() (*AIModelConfig, error)
	GetDefaultPrompt() (*PromptConfig, error)
	GetDefaultSystemPrompt() (string, error)
	GetDefaultOutputs() ([]*OutputConfig, error)
}

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
	return &profile,
		nil
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
		return "", fmt.Errorf("system prompt not found: %s", systemPrompt)
	}

	return systemPrompt, nil
}

func (c *Config) GetDefaultOutputs() ([]*OutputConfig, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return nil, err
	}

	outputs := make([]*OutputConfig, 0, len(profile.Outputs))
	for _, outputName := range profile.Outputs {
		output, outputFound := c.Outputs[outputName]
		if !outputFound {
			return nil, fmt.Errorf("output not found: %s", outputName)
		}
		o := output // Create a new variable to ensure we get a unique pointer for each iteration.
		outputs = append(outputs, &o)
	}

	return outputs, nil
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DefaultExecutionProfile string `yaml:"default_execution_profile"`
}

func (c *GeneralConfig) ToEntity() *entity.GeneralConfig {
	return &entity.GeneralConfig{
		DefaultExecutionProfile: c.DefaultExecutionProfile,
	}
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	RetentionDays int `yaml:"retention_days"`
}

func (c *CacheConfig) ToEntity() *entity.CacheConfig {
	return &entity.CacheConfig{
		RetentionDays: c.RetentionDays,
	}
}

// AIModelConfig holds configuration for a specific AI model.
type AIModelConfig struct {
	Type   string `yaml:"type"`
	APIKey string `yaml:"api_key"`
}

func (c *AIModelConfig) ToEntity() *entity.AIModelConfig {
	return &entity.AIModelConfig{
		Type:   c.Type,
		APIKey: c.APIKey,
	}
}

// PromptConfig holds configuration for a specific prompt.
type PromptConfig struct {
	CommentPromptTemplate string `yaml:"comment_prompt_template"`
}

func (c *PromptConfig) ToEntity() *entity.PromptConfig {
	return &entity.PromptConfig{
		CommentPromptTemplate: c.CommentPromptTemplate,
	}
}

// OutputConfig holds configuration for a specific output destination.
// This struct is used for initial unmarshaling to determine the type.
type OutputConfig struct {
	Type string `yaml:"type"`

	// Specific configurations, to be populated by UnmarshalYAML
	MisskeyConfig  *MisskeyConfig
	SlackAPIConfig *SlackAPIConfig
}

func (c *OutputConfig) ToEntity() *entity.OutputConfig {
	e := &entity.OutputConfig{
		Type: c.Type,
	}
	if c.MisskeyConfig != nil {
		e.MisskeyConfig = c.MisskeyConfig.ToEntity()
	}
	if c.SlackAPIConfig != nil {
		e.SlackAPIConfig = c.SlackAPIConfig.ToEntity()
	}
	return e
}

// MisskeyConfig holds configuration for Misskey output.
type MisskeyConfig struct {
	APIToken string `yaml:"api_token"`
	APIURL   string `yaml:"api_url"`
}

func (c *MisskeyConfig) ToEntity() *entity.MisskeyConfig {
	return &entity.MisskeyConfig{
		APIToken: c.APIToken,
		APIURL:   c.APIURL,
	}
}

// SlackAPIConfig holds configuration for Slack API output.
type SlackAPIConfig struct {
	APIToken string `yaml:"api_token"`
	Channel  string `yaml:"channel"`
}

func (c *SlackAPIConfig) ToEntity() *entity.SlackAPIConfig {
	return &entity.SlackAPIConfig{
		APIToken: c.APIToken,
		Channel:  c.Channel,
	}
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

func (p *ExecutionProfile) ToEntity() *entity.ExecutionProfile {
	return &entity.ExecutionProfile{
		AIModel:      p.AIModel,
		SystemPrompt: p.SystemPrompt,
		Prompt:       p.Prompt,
		Outputs:      p.Outputs,
	}
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
				Type: "slack-api",
				SlackAPIConfig: &SlackAPIConfig{
					APIToken: "xoxb-xxxxxx",
					Channel:  "#general",
				},
			},
			"任意の出力名(Misskey)": {
				Type: "misskey",
				MisskeyConfig: &MisskeyConfig{
					APIToken: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
					APIURL:   "https://misskey.social/api",
				},
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

type YamlConfigRepository struct {
	filePath string
}

func NewYamlConfigRepository(filePath string) *YamlConfigRepository {
	return &YamlConfigRepository{
		filePath: filePath,
	}
}

func (r *YamlConfigRepository) Save(config *Config) error {
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("config file already exists: %s", r.filePath)
		}
		return fmt.Errorf("failed to create config file: %s, %w", r.filePath, err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func (r *YamlConfigRepository) Load() (*Config, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
