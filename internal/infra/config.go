package infra

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

// Merge merges the non-nil fields from the other profile into the current profile.
func (p *Profile) Merge(other *Profile) {
	if other == nil {
		return
	}
	mergePtr(&p.AI, other.AI)
	mergePtr(&p.Prompt, other.Prompt)
	mergePtr(&p.Output, other.Output)
}

// ToEntity converts infra.Profile to entity.Profile
func (p *Profile) ToEntity() (*entity.Profile, error) {
	var aiEntity *entity.AIConfig
	if p.AI != nil {
		var err error
		aiEntity, err = p.AI.ToEntity()
		if err != nil {
			return nil, err
		}
	}

	var promptEntity *entity.PromptConfig
	if p.Prompt != nil {
		promptEntity = p.Prompt.ToEntity()
	}

	var outputEntity *entity.OutputConfig
	if p.Output != nil {
		var err error
		outputEntity, err = p.Output.ToEntity()
		if err != nil {
			return nil, err
		}
	}

	return &entity.Profile{
		AI:     aiEntity,
		Prompt: promptEntity,
		Output: outputEntity,
	}, nil
}

type AIConfig struct {
	Gemini *GeminiConfig `yaml:"gemini,omitempty"`
}

func (c *AIConfig) Merge(other *AIConfig) {
	if other == nil {
		return
	}
	mergePtr(&c.Gemini, other.Gemini)
}

func (c *AIConfig) ToEntity() (*entity.AIConfig, error) {
	var geminiEntity *entity.GeminiConfig
	if c.Gemini != nil {
		var err error
		geminiEntity, err = c.Gemini.ToEntity()
		if err != nil {
			return nil, err
		}
	}
	return &entity.AIConfig{
		Gemini: geminiEntity,
	}, nil
}

type GeminiConfig struct {
	Type      string `yaml:"type"`
	APIKey    string `yaml:"api_key"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"`
}

func (c *GeminiConfig) Merge(other *GeminiConfig) {
	if other == nil {
		return
	}
	mergeString(&c.Type, other.Type)

	// プロファイルファイルでapi_key_envが指定されている場合、
	// デフォルトのapi_keyを無効にして環境変数を優先する
	if other.APIKeyEnv != "" && other.APIKey == "" {
		c.APIKey = ""
		c.APIKeyEnv = other.APIKeyEnv
	} else {
		// 通常のマージ処理
		mergeString(&c.APIKey, other.APIKey)
		mergeString(&c.APIKeyEnv, other.APIKeyEnv)
	}
}

// resolveSecret は、直接指定された値または環境変数から値を解決する
func resolveSecret(value, envVar, configPath string) (string, error) {
	if value != "" {
		return value, nil
	}
	if envVar != "" {
		secret := os.Getenv(envVar)
		if secret == "" {
			return "", fmt.Errorf("環境変数 '%s' が設定されていません。%s で指定された環境変数を設定してください。", envVar, configPath)
		}
		return secret, nil
	}
	return "", nil
}

// resolveEnabled は、Enabledフィールドのデフォルト値処理を行う
func resolveEnabled(e *bool) bool {
	if e == nil {
		return true
	}
	return *e
}

func (c *GeminiConfig) ToEntity() (*entity.GeminiConfig, error) {
	apiKey, err := resolveSecret(c.APIKey, c.APIKeyEnv, "ai.gemini.api_key_env")
	if err != nil {
		return nil, err
	}

	return &entity.GeminiConfig{
		Type:   c.Type,
		APIKey: apiKey,
	}, nil
}

type PromptConfig struct {
	SystemPrompt          string `yaml:"system_prompt,omitempty"`
	CommentPromptTemplate string `yaml:"comment_prompt_template,omitempty"`
	FixedMessage          string `yaml:"fixed_message,omitempty"`
}

func (c *PromptConfig) Merge(other *PromptConfig) {
	if other == nil {
		return
	}
	mergeString(&c.SystemPrompt, other.SystemPrompt)
	mergeString(&c.CommentPromptTemplate, other.CommentPromptTemplate)
	mergeString(&c.FixedMessage, other.FixedMessage)
}

func (c *PromptConfig) ToEntity() *entity.PromptConfig {
	return &entity.PromptConfig{
		SystemPrompt:          c.SystemPrompt,
		CommentPromptTemplate: c.CommentPromptTemplate,
		FixedMessage:          c.FixedMessage,
	}
}

type OutputConfig struct {
	SlackAPI *SlackAPIConfig `yaml:"slack_api,omitempty"`
	Misskey  *MisskeyConfig  `yaml:"misskey,omitempty"`
}

func (c *OutputConfig) Merge(other *OutputConfig) {
	if other == nil {
		return
	}
	mergePtr(&c.SlackAPI, other.SlackAPI)
	mergePtr(&c.Misskey, other.Misskey)
}

func (c *OutputConfig) ToEntity() (*entity.OutputConfig, error) {
	var slackEntity *entity.SlackAPIConfig
	if c.SlackAPI != nil {
		var err error
		slackEntity, err = c.SlackAPI.ToEntity()
		if err != nil {
			return nil, err
		}
	}

	var misskeyEntity *entity.MisskeyConfig
	if c.Misskey != nil {
		var err error
		misskeyEntity, err = c.Misskey.ToEntity()
		if err != nil {
			return nil, err
		}
	}

	return &entity.OutputConfig{
		SlackAPI: slackEntity,
		Misskey:  misskeyEntity,
	}, nil
}

// convertMessageTemplate は、メッセージテンプレートの別名変換処理を行う共通ヘルパー関数
func convertMessageTemplate(template *string, converter *entity.TemplateAliasConverter) (*string, error) {
	if template != nil && *template != "" {
		converted, err := converter.Convert(*template)
		if err != nil {
			// 別名変換エラーの場合は、エラーをラップして返す
			return nil, fmt.Errorf("テンプレートエラー: %w", err)
		}
		return &converted, nil
	}
	return template, nil
}

type SlackAPIConfig struct {
	Enabled         *bool   `yaml:"enabled,omitempty"`
	APIToken        string  `yaml:"api_token"`
	APITokenEnv     string  `yaml:"api_token_env,omitempty"`
	Channel         string  `yaml:"channel"`
	MessageTemplate *string `yaml:"message_template,omitempty"`
}

func (c *SlackAPIConfig) Merge(other *SlackAPIConfig) {
	if other == nil {
		return
	}

	// Enabledフィールドのマージ（*boolポインタのマージ）
	if other.Enabled != nil {
		c.Enabled = other.Enabled
	}

	// プロファイルファイルでapi_token_envが指定されている場合、
	// デフォルトのapi_tokenを無効にして環境変数を優先する
	if other.APITokenEnv != "" && other.APIToken == "" {
		c.APIToken = ""
		c.APITokenEnv = other.APITokenEnv
	} else {
		// 通常のマージ処理
		mergeString(&c.APIToken, other.APIToken)
		mergeString(&c.APITokenEnv, other.APITokenEnv)
	}

	mergeString(&c.Channel, other.Channel)
	if other.MessageTemplate != nil {
		c.MessageTemplate = other.MessageTemplate
	}
}

func (c *SlackAPIConfig) ToEntity() (*entity.SlackAPIConfig, error) {
	// Enabledフィールドの後方互換性処理（省略時=true）
	enabled := resolveEnabled(c.Enabled)

	// 無効化されている場合は、APIトークンのバリデーションをスキップ
	var apiToken string
	if enabled {
		var err error
		apiToken, err = resolveSecret(c.APIToken, c.APITokenEnv, "output.slack_api.api_token_env")
		if err != nil {
			return nil, err
		}
	}

	// MessageTemplateの別名変換処理
	converter := entity.NewSlackTemplateAliasConverter()
	convertedTemplate, err := convertMessageTemplate(c.MessageTemplate, converter)
	if err != nil {
		return nil, err
	}

	return &entity.SlackAPIConfig{
		Enabled:         enabled,
		APIToken:        apiToken,
		Channel:         c.Channel,
		MessageTemplate: convertedTemplate,
	}, nil
}

type MisskeyConfig struct {
	Enabled         *bool   `yaml:"enabled,omitempty"`
	APIToken        string  `yaml:"api_token"`
	APITokenEnv     string  `yaml:"api_token_env,omitempty"`
	APIURL          string  `yaml:"api_url"`
	MessageTemplate *string `yaml:"message_template,omitempty"`
}

func (c *MisskeyConfig) Merge(other *MisskeyConfig) {
	if other == nil {
		return
	}

	// Enabledフィールドのマージ（*boolポインタのマージ）
	if other.Enabled != nil {
		c.Enabled = other.Enabled
	}

	// プロファイルファイルでapi_token_envが指定されている場合、
	// デフォルトのapi_tokenを無効にして環境変数を優先する
	if other.APITokenEnv != "" && other.APIToken == "" {
		c.APIToken = ""
		c.APITokenEnv = other.APITokenEnv
	} else {
		// 通常のマージ処理
		mergeString(&c.APIToken, other.APIToken)
		mergeString(&c.APITokenEnv, other.APITokenEnv)
	}

	mergeString(&c.APIURL, other.APIURL)
	if other.MessageTemplate != nil {
		c.MessageTemplate = other.MessageTemplate
	}
}

func (c *MisskeyConfig) ToEntity() (*entity.MisskeyConfig, error) {
	// Enabledフィールドの後方互換性処理（省略時=true）
	enabled := resolveEnabled(c.Enabled)

	// 無効化されている場合は、APIトークンのバリデーションをスキップ
	var apiToken string
	if enabled {
		var err error
		apiToken, err = resolveSecret(c.APIToken, c.APITokenEnv, "output.misskey.api_token_env")
		if err != nil {
			return nil, err
		}
	}

	// MessageTemplateの別名変換処理
	converter := entity.NewMisskeyTemplateAliasConverter()
	convertedTemplate, err := convertMessageTemplate(c.MessageTemplate, converter)
	if err != nil {
		return nil, err
	}

	return &entity.MisskeyConfig{
		Enabled:         enabled,
		APIToken:        apiToken,
		APIURL:          c.APIURL,
		MessageTemplate: convertedTemplate,
	}, nil
}

type ConfigRepository interface {
	Save(config *Config) error
	Load() (*Config, error)
}

type YamlConfigRepository struct {
	filePath string
}

func NewYamlConfigRepository(filePath string) *YamlConfigRepository {
	return &YamlConfigRepository{
		filePath: filePath,
	}
}

// SaveWithTemplate は、テンプレートを使用してコメント付きconfig.ymlファイルを生成する
func (r *YamlConfigRepository) SaveWithTemplate() error {
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("config file already exists: %s", r.filePath)
		}
		return fmt.Errorf("failed to create config file: %s, %w", r.filePath, err)
	}
	defer file.Close()

	// 埋め込まれたYAMLファイルの内容を取得してファイルに書き込み
	templateData, err := GetConfigTemplate()
	if err != nil {
		return fmt.Errorf("failed to get config template: %w", err)
	}

	_, err = file.Write(templateData)
	if err != nil {
		return fmt.Errorf("failed to write config template: %w", err)
	}

	return nil
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

	// 構造体をYAMLエンコードして書き込み（既存のテスト互換性のため）
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (r *YamlConfigRepository) Load() (*Config, error) {
	return LoadYAML[Config](r.filePath)
}

func mergeString(target *string, source string) {
	if source != "" {
		*target = source
	}
}

type merger[T any] interface {
	Merge(T)
}

func mergePtr[T any, P interface {
	*T
	merger[P]
}](target *P, source P) {
	if source != nil {
		if *target == nil {
			*target = new(T)
		}
		(*target).Merge(source)
	}
}
