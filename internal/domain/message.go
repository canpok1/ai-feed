package domain

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// MessageSender はメッセージ送信を行うインターフェース
type MessageSender interface {
	// SendRecommend は推薦記事を送信する
	SendRecommend(*entity.Recommend, string) error
	// ServiceName はサービス名を返す（ログ表示用）
	ServiceName() string
}
