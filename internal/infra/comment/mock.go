package comment

import (
	"context"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// mockCommentGenerator はテスト用のモックコメント生成器
type mockCommentGenerator struct {
	comment string // 固定で返すコメント
}

// newMockCommentGenerator は新しいモックコメント生成器を作成する
func newMockCommentGenerator(comment string) *mockCommentGenerator {
	return &mockCommentGenerator{comment: comment}
}

// Generate は設定された固定コメントを返す
func (g *mockCommentGenerator) Generate(_ context.Context, _ *entity.Article) (string, error) {
	return g.comment, nil
}
