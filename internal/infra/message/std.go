package message

import (
	"fmt"
	"io"
	"time"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// recommendTemplate は推薦記事の出力テンプレート
const recommendTemplate = `Title: {{ .Article.Title }}
Link: {{ .Article.Link }}
{{ if .Comment }}Comment: {{ .Comment }}
{{ end }}{{ if .FixedMessage }}Fixed Message: {{ .FixedMessage }}
{{ end }}`

// StdSender は標準出力にデータを送信するSender実装
type StdSender struct {
	loc            *time.Location
	writer         io.Writer
	messageBuilder *MessageBuilder
}

// NewStdSender は新しいStdSenderを作成する
func NewStdSender(writer io.Writer) (domain.MessageSender, error) {
	if writer == nil {
		return nil, fmt.Errorf("writer cannot be nil")
	}

	loc := time.Local

	messageBuilder, err := NewMessageBuilder(recommendTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create message builder: %w", err)
	}

	return &StdSender{
		loc:            loc,
		writer:         writer,
		messageBuilder: messageBuilder,
	}, nil
}

// SendRecommend は推薦記事を表示する
func (v *StdSender) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	if recommend == nil {
		fmt.Fprintln(v.writer, "No articles found in the feed.")
		return nil
	}

	message, err := v.messageBuilder.BuildRecommendMessage(recommend, fixedMessage)
	if err != nil {
		return fmt.Errorf("failed to build recommend message: %w", err)
	}

	fmt.Fprint(v.writer, message)
	return nil
}
