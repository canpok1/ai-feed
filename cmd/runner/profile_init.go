package runner

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal/infra/profile"
)

// ProfileInitRunner はprofile initコマンドのビジネスロジックを実行する構造体
type ProfileInitRunner struct {
	yamlRepo *profile.YamlProfileRepository
	stderr   io.Writer
}

// NewProfileInitRunner はProfileInitRunnerの新しいインスタンスを作成する
func NewProfileInitRunner(yamlRepo *profile.YamlProfileRepository, stderr io.Writer) *ProfileInitRunner {
	return &ProfileInitRunner{
		yamlRepo: yamlRepo,
		stderr:   stderr,
	}
}

// Run はprofile initコマンドのビジネスロジックを実行する
func (r *ProfileInitRunner) Run() error {
	// 進行状況メッセージ: テンプレート生成中
	fmt.Fprintln(r.stderr, "設定テンプレートを生成しています...")

	// テンプレートファイルを直接生成（明確で直接的）
	err := r.yamlRepo.SaveProfileTemplate()
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	return nil
}
