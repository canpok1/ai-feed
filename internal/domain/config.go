package domain

import "github.com/canpok1/ai-feed/internal/domain/entity"

// ConfigInitRepository は設定ファイルの初期化を担当するインターフェース
type ConfigInitRepository interface {
	// SaveWithTemplate はテンプレートを使用して設定ファイルを保存する
	SaveWithTemplate() error
}

// ConfigRepository は設定ファイルの読み込みを担当するインターフェース
type ConfigRepository interface {
	// Load は設定ファイルを読み込む
	Load() (*LoadedConfig, error)
}

// LoadedConfig は読み込まれた設定を表す構造体
type LoadedConfig struct {
	// DefaultProfile はデフォルトプロファイル設定
	DefaultProfile *entity.Profile
	// Cache はキャッシュ設定
	Cache *entity.CacheConfig
}

// ValidatorFactory は設定のバリデーターを作成するファクトリインターフェース
type ValidatorFactory interface {
	// Create はバリデーターを作成する
	Create(config *LoadedConfig, profile *entity.Profile) Validator
}
