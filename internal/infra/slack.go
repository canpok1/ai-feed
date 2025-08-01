package infra

import (
	"strings"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/slack-go/slack"
)

type SlackViewer struct {
	client    *slack.Client
	channelID string
}

func NewSlackViewer(config *entity.SlackAPIConfig) domain.Viewer {
	return &SlackViewer{
		client:    slack.New(config.APIToken),
		channelID: config.Channel,
	}
}

func (s *SlackViewer) ViewArticles(articles []entity.Article) error {
	// TODO 実装
	return nil
}

func (v *SlackViewer) ViewRecommend(recommend *entity.Recommend, fixedMessage string) error {
	var messages []string
	if recommend.Comment != nil && *recommend.Comment != "" {
		messages = make([]string, 0, 3)
		messages = append(messages, *recommend.Comment)
	} else {
		messages = make([]string, 0, 2)
	}
	messages = append(messages, recommend.Article.Title)
	messages = append(messages, recommend.Article.Link)

	msg := strings.Join(messages, "\n")

	// 引数で渡されたfixedMessageを使用
	if fixedMessage != "" {
		msg = msg + "\n" + fixedMessage
	}

	return v.postMessage(msg)
}

func (v *SlackViewer) postMessage(msg string) error {
	_, _, err := v.client.PostMessage(
		v.channelID,
		slack.MsgOptionText(msg, false),
	)
	return err
}
