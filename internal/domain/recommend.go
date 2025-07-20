package domain

import (
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type Recommender interface {
	Recommend(articles []entity.Article) (*entity.Recommend, error)
}

type RandomRecommender struct{}

func NewRandomRecommender() Recommender {
	return &RandomRecommender{}
}

func (r *RandomRecommender) Recommend(articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[rand.IntN(len(articles))]
	return &entity.Recommend{
		Article: article,
		Comment: nil,
	}, nil
}

type FirstRecommender struct{}

func NewFirstRecommender() Recommender {
	return &FirstRecommender{}
}

func (r *FirstRecommender) Recommend(articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	return &entity.Recommend{
		Article: article,
		Comment: nil,
	}, nil
}
