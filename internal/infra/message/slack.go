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

type slackClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

type SlackSender struct {
	client slackClient
	config *entity.SlackAPIConfig
	tmpl   *template.Template
}

func NewSlackSender(config *entity.SlackAPIConfig, client slackClient) domain.MessageSender {
	// 設定読み込み時にテンプレートは検証済みのため、template.Mustが安全に使用できる
	// ただし、テストやバリデーション前の呼び出しに対応するため念のためnilチェックを行う
	if config.MessageTemplate == nil || *config.MessageTemplate == "" {
		panic("MessageTemplate is required and must be validated before creating SlackSender")
	}
	tmpl := template.Must(template.New("slack_message").Parse(*config.MessageTemplate))

	return &SlackSender{
		client: client,
		config: config,
		tmpl:   tmpl,
	}
}

func (s *SlackSender) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	// テンプレートデータを作成
	templateData := &SlackTemplateData{
		Article:      &recommend.Article,
		Comment:      recommend.Comment,
		FixedMessage: fixedMessage,
	}

	// パース済みテンプレートを直接実行
	var buf bytes.Buffer
	if err := s.tmpl.Execute(&buf, templateData); err != nil {
		return err
	}

	return s.postMessage(buf.String())
}

func (s *SlackSender) postMessage(msg string) error {
	options := []slack.MsgOption{
		slack.MsgOptionText(msg, false),
	}
	if s.config.Username != nil && *s.config.Username != "" {
		options = append(options, slack.MsgOptionUsername(*s.config.Username))
	}
	if s.config.IconURL != nil && *s.config.IconURL != "" {
		options = append(options, slack.MsgOptionIconURL(*s.config.IconURL))
	}
	if s.config.IconEmoji != nil && *s.config.IconEmoji != "" {
		options = append(options, slack.MsgOptionIconEmoji(*s.config.IconEmoji))
	}

	_, _, err := s.client.PostMessage(
		s.config.Channel,
		options...,
	)
	return err
}

// ServiceName はサービス名を返す
func (s *SlackSender) ServiceName() string {
	return "Slack"
}
