package profile

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
)

// profileYmlTemplate は、profile initコマンドで生成するprofile.ymlのテンプレート文字列
var profileYmlTemplate = `# AI Feedのプロファイル設定ファイル
# このファイルの設定は config.yml のデフォルトプロファイル設定を上書きします
` + infra.ProfileTemplateCore

// NewDefaultProfile はデフォルトのプロファイルを生成する
func NewDefaultProfile() *infra.Profile {
	return &infra.Profile{
		AI: &infra.AIConfig{
			Gemini: &infra.GeminiConfig{
				Type:   "gemini-1.5-flash",
				APIKey: entity.DefaultGeminiAPIKey,
			},
		},
		Prompt: &infra.PromptConfig{
			SystemPrompt: "あなたはXXXXなAIアシスタントです。",
			CommentPromptTemplate: `以下の記事の紹介文を100字以内で作成してください。
---
記事タイトル: {{title}}
記事URL: {{url}}
記事内容:
{{content}}`,
			FixedMessage: "固定の文言です。",
		},
		Output: &infra.OutputConfig{
			SlackAPI: &infra.SlackAPIConfig{
				APIToken: entity.DefaultSlackAPIToken,
				Channel:  "#general",
			},
			Misskey: &infra.MisskeyConfig{
				APIToken: entity.DefaultMisskeyAPIToken,
				APIURL:   "https://misskey.social/api",
			},
		},
	}
}
