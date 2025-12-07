package comment

import (
	"context"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewMockCommentGenerator(t *testing.T) {
	t.Run("正常系_コメント付きで生成", func(t *testing.T) {
		comment := "テストコメント"
		generator := newMockCommentGenerator(comment)
		assert.NotNil(t, generator)
		assert.Equal(t, comment, generator.comment)
	})

	t.Run("正常系_空コメントで生成", func(t *testing.T) {
		generator := newMockCommentGenerator("")
		assert.NotNil(t, generator)
		assert.Equal(t, "", generator.comment)
	})
}

func TestMockCommentGenerator_Generate(t *testing.T) {
	now := time.Now()
	article := &entity.Article{
		Title:     "テスト記事",
		Link:      "https://example.com/test",
		Published: &now,
		Content:   "テスト記事の内容",
	}

	t.Run("正常系_設定されたコメントを返す", func(t *testing.T) {
		expectedComment := "これはモックコメントです"
		generator := newMockCommentGenerator(expectedComment)

		result, err := generator.Generate(context.Background(), article)
		assert.NoError(t, err)
		assert.Equal(t, expectedComment, result)
	})

	t.Run("正常系_空コメントを返す", func(t *testing.T) {
		generator := newMockCommentGenerator("")

		result, err := generator.Generate(context.Background(), article)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("正常系_日本語コメントを返す", func(t *testing.T) {
		expectedComment := "この記事は面白いですね！おすすめです。"
		generator := newMockCommentGenerator(expectedComment)

		result, err := generator.Generate(context.Background(), article)
		assert.NoError(t, err)
		assert.Equal(t, expectedComment, result)
	})

	t.Run("正常系_nilの記事でも動作する", func(t *testing.T) {
		expectedComment := "固定コメント"
		generator := newMockCommentGenerator(expectedComment)

		result, err := generator.Generate(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedComment, result)
	})
}
