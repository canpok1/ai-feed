package infra

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/slack-go/slack"
)

const DefaultSlackMessageTemplate = `{{if .Comment}}{{.Comment}}
{{end}}{{.Article.Title}}
{{.Article.Link}}{{if .FixedMessage}}
{{.FixedMessage}}{{end}}`

type SlackTemplateData struct {
	Article      *entity.Article
	Comment      *string
	FixedMessage string
}

type SlackViewer struct {
	client    *slack.Client
	channelID string
	tmpl      *template.Template
}

func NewSlackViewer(config *entity.SlackAPIConfig) domain.Viewer {
	// メッセージテンプレートの設定
	messageTemplate := DefaultSlackMessageTemplate
	if config.MessageTemplate != nil && strings.TrimSpace(*config.MessageTemplate) != "" {
		messageTemplate = *config.MessageTemplate
	}

	// 設定読み込み時にテンプレートは検証済みのため、template.Mustが安全に使用できる
	tmpl := template.Must(template.New("slack_message").Parse(messageTemplate))

	return &SlackViewer{
		client:    slack.New(config.APIToken),
		channelID: config.Channel,
		tmpl:      tmpl,
	}
}

func (s *SlackViewer) ViewArticles(articles []entity.Article) error {
	// TODO 実装
	return nil
}

func (v *SlackViewer) ViewRecommend(recommend *entity.Recommend, fixedMessage string) error {
	// テンプレートデータを作成
	templateData := &SlackTemplateData{
		Article:      &recommend.Article,
		Comment:      recommend.Comment,
		FixedMessage: fixedMessage,
	}

	// パース済みテンプレートを直接実行
	var buf bytes.Buffer
	if err := v.tmpl.Execute(&buf, templateData); err != nil {
		return err
	}

	return v.postMessage(buf.String())
}

func (v *SlackViewer) postMessage(msg string) error {
	_, _, err := v.client.PostMessage(
		v.channelID,
		slack.MsgOptionText(msg, false),
	)
	return err
}
