package domain

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(context.Context, *entity.AIModelConfig, *entity.PromptConfig, string, []entity.Article) (*entity.Recommend, error)
}

type CommentGenerator interface {
	Generate(context.Context, *entity.Article) (string, error)
}

type CommentGeneratorFactory interface {
	MakeCommentGenerator(*entity.AIModelConfig, *entity.PromptConfig, string) (CommentGenerator, error)
}

type RandomRecommender struct {
	factory CommentGeneratorFactory
}

func NewRandomRecommender(f CommentGeneratorFactory) Recommender {
	return &RandomRecommender{
		factory: f,
	}
}

func (r *RandomRecommender) Recommend(
	ctx context.Context,
	model *entity.AIModelConfig,
	prompt *entity.PromptConfig,
	systemPrompt string,
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found")
	}

	article := articles[rand.IntN(len(articles))]
	recommend := entity.Recommend{
		Article: article,
	}

	if (r.factory != nil) && (model != nil) && (prompt != nil) {
		comment, err := generateComment(r.factory, model, prompt, systemPrompt, ctx, &article)
		if err != nil {
			return nil, err
		}
		recommend.Comment = comment
	}

	return &recommend, nil
}

type FirstRecommender struct {
	factory CommentGeneratorFactory
}

func NewFirstRecommender(f CommentGeneratorFactory) Recommender {
	return &FirstRecommender{
		factory: f,
	}
}

func (r *FirstRecommender) Recommend(
	ctx context.Context,
	model *entity.AIModelConfig,
	prompt *entity.PromptConfig,
	systemPrompt string,
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	comment, err := generateComment(r.factory, model, prompt, systemPrompt, ctx, &article)
	if err != nil {
		return nil, err
	}
	return &entity.Recommend{
		Article: article,
		Comment: comment,
	}, nil
}

func generateComment(
	factory CommentGeneratorFactory,
	model *entity.AIModelConfig,
	prompt *entity.PromptConfig,
	systemPrompt string,
	ctx context.Context,
	article *entity.Article) (*string, error) {
	if factory == nil || model == nil || prompt == nil {
		return nil, fmt.Errorf("factory, model, or prompt is nil")
	}

	commentGenerator, err := factory.MakeCommentGenerator(model, prompt, systemPrompt)
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
