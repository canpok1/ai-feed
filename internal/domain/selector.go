package domain

import (
	"context"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ArticleSelector は複数の記事から1つを選択するインターフェース
type ArticleSelector interface {
	Select(ctx context.Context, articles []entity.Article) (*entity.Article, error)
}
