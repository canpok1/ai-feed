package domain

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(context.Context, []entity.Article) (*entity.Recommend, error)
}

type FirstRecommender struct {
	factory      CommentGeneratorFactory
	aiConfig     *entity.AIConfig
	promptConfig *entity.PromptConfig
}

func NewFirstRecommender(f CommentGeneratorFactory, ai *entity.AIConfig, prompt *entity.PromptConfig) Recommender {
	return &FirstRecommender{
		factory:      f,
		aiConfig:     ai,
		promptConfig: prompt,
	}
}

func (r *FirstRecommender) Recommend(ctx context.Context, articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	comment, err := generateComment(r.factory, r.aiConfig, r.promptConfig, ctx, &article)
	if err != nil {
		return nil, err
	}
	return &entity.Recommend{
		Article: article,
		Comment: comment,
	}, nil
}

// SelectorBasedRecommender は ArticleSelector を使用して記事を選択するRecommender
type SelectorBasedRecommender struct {
	selector       ArticleSelector
	commentFactory CommentGeneratorFactory
	aiConfig       *entity.AIConfig
	promptConfig   *entity.PromptConfig
}

// NewSelectorBasedRecommender は新しいSelectorBasedRecommenderを作成する
func NewSelectorBasedRecommender(
	selector ArticleSelector,
	factory CommentGeneratorFactory,
	ai *entity.AIConfig,
	prompt *entity.PromptConfig,
) Recommender {
	return &SelectorBasedRecommender{
		selector:       selector,
		commentFactory: factory,
		aiConfig:       ai,
		promptConfig:   prompt,
	}
}

func (r *SelectorBasedRecommender) Recommend(ctx context.Context, articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found")
	}

	// セレクターに記事選択を委譲
	slog.Info("Starting article selection", "article_count", len(articles))
	article, err := r.selector.Select(ctx, articles)
	if err != nil {
		slog.Error("Failed to select article", "error", err, "article_count", len(articles))
		return nil, fmt.Errorf("failed to select article: %w", err)
	}
	slog.Info("Article selected successfully", "title", article.Title, "link", article.Link)

	// コメント生成
	var comment *string
	if r.commentFactory != nil && r.aiConfig != nil && r.promptConfig != nil {
		comment, err = generateComment(r.commentFactory, r.aiConfig, r.promptConfig, ctx, article)
		if err != nil {
			return nil, fmt.Errorf("failed to generate comment: %w", err)
		}
	}

	return &entity.Recommend{
		Article: *article,
		Comment: comment,
	}, nil
}

func generateComment(
	factory CommentGeneratorFactory,
	model *entity.AIConfig,
	prompt *entity.PromptConfig,
	ctx context.Context,
	article *entity.Article) (*string, error) {
	if factory == nil || model == nil || prompt == nil {
		return nil, fmt.Errorf("factory, model, or prompt is nil")
	}

	commentGenerator, err := factory.MakeCommentGenerator(model, prompt)
	if err != nil {
		return nil, err
	}

	if commentGenerator == nil {
		return nil, fmt.Errorf("comment generator is nil")
	}

	c, err := commentGenerator.Generate(ctx, article)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
