package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// モックのCommentGenerator
type mockCommentGenerator struct {
	generateFunc func(context.Context, *entity.Article) (string, error)
}

func (m *mockCommentGenerator) Generate(ctx context.Context, article *entity.Article) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, article)
	}
	return "", nil
}

// モックのCommentGeneratorFactory
type mockCommentGeneratorFactory struct {
	makeFunc func(*entity.AIConfig, *entity.PromptConfig) (CommentGenerator, error)
}

func (m *mockCommentGeneratorFactory) MakeCommentGenerator(aiConfig *entity.AIConfig, promptConfig *entity.PromptConfig) (CommentGenerator, error) {
	if m.makeFunc != nil {
		return m.makeFunc(aiConfig, promptConfig)
	}
	return nil, nil
}

func TestNewRandomRecommender(t *testing.T) {
	t.Run("コンストラクタが正しく動作する", func(t *testing.T) {
		factory := &mockCommentGeneratorFactory{}
		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}

		recommender := NewRandomRecommender(factory, aiConfig, promptConfig)

		assert.NotNil(t, recommender)
		randomRecommender, ok := recommender.(*RandomRecommender)
		assert.True(t, ok)
		assert.Equal(t, factory, randomRecommender.factory)
		assert.Equal(t, aiConfig, randomRecommender.aiConfig)
		assert.Equal(t, promptConfig, randomRecommender.promptConfig)
	})
}

func TestRandomRecommender_Recommend(t *testing.T) {
	ctx := context.Background()

	t.Run("記事が1つの場合_コメントなし", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		recommender := NewRandomRecommender(nil, nil, nil)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, articles[0], result.Article)
		assert.Nil(t, result.Comment)
	})

	t.Run("記事が複数の場合_コメントなし", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
			{Title: "Article 2", Link: "https://example.com/2"},
			{Title: "Article 3", Link: "https://example.com/3"},
		}

		recommender := NewRandomRecommender(nil, nil, nil)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		// いずれかの記事が選択されることを確認
		found := false
		for _, article := range articles {
			if result.Article.Link == article.Link {
				found = true
				break
			}
		}
		assert.True(t, found, "選択された記事が入力記事のいずれかであること")
		assert.Nil(t, result.Comment)
	})

	t.Run("記事が1つの場合_コメントあり", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedComment := "テストコメント"
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return &mockCommentGenerator{
					generateFunc: func(ctx context.Context, article *entity.Article) (string, error) {
						return expectedComment, nil
					},
				}, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewRandomRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, articles[0], result.Article)
		require.NotNil(t, result.Comment)
		assert.Equal(t, expectedComment, *result.Comment)
	})

	t.Run("記事配列が空の場合", func(t *testing.T) {
		articles := []entity.Article{}

		recommender := NewRandomRecommender(nil, nil, nil)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no articles found")
	})

	t.Run("CommentGenerator生成に失敗した場合", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedErr := errors.New("generator作成エラー")
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return nil, expectedErr
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewRandomRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("コメント生成に失敗した場合", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedErr := errors.New("コメント生成エラー")
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return &mockCommentGenerator{
					generateFunc: func(ctx context.Context, article *entity.Article) (string, error) {
						return "", expectedErr
					},
				}, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewRandomRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNewFirstRecommender(t *testing.T) {
	t.Run("コンストラクタが正しく動作する", func(t *testing.T) {
		factory := &mockCommentGeneratorFactory{}
		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}

		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)

		assert.NotNil(t, recommender)
		firstRecommender, ok := recommender.(*FirstRecommender)
		assert.True(t, ok)
		assert.Equal(t, factory, firstRecommender.factory)
		assert.Equal(t, aiConfig, firstRecommender.aiConfig)
		assert.Equal(t, promptConfig, firstRecommender.promptConfig)
	})
}

func TestFirstRecommender_Recommend(t *testing.T) {
	ctx := context.Background()

	t.Run("記事が1つの場合", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedComment := "テストコメント"
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return &mockCommentGenerator{
					generateFunc: func(ctx context.Context, article *entity.Article) (string, error) {
						return expectedComment, nil
					},
				}, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, articles[0], result.Article)
		require.NotNil(t, result.Comment)
		assert.Equal(t, expectedComment, *result.Comment)
	})

	t.Run("記事が複数の場合_最初の記事が選ばれる", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
			{Title: "Article 2", Link: "https://example.com/2"},
			{Title: "Article 3", Link: "https://example.com/3"},
		}

		expectedComment := "テストコメント"
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return &mockCommentGenerator{
					generateFunc: func(ctx context.Context, article *entity.Article) (string, error) {
						return expectedComment, nil
					},
				}, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		// 最初の記事が選択されることを確認
		assert.Equal(t, articles[0], result.Article)
		assert.Equal(t, "Article 1", result.Article.Title)
		require.NotNil(t, result.Comment)
		assert.Equal(t, expectedComment, *result.Comment)
	})

	t.Run("記事配列が空の場合_nilを返す", func(t *testing.T) {
		articles := []entity.Article{}

		factory := &mockCommentGeneratorFactory{}
		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("CommentGenerator生成に失敗した場合", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedErr := errors.New("generator作成エラー")
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return nil, expectedErr
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("コメント生成に失敗した場合", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedErr := errors.New("コメント生成エラー")
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return &mockCommentGenerator{
					generateFunc: func(ctx context.Context, article *entity.Article) (string, error) {
						return "", expectedErr
					},
				}, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		recommender := NewFirstRecommender(factory, aiConfig, promptConfig)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}
