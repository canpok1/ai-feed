package internal

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type Article struct {
	Title     string
	Link      string
	Published *time.Time
	Content   string
}

func FetchFeed(url string) ([]Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	var articles []Article
	for _, item := range feed.Items {
		content := ""
		if item.Content != "" {
			content = item.Content
		} else if item.Description != "" {
			content = item.Description
		}

		articles = append(articles, Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: item.PublishedParsed,
			Content:   content,
		})
	}
	return articles, nil
}
