package infra

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

// Config is the root of the configuration structure.
type Config struct {
	General           entity.GeneralConfig               `yaml:"general"`
	Cache             entity.CacheConfig                 `yaml:"cache"`
	AIModels          map[string]entity.AIModelConfig    `yaml:"ai_models"`
	SystemPrompts     map[string]string                  `yaml:"system_prompts"`
	Prompts           map[string]entity.PromptConfig     `yaml:"prompts"`
	Outputs           map[string]entity.OutputConfig     `yaml:"outputs"`
	ExecutionProfiles map[string]entity.ExecutionProfile `yaml:"execution_profiles"`
}

func (c *Config) getDefaultExecutionProfile() (*entity.ExecutionProfile, error) {
	profile, ok := c.ExecutionProfiles[c.General.DefaultExecutionProfile]
	if !ok {
		return nil, fmt.Errorf("default execution profile not found: %s", c.General.DefaultExecutionProfile)
	}
	return &profile,
		nil
}

func (c *Config) GetDefaultAIModel() (*entity.AIModelConfig, error) {
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

func (c *Config) GetDefaultPrompt() (*entity.PromptConfig, error) {
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

func (c *Config) GetDefaultOutputs() ([]*entity.OutputConfig, error) {
	profile, err := c.getDefaultExecutionProfile()
	if err != nil {
		return nil, err
	}

	outputs := make([]*entity.OutputConfig, 0, len(profile.Outputs))
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

func MakeDefaultConfig() *Config {
	return &Config{
		General: entity.GeneralConfig{
			DefaultExecutionProfile: "任意のプロファイル名",
		},
		Cache: entity.CacheConfig{
			RetentionDays: 7,
		},
		AIModels: map[string]entity.AIModelConfig{
			"任意のAIモデル名": {
				Type:   "gemini-2.5-flash または gemini-2.5-pro",
				APIKey: "xxxxxx",
			},
		},
		SystemPrompts: map[string]string{
			"任意のシステムプロンプト名": "あなたはXXXXなAIアシスタントです。",
		},
		Prompts: map[string]entity.PromptConfig{
			"任意のプロンプト名": {
				CommentPromptTemplate: `以下の記事の紹介文を100字以内で作成してください。\n---\n記事タイトル: {{title}}\n記事URL: {{url}}\n記事内容:\n{{content}}`,
			},
		},
		Outputs: map[string]entity.OutputConfig{
			"任意の出力名(Slack)": {
				Type: "slack-api",
				SlackAPIConfig: &entity.SlackAPIConfig{
					APIToken: "xoxb-xxxxxx",
					Channel:  "#general",
				},
			},
			"任意の出力名(Misskey)": {
				Type: "misskey",
				MisskeyConfig: &entity.MisskeyConfig{
					APIToken: "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
					APIURL:   "https://misskey.social/api",
				},
			},
		},
		ExecutionProfiles: map[string]entity.ExecutionProfile{
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

func NewYamlConfigRepository(filePath string) domain.ConfigRepository {
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
		return nil, fmt.Errorf("failed to read config file: %s, %w", r.filePath, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
