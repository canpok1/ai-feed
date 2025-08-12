package runner

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
)

// ProfileInitRunner はprofile initコマンドのビジネスロジックを実行する構造体
type ProfileInitRunner struct {
	profileRepo domain.ProfileRepository
}

// NewProfileInitRunner はProfileInitRunnerの新しいインスタンスを作成する
func NewProfileInitRunner(profileRepo domain.ProfileRepository) *ProfileInitRunner {
	return &ProfileInitRunner{
		profileRepo: profileRepo,
	}
}

// Run はprofile initコマンドのビジネスロジックを実行する
func (r *ProfileInitRunner) Run() error {
	// テンプレートを使用してコメント付きprofile.ymlを生成
	err := r.profileRepo.SaveProfileWithTemplate()
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	return nil
}
