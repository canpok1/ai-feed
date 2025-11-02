package cache

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewNopCache(t *testing.T) {
	t.Run("コンストラクタが正しくインスタンスを返す", func(t *testing.T) {
		cache := NewNopCache()

		assert.NotNil(t, cache)
		assert.IsType(t, &NopCache{}, cache)
	})
}

func TestNopCache_Initialize(t *testing.T) {
	t.Run("常にnilを返す", func(t *testing.T) {
		cache := NewNopCache()

		err := cache.Initialize()

		assert.NoError(t, err)
		assert.Nil(t, err)
	})
}

func TestNopCache_IsCached(t *testing.T) {
	t.Run("任意のURLに対して常にfalseを返す", func(t *testing.T) {
		cache := NewNopCache()

		testCases := []struct {
			name string
			url  string
		}{
			{name: "通常のURL", url: "https://example.com/article"},
			{name: "空文字列", url: ""},
			{name: "不正なURL", url: "invalid-url"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := cache.IsCached(tc.url)
				assert.False(t, result, "IsCachedは常にfalseを返すべき")
			})
		}
	})
}

func TestNopCache_AddEntry(t *testing.T) {
	t.Run("任意のURL_タイトルに対して常にnilを返す", func(t *testing.T) {
		cache := NewNopCache()

		testCases := []struct {
			name  string
			url   string
			title string
		}{
			{name: "通常のエントリ", url: "https://example.com/article", title: "Test Article"},
			{name: "空文字列", url: "", title: ""},
			{name: "長いタイトル", url: "https://example.com/test", title: "これは非常に長いタイトルのテストです"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := cache.AddEntry(tc.url, tc.title)
				assert.NoError(t, err)
				assert.Nil(t, err)
			})
		}
	})
}

func TestNopCache_Close(t *testing.T) {
	t.Run("常にnilを返す", func(t *testing.T) {
		cache := NewNopCache()

		err := cache.Close()

		assert.NoError(t, err)
		assert.Nil(t, err)
	})
}

func TestNopCache_ImplementsInterface(t *testing.T) {
	t.Run("RecommendCacheインターフェースを実装している", func(t *testing.T) {
		var _ domain.RecommendCache = (*NopCache)(nil)
		// コンパイルエラーが発生しなければテストは成功
	})
}
