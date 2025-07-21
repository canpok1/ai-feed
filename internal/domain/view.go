package domain

import (
	"fmt"
	"io"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Viewer interface {
	ViewArticles([]entity.Article) error
	ViewRecommend(*entity.Recommend) error
}

type StdViewer struct {
	loc    *time.Location
	writer io.Writer
}

func NewStdViewer(writer io.Writer) (Viewer, error) {
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

func (v *StdViewer) ViewRecommend(recommend *entity.Recommend) error {
	if recommend == nil {
		fmt.Fprintln(v.writer, "No articles found in the feed.")
		return nil
	}

	fmt.Fprintf(v.writer, "Title: %s\n", recommend.Article.Title)
	fmt.Fprintf(v.writer, "Link: %s\n", recommend.Article.Link)
	if recommend.Comment != nil {
		fmt.Fprintf(v.writer, "Comment: %s\n", *recommend.Comment)
	}
	return nil
}
