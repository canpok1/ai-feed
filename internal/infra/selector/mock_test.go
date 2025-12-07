package selector

import (
	"context"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockArticleSelector(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{
			name:        "正常系_firstモード",
			mode:        "first",
			expectError: false,
		},
		{
			name:        "正常系_lastモード",
			mode:        "last",
			expectError: false,
		},
		{
			name:        "正常系_randomモード",
			mode:        "random",
			expectError: false,
		},
		{
			name:        "異常系_不正なモード",
			mode:        "invalid",
			expectError: true,
		},
		{
			name:        "異常系_空文字列",
			mode:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := newMockArticleSelector(tt.mode)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, selector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, selector)
				assert.Equal(t, tt.mode, selector.mode)
			}
		})
	}
}

func TestMockArticleSelector_Select(t *testing.T) {
	now := time.Now()
	articles := []entity.Article{
		{Title: "Article 1", Link: "https://example.com/1", Published: &now, Content: "Content 1"},
		{Title: "Article 2", Link: "https://example.com/2", Published: &now, Content: "Content 2"},
		{Title: "Article 3", Link: "https://example.com/3", Published: &now, Content: "Content 3"},
	}

	t.Run("firstモードは最初の記事を返す", func(t *testing.T) {
		selector, err := newMockArticleSelector("first")
		require.NoError(t, err)

		article, err := selector.Select(context.Background(), articles)
		require.NoError(t, err)
		assert.Equal(t, "Article 1", article.Title)
	})

	t.Run("lastモードは最後の記事を返す", func(t *testing.T) {
		selector, err := newMockArticleSelector("last")
		require.NoError(t, err)

		article, err := selector.Select(context.Background(), articles)
		require.NoError(t, err)
		assert.Equal(t, "Article 3", article.Title)
	})

	t.Run("randomモードは記事を返す", func(t *testing.T) {
		selector, err := newMockArticleSelector("random")
		require.NoError(t, err)

		article, err := selector.Select(context.Background(), articles)
		require.NoError(t, err)
		assert.NotNil(t, article)
		// randomなので、どの記事が返ってきてもOK
		assert.Contains(t, []string{"Article 1", "Article 2", "Article 3"}, article.Title)
	})

	t.Run("空の記事リストはエラーを返す", func(t *testing.T) {
		selector, err := newMockArticleSelector("first")
		require.NoError(t, err)

		article, err := selector.Select(context.Background(), []entity.Article{})
		assert.Error(t, err)
		assert.Nil(t, article)
		assert.Contains(t, err.Error(), "no articles to select from")
	})

	t.Run("単一記事のリスト", func(t *testing.T) {
		singleArticle := []entity.Article{articles[0]}

		selector, err := newMockArticleSelector("last")
		require.NoError(t, err)

		article, err := selector.Select(context.Background(), singleArticle)
		require.NoError(t, err)
		assert.Equal(t, "Article 1", article.Title)
	})
}
