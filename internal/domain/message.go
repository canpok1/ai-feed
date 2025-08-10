package domain

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type MessageSender interface {
	SendRecommend(*entity.Recommend, string) error
}
