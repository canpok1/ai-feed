package domain

import (
	"context"
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(context.Context, *entity.AIModelConfig, *entity.PromptConfig, []entity.Article) (*entity.Recommend, error)
}

type CommentGenerator interface {
	Generate(context.Context, *entity.Article) (string, error)
}

type CommentGeneratorFactory interface {
	MakeCommentGenerator(*entity.AIModelConfig, *entity.PromptConfig) (CommentGenerator, error)
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
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[rand.IntN(len(articles))]
	comment, err := generateComment(r.factory, model, prompt, ctx, &article)
	if err != nil {
		return nil, err
	}

	return &entity.Recommend{
		Article: article,
		Comment: comment,
	}, nil
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
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	comment, err := generateComment(r.factory, model, prompt, ctx, &article)
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
	ctx context.Context,
	article *entity.Article) (*string, error) {
	if factory == nil || model == nil || prompt == nil {
		return nil, nil
	}

	commentGenerator, err := factory.MakeCommentGenerator(model, prompt)
	if err != nil {
		return nil, err
	}

	if commentGenerator == nil {
		return nil, nil
	}

	c, err := commentGenerator.Generate(ctx, article)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
