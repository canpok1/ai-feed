package runner

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal/domain"
)

// DefaultConfigFilePath はデフォルトの設定ファイルパス
const DefaultConfigFilePath = "./config.yml"

// ConfigInitRunner はconfig initコマンドのビジネスロジックを実行する構造体
type ConfigInitRunner struct {
	configRepo domain.ConfigInitRepository
	stderr     io.Writer
}

// NewConfigInitRunner はConfigInitRunnerの新しいインスタンスを作成する
func NewConfigInitRunner(configRepo domain.ConfigInitRepository, stderr io.Writer) *ConfigInitRunner {
	return &ConfigInitRunner{
		configRepo: configRepo,
		stderr:     stderr,
	}
}

// Run はconfig initコマンドのビジネスロジックを実行する
func (r *ConfigInitRunner) Run() error {
	// 進行状況メッセージ: テンプレート生成中
	fmt.Fprintln(r.stderr, "設定テンプレートを生成しています...")

	// テンプレートファイルを生成
	if err := r.configRepo.SaveWithTemplate(); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
}
