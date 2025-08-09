package infra

import (
	"fmt"
	"os"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

// profileYmlTemplate は、profile initコマンドで生成するprofile.ymlのテンプレート文字列
const profileYmlTemplate = `# AI Feedのプロファイル設定ファイル
# このファイルの設定は config.yml のデフォルトプロファイル設定を上書きします
# AI設定
ai:
  gemini:
    type: gemini-2.5-flash                  # 使用するGeminiモデル
    api_key: xxxxxx                         # Google AI Studio APIキー

# プロンプト設定
system_prompt: あなたはXXXXなAIアシスタントです。    # AIに与えるシステムプロンプト
comment_prompt_template: |                         # 記事紹介文生成用のプロンプトテンプレート
  以下の記事の紹介文を100字以内で作成してください。
  ---
  記事タイトル: {{"{{title}}"}}
  記事URL: {{"{{url}}"}}
  記事内容:
  {{"{{content}}"}}
fixed_message: 固定の文言です。                     # 記事紹介文に追加する固定文言

# 出力先設定
output:
  # Slack投稿設定
  slack_api:
    api_token: xxxxxx                       # Slack Bot Token
    api_url: https://example.com            # Slack API URL
    channel: "#general"                     # 投稿先チャンネル
  
  # Misskey投稿設定
  misskey:
    api_token: xxxxxx                       # Misskeyアクセストークン
    api_url: https://misskey.social/api     # MisskeyのAPIエンドポイント
`

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

// SaveProfileWithTemplate は、テンプレートを使用してコメント付きprofile.ymlファイルを生成する
func (r *YamlProfileRepository) SaveProfileWithTemplate() error {
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("profile file already exists: %s", r.filePath)
		}
		return fmt.Errorf("failed to create profile file: %s, %w", r.filePath, err)
	}
	defer file.Close()

	// テンプレートを実行してファイルに書き込み
	tmpl, err := template.New("profile").Parse(profileYmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse profile template: %w", err)
	}

	err = tmpl.Execute(file, nil)
	if err != nil {
		return fmt.Errorf("failed to execute profile template: %w", err)
	}

	return nil
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
