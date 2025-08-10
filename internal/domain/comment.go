package domain

import (
	"context"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// CommentGenerator はAIを使用してコメントを生成するインターフェース
type CommentGenerator interface {
	Generate(context.Context, *entity.Article) (string, error)
}

// CommentGeneratorFactory はCommentGeneratorを生成するファクトリのインターフェース
type CommentGeneratorFactory interface {
	MakeCommentGenerator(*entity.AIConfig, *entity.PromptConfig) (CommentGenerator, error)
}
