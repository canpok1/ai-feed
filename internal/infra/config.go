package infra

//go:generate mockgen -source=config.go -destination=mock_infra/mock_config.go -package=mock_infra ConfigRepository

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProfile *Profile `yaml:"default_profile,omitempty"`
}

type Profile struct {
	AI     *AIConfig     `yaml:"ai,omitempty"`
	Prompt *PromptConfig `yaml:",inline,omitempty"`
	Output *OutputConfig `yaml:"output,omitempty"`
}

type AIConfig struct {
	Gemini *GeminiConfig `yaml:"gemini,omitempty"`
}

func (c *AIConfig) ToEntity() *entity.AIConfig {
	return &entity.AIConfig{
		Gemini: c.Gemini.ToEntity(),
	}
}

type GeminiConfig struct {
	Type   string `yaml:"type"`
	APIKey string `yaml:"api_key"`
}

func (c *GeminiConfig) ToEntity() *entity.GeminiConfig {
	return &entity.GeminiConfig{
		Type:   c.Type,
		APIKey: c.APIKey,
	}
}

type PromptConfig struct {
	SystemPrompt          string `yaml:"system_prompt,omitempty"`
	CommentPromptTemplate string `yaml:"comment_prompt_template,omitempty"`
}

func (c *PromptConfig) ToEntity() *entity.PromptConfig {
	return &entity.PromptConfig{
		SystemPrompt:          c.SystemPrompt,
		CommentPromptTemplate: c.CommentPromptTemplate,
	}
}

type SystemPromptConfig struct {
	Value string `yaml:"value"`
}

type OutputConfig struct {
	SlackAPI *SlackAPIConfig `yaml:"slack_api,omitempty"`
	Misskey  *MisskeyConfig  `yaml:"misskey,omitempty"`
}

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

type MisskeyConfig struct {
	APIToken string `yaml:"api_token"`
	APIURL   string `yaml:"api_url"`
}

func MakeDefaultConfig() *Config {
	return &Config{
		DefaultProfile: &Profile{
			AI: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "gemini-2.5-flash",
					APIKey: "xxxxxx",
				},
			},
			Prompt: &PromptConfig{
				SystemPrompt: "あなたはXXXXなAIアシスタントです。",
				CommentPromptTemplate: `以下の記事の紹介文を100字以内で作成してください。
---
記事タイトル: {{title}}
記事URL: {{url}}
記事内容:
{{content}}`,
			},
			Output: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					APIToken: "xoxb-xxxxxx",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					APIToken: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
					APIURL:   "https://misskey.social/api",
				},
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
