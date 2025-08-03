package infra

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

type YamlProfileRepository struct {
	filePath string
}

func NewYamlProfileRepository(filePath string) *YamlProfileRepository {
	return &YamlProfileRepository{
		filePath: filePath,
	}
}

func (r *YamlProfileRepository) LoadProfile() (*Profile, error) {
	return loadYaml[Profile](r.filePath)
}

func (r *YamlProfileRepository) SaveProfile(profile *Profile) error {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile to YAML: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile to file %q: %w", r.filePath, err)
	}
	return nil
}

func NewDefaultProfile() *Profile {
	return &Profile{
		AI: &AIConfig{
			Gemini: &GeminiConfig{
				Type:   "gemini-1.5-flash",
				APIKey: entity.DefaultGeminiAPIKey,
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
				APIToken: entity.DefaultSlackAPIToken,
				Channel:  "#general",
			},
			Misskey: &MisskeyConfig{
				APIToken: entity.DefaultMisskeyAPIToken,
				APIURL:   "https://misskey.social/api",
			},
		},
	}
}
