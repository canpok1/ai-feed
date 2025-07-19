package domain

import "sort"

type FetchClient interface {
	Fetch(url string) ([]Article, error)
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

func (f *Fetcher) Fetch(urls []string, limit int) ([]Article, error) {
	var allArticles []Article
	for _, url := range urls {
		articles, err := f.client.Fetch(url)
		if err != nil {
			if err := f.errorCallback(url, err); err != nil {
				return nil, err
			}
			continue
		}

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
