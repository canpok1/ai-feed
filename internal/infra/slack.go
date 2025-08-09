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
	client          *slack.Client
	channelID       string
	messageTemplate string
}

func NewSlackViewer(config *entity.SlackAPIConfig) domain.Viewer {
	// メッセージテンプレートの設定
	messageTemplate := DefaultSlackMessageTemplate
	if config.MessageTemplate != nil && strings.TrimSpace(*config.MessageTemplate) != "" {
		messageTemplate = *config.MessageTemplate
	}

	return &SlackViewer{
		client:          slack.New(config.APIToken),
		channelID:       config.Channel,
		messageTemplate: messageTemplate,
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

	// テンプレートをパースして実行
	tmpl, err := template.New("slack_message").Parse(v.messageTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
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
