package domain

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ProfileRepository はプロファイルの読み込みを担当するインターフェース
type ProfileRepository interface {
	LoadProfile() (*entity.Profile, error)
}

// ProfileTemplateRepository はプロファイルテンプレートの保存を担当するインターフェース
type ProfileTemplateRepository interface {
	SaveProfileTemplate() error
}
