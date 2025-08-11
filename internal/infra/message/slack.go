package message

import (
	"bytes"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/slack-go/slack"
)

type SlackTemplateData struct {
	Article      *entity.Article
	Comment      *string
	FixedMessage string
}

type SlackSender struct {
	client    *slack.Client
	channelID string
	tmpl      *template.Template
}

func NewSlackSender(config *entity.SlackAPIConfig) domain.MessageSender {
	// 設定読み込み時にテンプレートは検証済みのため、template.Mustが安全に使用できる
	// ただし、テストやバリデーション前の呼び出しに対応するため念のためnilチェックを行う
	if config.MessageTemplate == nil || *config.MessageTemplate == "" {
		panic("MessageTemplate is required and must be validated before creating SlackSender")
	}
	tmpl := template.Must(template.New("slack_message").Parse(*config.MessageTemplate))

	return &SlackSender{
		client:    slack.New(config.APIToken),
		channelID: config.Channel,
		tmpl:      tmpl,
	}
}

func (v *SlackSender) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
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

func (v *SlackSender) postMessage(msg string) error {
	_, _, err := v.client.PostMessage(
		v.channelID,
		slack.MsgOptionText(msg, false),
	)
	return err
}
