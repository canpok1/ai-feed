//go:build e2e

// Package mock はe2eテスト用のモックサーバーを提供する
package mock

import (
	"net/http"
)

// NewMockRSSHandler は標準的なRSSフィードを返すモックハンドラを生成する
func NewMockRSSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		// レスポンスの書き込みエラーは通常発生しないが、
		// クライアントが接続を切断した場合などに備えてエラーを無視
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS Feed</title>
    <link>https://example.com</link>
    <description>Test RSS Feed for E2E Testing</description>
    <item>
      <title>Test Article 1</title>
      <link>https://example.com/article1</link>
      <description>This is test article 1</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Test Article 2</title>
      <link>https://example.com/article2</link>
      <description>This is test article 2</description>
      <pubDate>Tue, 02 Jan 2024 00:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Test Article 3</title>
      <link>https://example.com/article3</link>
      <description>This is test article 3</description>
      <pubDate>Wed, 03 Jan 2024 00:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`))
	})
}

// NewMockAtomHandler はAtomフィードを返すモックハンドラを生成する
func NewMockAtomHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		// レスポンスの書き込みエラーは通常発生しないが、
		// クライアントが接続を切断した場合などに備えてエラーを無視
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Atom Feed</title>
  <link href="https://example.com"/>
  <updated>2024-01-01T00:00:00Z</updated>
  <id>https://example.com/feed</id>
  <entry>
    <title>Test Atom Article 1</title>
    <link href="https://example.com/atom1"/>
    <id>https://example.com/atom1</id>
    <updated>2024-01-01T00:00:00Z</updated>
    <summary>This is test atom article 1</summary>
  </entry>
  <entry>
    <title>Test Atom Article 2</title>
    <link href="https://example.com/atom2"/>
    <id>https://example.com/atom2</id>
    <updated>2024-01-02T00:00:00Z</updated>
    <summary>This is test atom article 2</summary>
  </entry>
</feed>`))
	})
}

// NewMockEmptyFeedHandler は記事が0件の空フィードを返すモックハンドラを生成する
func NewMockEmptyFeedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		// レスポンスの書き込みエラーは通常発生しないが、
		// クライアントが接続を切断した場合などに備えてエラーを無視
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>Empty Feed for Testing</description>
  </channel>
</rss>`))
	})
}

// NewMockInvalidFeedHandler は不正な形式のXMLを返すモックハンドラを生成する
func NewMockInvalidFeedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		// レスポンスの書き込みエラーは通常発生しないが、
		// クライアントが接続を切断した場合などに備えてエラーを無視
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Invalid Feed</title>
    <link>https://example.com</link>
    <description>Invalid Feed for Testing</description>
    <item>
      <title>Unclosed Item
      <link>https://example.com/broken
    </item>
  </channel>`))
	})
}

// NewMockErrorHandler は指定されたHTTPステータスコードを返すモックハンドラを生成する
func NewMockErrorHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	})
}
