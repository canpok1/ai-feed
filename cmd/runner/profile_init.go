package runner

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/infra/profile"
)

// ProfileInitRunner はprofile initコマンドのビジネスロジックを実行する構造体
type ProfileInitRunner struct {
	profileRepo *profile.YamlProfileRepository
}

// NewProfileInitRunner はProfileInitRunnerの新しいインスタンスを作成する
func NewProfileInitRunner(filePath string) *ProfileInitRunner {
	return &ProfileInitRunner{
		profileRepo: profile.NewYamlProfileRepositoryImpl(filePath),
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
