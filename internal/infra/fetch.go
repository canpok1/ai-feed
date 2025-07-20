package infra

import (
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/mmcdole/gofeed"
)

type FetchClient struct{}

func NewFetchClient() domain.FetchClient {
	return &FetchClient{}
}

func (f *FetchClient) Fetch(url string) ([]entity.Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	var articles []entity.Article
	for _, item := range feed.Items {
		content := ""
		if item.Content != "" {
			content = item.Content
		} else if item.Description != "" {
			content = item.Description
		}

		articles = append(articles, entity.Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: item.PublishedParsed,
			Content:   content,
		})
	}
	return articles, nil
}
