package domain

import (
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(articles []entity.Article) (*entity.Recommend, error)
}

type CommentGenerator interface {
	Generate(article entity.Article) (string, error)
}

type RandomRecommender struct {
	commentGenerator CommentGenerator
}

func NewRandomRecommender(g CommentGenerator) Recommender {
	return &RandomRecommender{
		commentGenerator: g,
	}
}

func (r *RandomRecommender) Recommend(articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[rand.IntN(len(articles))]
	var comment *string
	if r.commentGenerator != nil {
		if c, err := r.commentGenerator.Generate(article); err != nil {
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
	commentGenerator CommentGenerator
}

func NewFirstRecommender(g CommentGenerator) Recommender {
	return &FirstRecommender{
		commentGenerator: g,
	}
}

func (r *FirstRecommender) Recommend(articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	var comment *string
	if r.commentGenerator != nil {
		if c, err := r.commentGenerator.Generate(article); err != nil {
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
