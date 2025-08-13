package runner

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/infra/profile"
)

// ProfileInitRunner はprofile initコマンドのビジネスロジックを実行する構造体
type ProfileInitRunner struct {
	yamlRepo *profile.YamlProfileRepository
}

// NewProfileInitRunner はProfileInitRunnerの新しいインスタンスを作成する
func NewProfileInitRunner(yamlRepo *profile.YamlProfileRepository) *ProfileInitRunner {
	return &ProfileInitRunner{
		yamlRepo: yamlRepo,
	}
}

// Run はprofile initコマンドのビジネスロジックを実行する
func (r *ProfileInitRunner) Run() error {
	// テンプレートファイルを直接生成（明確で直接的）
	err := r.yamlRepo.SaveProfileTemplate()
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	return nil
}
