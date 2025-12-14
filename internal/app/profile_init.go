package app

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal/domain"
)

// ProfileInitRunner はprofile initコマンドのビジネスロジックを実行する構造体
type ProfileInitRunner struct {
	templateRepo domain.ProfileTemplateRepository
	stderr       io.Writer
}

// NewProfileInitRunner はProfileInitRunnerの新しいインスタンスを作成する
func NewProfileInitRunner(templateRepo domain.ProfileTemplateRepository, stderr io.Writer) (*ProfileInitRunner, error) {
	if templateRepo == nil {
		return nil, fmt.Errorf("templateRepo cannot be nil")
	}
	if stderr == nil {
		return nil, fmt.Errorf("stderr cannot be nil")
	}
	return &ProfileInitRunner{
		templateRepo: templateRepo,
		stderr:       stderr,
	}, nil
}

// Run はprofile initコマンドのビジネスロジックを実行する
func (r *ProfileInitRunner) Run() error {
	// 進行状況メッセージ: テンプレート生成中
	fmt.Fprintln(r.stderr, "設定テンプレートを生成しています...")

	// テンプレートファイルを直接生成（明確で直接的）
	err := r.templateRepo.SaveProfileTemplate()
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	return nil
}
