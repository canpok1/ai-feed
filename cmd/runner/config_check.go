package runner

import (
	"io"
)

// ConfigCheckParams はconfig checkコマンドの実行パラメータを表す構造体
type ConfigCheckParams struct {
	ProfilePath string
	VerboseFlag bool
}

// ConfigCheckRunner はconfig checkコマンドのビジネスロジックを実行する構造体
type ConfigCheckRunner struct {
	configPath string
	stdout     io.Writer
	stderr     io.Writer
}

// NewConfigCheckRunner はConfigCheckRunnerの新しいインスタンスを作成する
func NewConfigCheckRunner(configPath string, stdout io.Writer, stderr io.Writer) *ConfigCheckRunner {
	return &ConfigCheckRunner{
		configPath: configPath,
		stdout:     stdout,
		stderr:     stderr,
	}
}
