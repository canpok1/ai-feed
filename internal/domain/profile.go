package domain

import (
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ProfileRepository はプロファイルの永続化を担当するインターフェース
type ProfileRepository interface {
	LoadProfile() (*entity.Profile, error)
	SaveProfileWithTemplate() error
}
