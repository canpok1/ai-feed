package infra

import (
	"fmt"
	"io"
	"time"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// StdViewer は標準出力にデータを表示するViewer実装
type StdViewer struct {
	loc    *time.Location
	writer io.Writer
}

// NewStdViewer は新しいStdViewerを作成する
func NewStdViewer(writer io.Writer) (domain.Viewer, error) {
	if writer == nil {
		return nil, fmt.Errorf("writer cannot be nil")
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}

	return &StdViewer{
		loc:    loc,
		writer: writer,
	}, nil
}

// ViewArticles は記事のリストを表示する
func (v *StdViewer) ViewArticles(articles []entity.Article) error {
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

// ViewRecommend は推薦記事を表示する
func (v *StdViewer) ViewRecommend(recommend *entity.Recommend, fixedMessage string) error {
	if recommend == nil {
		fmt.Fprintln(v.writer, "No articles found in the feed.")
		return nil
	}

	fmt.Fprintf(v.writer, "Title: %s\n", recommend.Article.Title)
	fmt.Fprintf(v.writer, "Link: %s\n", recommend.Article.Link)
	if recommend.Comment != nil {
		fmt.Fprintf(v.writer, "Comment: %s\n", *recommend.Comment)
	}
	// fixedMessage を追加
	if fixedMessage != "" {
		fmt.Fprintf(v.writer, "Fixed Message: %s\n", fixedMessage)
	}
	return nil
}
