package domain

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ProfileRepository はプロファイルの永続化を担当するインターフェース
type ProfileRepository interface {
	LoadProfile() (*entity.Profile, error)
}

// ProfileService はプロファイル操作のサービス層インターフェース
type ProfileService interface {
	// ValidateProfile はプロファイルファイルを読み込み、バリデーションを実行する
	ValidateProfile(path string) (*ValidationResult, error)
}

// ProfileServiceImpl はProfileServiceの実装
type ProfileServiceImpl struct {
	validator  ProfileValidator
	repoFactory func(string) ProfileRepository
}

// NewProfileService はProfileServiceImplの新しいインスタンスを作成する
func NewProfileService(validator ProfileValidator, repoFactory func(string) ProfileRepository) ProfileService {
	return &ProfileServiceImpl{
		validator:   validator,
		repoFactory: repoFactory,
	}
}

// ValidateProfile はプロファイルファイルを読み込み、バリデーションを実行する
func (s *ProfileServiceImpl) ValidateProfile(path string) (*ValidationResult, error) {
	// ファイルの存在確認
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile file not found at %s", path)
	} else if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// プロファイルファイルの読み込み
	profileRepo := s.repoFactory(path)
	profile, err := profileRepo.LoadProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// バリデーション実行
	result := s.validator.Validate(profile)
	return result, nil
}

