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

// モックのArticleSelector
type mockArticleSelector struct {
	selectFunc func(context.Context, []entity.Article) (*entity.Article, error)
}

func (m *mockArticleSelector) Select(ctx context.Context, articles []entity.Article) (*entity.Article, error) {
	if m.selectFunc != nil {
		return m.selectFunc(ctx, articles)
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
		assert.Contains(t, articles, result.Article, "選択された記事が入力記事のいずれかであること")
		assert.Nil(t, result.Comment)
	})

	t.Run("記事が複数の場合_ランダム性の確認", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
			{Title: "Article 2", Link: "https://example.com/2"},
			{Title: "Article 3", Link: "https://example.com/3"},
		}

		recommender := NewRandomRecommender(nil, nil, nil)

		// 統計的な検証: 十分な回数実行して各記事が少なくとも1回は選ばれることを確認
		selected := make(map[string]bool)
		for i := 0; i < 30; i++ {
			result, err := recommender.Recommend(ctx, articles)
			require.NoError(t, err)
			selected[result.Article.Link] = true
		}

		// すべての記事が少なくとも一度は選択されたことを確認
		assert.Len(t, selected, 3, "十分な回数実行すれば全記事が選ばれるはず")
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
		require.NotNil(t, result.Comment)
		assert.Equal(t, expectedComment, *result.Comment)
	})

	t.Run("記事が1つの場合_コメントなし", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		recommender := NewFirstRecommender(nil, nil, nil)
		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "factory, model, or prompt is nil")
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

func Test_generateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系_正しいパラメータでコメントが生成される", func(t *testing.T) {
		expectedComment := "生成されたコメント"
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
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, aiConfig, promptConfig, ctx, article)

		require.NoError(t, err)
		require.NotNil(t, comment)
		assert.Equal(t, expectedComment, *comment)
	})

	t.Run("異常系_factoryがnil", func(t *testing.T) {
		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(nil, aiConfig, promptConfig, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "factory, model, or prompt is nil")
	})

	t.Run("異常系_modelがnil", func(t *testing.T) {
		factory := &mockCommentGeneratorFactory{}
		promptConfig := &entity.PromptConfig{}
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, nil, promptConfig, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "factory, model, or prompt is nil")
	})

	t.Run("異常系_promptがnil", func(t *testing.T) {
		factory := &mockCommentGeneratorFactory{}
		aiConfig := &entity.AIConfig{}
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, aiConfig, nil, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "factory, model, or prompt is nil")
	})

	t.Run("異常系_CommentGenerator生成に失敗", func(t *testing.T) {
		expectedErr := errors.New("generator生成エラー")
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return nil, expectedErr
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, aiConfig, promptConfig, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("異常系_CommentGeneratorがnil", func(t *testing.T) {
		factory := &mockCommentGeneratorFactory{
			makeFunc: func(ai *entity.AIConfig, prompt *entity.PromptConfig) (CommentGenerator, error) {
				return nil, nil
			},
		}

		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, aiConfig, promptConfig, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "comment generator is nil")
	})

	t.Run("異常系_コメント生成に失敗", func(t *testing.T) {
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
		article := &entity.Article{Title: "Test Article", Link: "https://example.com/test"}

		comment, err := generateComment(factory, aiConfig, promptConfig, ctx, article)

		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNewSelectorBasedRecommender(t *testing.T) {
	t.Run("コンストラクタが正しく動作する", func(t *testing.T) {
		selector := &mockArticleSelector{}
		factory := &mockCommentGeneratorFactory{}
		aiConfig := &entity.AIConfig{}
		promptConfig := &entity.PromptConfig{}

		recommender := NewSelectorBasedRecommender(selector, factory, aiConfig, promptConfig)

		assert.NotNil(t, recommender)
		selectorRecommender, ok := recommender.(*SelectorBasedRecommender)
		assert.True(t, ok)
		assert.Equal(t, selector, selectorRecommender.selector)
		assert.Equal(t, factory, selectorRecommender.commentFactory)
		assert.Equal(t, aiConfig, selectorRecommender.aiConfig)
		assert.Equal(t, promptConfig, selectorRecommender.promptConfig)
	})
}

func TestSelectorBasedRecommender_Recommend(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系_記事選択とコメント生成が成功", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
			{Title: "Article 2", Link: "https://example.com/2"},
			{Title: "Article 3", Link: "https://example.com/3"},
		}

		selectedArticle := &articles[1]
		expectedComment := "AIが生成したコメント"

		selector := &mockArticleSelector{
			selectFunc: func(ctx context.Context, arts []entity.Article) (*entity.Article, error) {
				return selectedArticle, nil
			},
		}

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
		recommender := NewSelectorBasedRecommender(selector, factory, aiConfig, promptConfig)

		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, *selectedArticle, result.Article)
		require.NotNil(t, result.Comment)
		assert.Equal(t, expectedComment, *result.Comment)
	})

	t.Run("正常系_コメント生成なし", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		selectedArticle := &articles[0]

		selector := &mockArticleSelector{
			selectFunc: func(ctx context.Context, arts []entity.Article) (*entity.Article, error) {
				return selectedArticle, nil
			},
		}

		recommender := NewSelectorBasedRecommender(selector, nil, nil, nil)

		result, err := recommender.Recommend(ctx, articles)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, *selectedArticle, result.Article)
		assert.Nil(t, result.Comment)
	})

	t.Run("異常系_記事配列が空", func(t *testing.T) {
		articles := []entity.Article{}

		selector := &mockArticleSelector{}
		recommender := NewSelectorBasedRecommender(selector, nil, nil, nil)

		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no articles found")
	})

	t.Run("異常系_記事選択に失敗", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		expectedErr := errors.New("記事選択エラー")
		selector := &mockArticleSelector{
			selectFunc: func(ctx context.Context, arts []entity.Article) (*entity.Article, error) {
				return nil, expectedErr
			},
		}

		recommender := NewSelectorBasedRecommender(selector, nil, nil, nil)

		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to select article")
	})

	t.Run("異常系_コメント生成に失敗", func(t *testing.T) {
		articles := []entity.Article{
			{Title: "Article 1", Link: "https://example.com/1"},
		}

		selectedArticle := &articles[0]
		expectedErr := errors.New("コメント生成エラー")

		selector := &mockArticleSelector{
			selectFunc: func(ctx context.Context, arts []entity.Article) (*entity.Article, error) {
				return selectedArticle, nil
			},
		}

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
		recommender := NewSelectorBasedRecommender(selector, factory, aiConfig, promptConfig)

		result, err := recommender.Recommend(ctx, articles)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to generate comment")
	})
}
