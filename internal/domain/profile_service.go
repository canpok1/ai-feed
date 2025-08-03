package domain

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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
	// ResolvePath はパス文字列を解決する
	ResolvePath(path string) (string, error)
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
	// パス解決
	resolvedPath, err := s.ResolvePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// ファイルの存在確認
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile file not found at %s", resolvedPath)
	} else if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// プロファイルファイルの読み込み
	profileRepo := s.repoFactory(resolvedPath)
	profile, err := profileRepo.LoadProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// バリデーション実行
	result := s.validator.Validate(profile)
	return result, nil
}

// ResolvePath はパス文字列を解決する（ホームディレクトリや環境変数を展開）
func (s *ProfileServiceImpl) ResolvePath(path string) (string, error) {
	// 環境変数の展開
	path = os.ExpandEnv(path)

	// ホームディレクトリの展開
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}