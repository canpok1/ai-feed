package domain

import (
	"fmt"
	"io"
	"time"
)

type Viewer interface {
	ViewArticles(io.Writer, []Article) error
	ViewRecommend(io.Writer, *Recommend) error
}

type StdViewer struct{}

func NewStdViewer() Viewer {
	return &StdViewer{}
}

func (v *StdViewer) ViewArticles(w io.Writer, articles []Article) error {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	for _, article := range articles {
		fmt.Fprintf(w, "Title: %s\n", article.Title)
		fmt.Fprintf(w, "Link: %s\n", article.Link)
		if article.Published != nil {
			fmt.Fprintf(w, "Published: %s\n", article.Published.In(loc).Format("2006-01-02 15:04:05 JST"))
		}
		fmt.Fprintf(w, "Content: %s\n", article.Content)
		fmt.Fprintln(w, "---")
	}
	return nil
}

func (v *StdViewer) ViewRecommend(w io.Writer, recommend *Recommend) error {
	if recommend == nil {
		fmt.Fprintln(w, "No articles found in the feed.")
		return nil
	}

	fmt.Fprintf(w, "Title: %s\n", recommend.Article.Title)
	fmt.Fprintf(w, "Link: %s\n", recommend.Article.Link)
	return nil
}
