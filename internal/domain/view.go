package domain

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Viewer interface {
	ViewArticles([]entity.Article) error
	ViewRecommend(*entity.Recommend, string) error
}
