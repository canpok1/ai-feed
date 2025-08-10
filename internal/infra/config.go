package infra

import (
	"fmt"
	"os"
	"strings"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

// indentLines は文字列の各行に指定されたインデントを追加する
func indentLines(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// configYmlTemplate は、config initコマンドで生成するconfig.ymlのテンプレート文字列
var configYmlTemplate = `# AI Feedの設定ファイル
# このファイルには全プロファイル共通のデフォルト設定を定義します
default_profile:
` + indentLines(ProfileTemplateCore, "  ")

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
func (p *Profile) ToEntity() *entity.Profile {
	var aiEntity *entity.AIConfig
	if p.AI != nil {
		aiEntity = p.AI.ToEntity()
	}

	var promptEntity *entity.PromptConfig
	if p.Prompt != nil {
		promptEntity = p.Prompt.ToEntity()
	}

	var outputEntity *entity.OutputConfig
	if p.Output != nil {
		outputEntity = p.Output.ToEntity()
	}

	return &entity.Profile{
		AI:     aiEntity,
		Prompt: promptEntity,
		Output: outputEntity,
	}
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

func (c *AIConfig) ToEntity() *entity.AIConfig {
	var geminiEntity *entity.GeminiConfig
	if c.Gemini != nil {
		geminiEntity = c.Gemini.ToEntity()
	}
	return &entity.AIConfig{
		Gemini: geminiEntity,
	}
}

type GeminiConfig struct {
	Type   string `yaml:"type"`
	APIKey string `yaml:"api_key"`
}

func (c *GeminiConfig) Merge(other *GeminiConfig) {
	if other == nil {
		return
	}
	mergeString(&c.Type, other.Type)
	mergeString(&c.APIKey, other.APIKey)
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

func (c *OutputConfig) ToEntity() *entity.OutputConfig {
	var slackEntity *entity.SlackAPIConfig
	if c.SlackAPI != nil {
		slackEntity = c.SlackAPI.ToEntity()
	}

	var misskeyEntity *entity.MisskeyConfig
	if c.Misskey != nil {
		misskeyEntity = c.Misskey.ToEntity()
	}

	return &entity.OutputConfig{
		SlackAPI: slackEntity,
		Misskey:  misskeyEntity,
	}
}

type SlackAPIConfig struct {
	APIToken        string  `yaml:"api_token"`
	Channel         string  `yaml:"channel"`
	MessageTemplate *string `yaml:"message_template,omitempty"`
}

func (c *SlackAPIConfig) Merge(other *SlackAPIConfig) {
	if other == nil {
		return
	}
	mergeString(&c.APIToken, other.APIToken)
	mergeString(&c.Channel, other.Channel)
	if other.MessageTemplate != nil {
		c.MessageTemplate = other.MessageTemplate
	}
}

func (c *SlackAPIConfig) ToEntity() *entity.SlackAPIConfig {
	return &entity.SlackAPIConfig{
		APIToken:        c.APIToken,
		Channel:         c.Channel,
		MessageTemplate: c.MessageTemplate,
	}
}

type MisskeyConfig struct {
	APIToken string `yaml:"api_token"`
	APIURL   string `yaml:"api_url"`
}

func (c *MisskeyConfig) Merge(other *MisskeyConfig) {
	if other == nil {
		return
	}
	mergeString(&c.APIToken, other.APIToken)
	mergeString(&c.APIURL, other.APIURL)
}

func (c *MisskeyConfig) ToEntity() *entity.MisskeyConfig {
	return &entity.MisskeyConfig{
		APIToken: c.APIToken,
		APIURL:   c.APIURL,
	}
}

func MakeDefaultConfig() *Config {
	return &Config{
		DefaultProfile: &Profile{
			AI: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "gemini-1.5-flash",
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
				FixedMessage: "固定の文言です。",
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
