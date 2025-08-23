package domain

import (
	"log/slog"
	"sort"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type FetchClient interface {
	Fetch(url string) ([]entity.Article, error)
}

type Fetcher struct {
	client        FetchClient
	errorCallback ErrorCallback
}

type ErrorCallback func(string, error) error

func NewFetcher(client FetchClient, errorCallback ErrorCallback) *Fetcher {
	return &Fetcher{
		client:        client,
		errorCallback: errorCallback,
	}
}

func (f *Fetcher) Fetch(urls []string, limit int) ([]entity.Article, error) {
	var allArticles []entity.Article
	for _, url := range urls {
		articles, err := f.client.Fetch(url)
		if err != nil {
			if err := f.errorCallback(url, err); err != nil {
				return nil, err
			}
			continue
		}

		slog.Debug("記事を取得しました", "feed_url", url, "article_count", len(articles))
		allArticles = append(allArticles, articles...)
	}

	sort.Slice(allArticles, func(i, j int) bool {
		if allArticles[i].Published == nil {
			return false
		}
		if allArticles[j].Published == nil {
			return true
		}
		return allArticles[i].Published.After(*allArticles[j].Published)
	})

	if limit > 0 && len(allArticles) > limit {
		allArticles = allArticles[:limit]
	}

	return allArticles, nil
}
