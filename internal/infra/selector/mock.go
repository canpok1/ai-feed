package selector

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// mockArticleSelector はテスト用のモック記事選択器
type mockArticleSelector struct {
	mode string // "first", "random", "last"
}

// newMockArticleSelector は新しいモック記事選択器を作成する
func newMockArticleSelector(mode string) (*mockArticleSelector, error) {
	if !entity.IsValidMockSelectorMode(mode) {
		return nil, fmt.Errorf("invalid selector mode: %s (must be first, random, or last)", mode)
	}
	return &mockArticleSelector{mode: mode}, nil
}

// Select は設定されたモードに基づいて記事を選択する
func (s *mockArticleSelector) Select(_ context.Context, articles []entity.Article) (*entity.Article, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles to select from")
	}

	var index int
	switch s.mode {
	case "first":
		index = 0
	case "last":
		index = len(articles) - 1
	case "random":
		index = rand.Intn(len(articles))
	default:
		index = 0
	}

	return &articles[index], nil
}
