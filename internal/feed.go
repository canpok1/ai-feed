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

// FetchFeedFunc defines the signature for the FetchFeed function.
type FetchFeedFunc func(url string) ([]Article, error)

// defaultFetchFeed is the default implementation of FetchFeed.
var defaultFetchFeed FetchFeedFunc = func(url string) ([]Article, error) {
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

// FetchFeed is the function that should be called to fetch feeds. It can be overridden for testing.
var FetchFeed = defaultFetchFeed
