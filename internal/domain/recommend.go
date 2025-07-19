package domain

import (
	"math/rand/v2"
)

type Recommender interface {
	Recommend(articles []Article) (*Recommend, error)
}

type RandomRecommender struct{}

func NewRandomRecommender() Recommender {
	return &RandomRecommender{}
}

func (r *RandomRecommender) Recommend(articles []Article) (*Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[rand.IntN(len(articles))]
	return &Recommend{
		Article: article,
		Comment: nil,
	}, nil
}

type FirstRecommender struct{}

func NewFirstRecommender() Recommender {
	return &FirstRecommender{}
}

func (r *FirstRecommender) Recommend(articles []Article) (*Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	article := articles[0]
	return &Recommend{
		Article: article,
		Comment: nil,
	}, nil
}
