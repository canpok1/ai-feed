package domain

import (
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(entity.AIModelConfig, entity.PromptConfig, []entity.Article) (*entity.Recommend, error)
}

type CommentGenerator interface {
	Generate(entity.Article) (string, error)
}

type CommentGeneratorFactory interface {
	MakeCommentGenerator(entity.AIModelConfig, entity.PromptConfig) (CommentGenerator, error)
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
	model entity.AIModelConfig,
	prompt entity.PromptConfig,
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	var commentGenerator CommentGenerator
	if r.factory != nil {
		g, err := r.factory.MakeCommentGenerator(model, prompt)
		if err != nil {
			return nil, err
		}
		commentGenerator = g
	}

	article := articles[rand.IntN(len(articles))]
	var comment *string
	if commentGenerator != nil {
		if c, err := commentGenerator.Generate(article); err != nil {
			return nil, err
		} else {
			comment = &c
		}
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
	model entity.AIModelConfig,
	prompt entity.PromptConfig,
	articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	var commentGenerator CommentGenerator
	if r.factory != nil {
		g, err := r.factory.MakeCommentGenerator(model, prompt)
		if err != nil {
			return nil, err
		}
		commentGenerator = g
	}

	article := articles[0]
	var comment *string
	if commentGenerator != nil {
		if c, err := commentGenerator.Generate(article); err != nil {
			return nil, err
		} else {
			comment = &c
		}
	}
	return &entity.Recommend{
		Article: article,
		Comment: comment,
	}, nil
}
