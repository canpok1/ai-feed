package domain

import "github.com/canpok1/ai-feed/internal/domain/entity"

type ConfigRepository interface {
	Save(config *entity.Config) error
	Load() (*entity.Config, error)
}
