package domain

import "time"

// Article represents a single article in a feed.
type Article struct {
	Title     string
	Link      string
	Published *time.Time
	Content   string
}

type Recommend struct {
	Article Article
	Comment *string
}

type Config struct {
}

func MakeDefaultConfig() *Config {
	return &Config{}
}
