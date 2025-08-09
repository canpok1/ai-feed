package infra

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

// StdViewer は標準出力にデータを表示するViewer実装
type StdViewer struct {
	loc            *time.Location
	writer         io.Writer
	messageBuilder *MessageBuilder
}

// NewStdViewer は新しいStdViewerを作成する
func NewStdViewer(writer io.Writer) (domain.MessageSender, error) {
	if writer == nil {
		return nil, fmt.Errorf("writer cannot be nil")
	}

	loc := time.Local

	messageBuilder, err := NewMessageBuilder(recommendTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create message builder: %w", err)
	}

	return &StdViewer{
		loc:            loc,
		writer:         writer,
		messageBuilder: messageBuilder,
	}, nil
}

// SendArticles は記事のリストを表示する
func (v *StdViewer) SendArticles(articles []entity.Article) error {
	for _, article := range articles {
		fmt.Fprintf(v.writer, "Title: %s\n", article.Title)
		fmt.Fprintf(v.writer, "Link: %s\n", article.Link)
		if article.Published != nil {
			fmt.Fprintf(v.writer, "Published: %s\n", article.Published.In(v.loc).Format("2006-01-02 15:04:05 JST"))
		}
		fmt.Fprintf(v.writer, "Content: %s\n", article.Content)
		fmt.Fprintln(v.writer, "---")
	}
	return nil
}

// SendRecommend は推薦記事を表示する
func (v *StdViewer) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
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
