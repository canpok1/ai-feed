package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFetchClient は、テスト用のFetchClientの実装
type mockFetchClient struct {
	responses map[string][]entity.Article
	errors    map[string]error
}

func newMockFetchClient() *mockFetchClient {
	return &mockFetchClient{
		responses: make(map[string][]entity.Article),
		errors:    make(map[string]error),
	}
}

func (m *mockFetchClient) setResponse(url string, articles []entity.Article) {
	m.responses[url] = articles
}

func (m *mockFetchClient) setError(url string, err error) {
	m.errors[url] = err
}

func (m *mockFetchClient) Fetch(url string) ([]entity.Article, error) {
	if err, exists := m.errors[url]; exists {
		return nil, err
	}
	if articles, exists := m.responses[url]; exists {
		return articles, nil
	}
	return []entity.Article{}, nil
}

func TestFetcherLogOutput(t *testing.T) {
	t.Run("記事取得成功時のデバッグログ出力", func(t *testing.T) {
		// モッククライアントの準備
		mockClient := newMockFetchClient()

		// 記事データの準備
		now := time.Now()
		mockClient.setResponse("https://example1.com/feed.xml", []entity.Article{
			{Title: "記事1", Link: "https://example1.com/article1", Published: &now},
		})
		mockClient.setResponse("https://example2.com/feed.xml", []entity.Article{
			{Title: "記事2", Link: "https://example2.com/article2", Published: &now},
		})

		// エラーコールバックの準備
		var errorCallbacks []string
		errorCallback := func(url string, err error) error {
			errorCallbacks = append(errorCallbacks, url+": "+err.Error())
			return err
		}

		// Fetcherの作成
		fetcher := NewFetcher(mockClient, errorCallback)

		// テスト実行
		urls := []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml"}
		articles, err := fetcher.Fetch(urls, 0)

		// 結果の検証
		assert.NoError(t, err)
		assert.Len(t, articles, 2)
		assert.Empty(t, errorCallbacks) // エラーが発生していないことを確認
		assert.Equal(t, "記事1", articles[0].Title)
		assert.Equal(t, "記事2", articles[1].Title)
	})

	t.Run("記事取得失敗時のエラーコールバック呼び出し", func(t *testing.T) {
		// モッククライアントの準備
		mockClient := newMockFetchClient()

		// 記事データとエラーの準備
		now := time.Now()
		mockClient.setResponse("https://example1.com/feed.xml", []entity.Article{
			{Title: "記事1", Link: "https://example1.com/article1", Published: &now},
		})
		mockClient.setError("https://example2.com/feed.xml", errors.New("fetch error"))

		// エラーコールバックの準備
		var errorCallbacks []string
		errorCallback := func(url string, err error) error {
			errorCallbacks = append(errorCallbacks, url+": "+err.Error())
			return nil // エラーを無視してcontinueさせる
		}

		// Fetcherの作成
		fetcher := NewFetcher(mockClient, errorCallback)

		// テスト実行
		urls := []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml"}
		articles, err := fetcher.Fetch(urls, 0)

		// 結果の検証
		assert.NoError(t, err)
		assert.Len(t, articles, 1)       // 成功した記事のみ
		assert.Len(t, errorCallbacks, 1) // エラーコールバックが1回呼ばれている
		assert.True(t, strings.Contains(errorCallbacks[0], "https://example2.com/feed.xml: fetch error"))
		assert.Equal(t, "記事1", articles[0].Title)
	})
}

func TestFetcherMultipleURLs(t *testing.T) {
	t.Run("複数URLからの記事取得", func(t *testing.T) {
		// モッククライアントの準備
		mockClient := newMockFetchClient()

		// 記事データの準備（異なる公開日時で順序を確認）
		time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)
		time3 := time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC)

		mockClient.setResponse("https://example1.com/feed.xml", []entity.Article{
			{Title: "古い記事", Link: "https://example1.com/old", Published: &time1},
			{Title: "新しい記事", Link: "https://example1.com/new", Published: &time3},
		})
		mockClient.setResponse("https://example2.com/feed.xml", []entity.Article{
			{Title: "中間の記事", Link: "https://example2.com/middle", Published: &time2},
		})

		// エラーコールバックの準備
		errorCallback := func(url string, err error) error {
			return err
		}

		// Fetcherの作成
		fetcher := NewFetcher(mockClient, errorCallback)

		// テスト実行
		urls := []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml"}
		articles, err := fetcher.Fetch(urls, 0)

		// 結果の検証
		require.NoError(t, err)
		require.Len(t, articles, 3)

		// 記事が公開日時順（新しい順）でソートされていることを確認
		assert.Equal(t, "新しい記事", articles[0].Title)
		assert.Equal(t, "中間の記事", articles[1].Title)
		assert.Equal(t, "古い記事", articles[2].Title)
	})
}
