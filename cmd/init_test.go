package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitCommand_NoArguments は引数なしでのコマンド実行をテストする
// コマンドメタデータの検証を行う
func TestInitCommand_NoArguments(t *testing.T) {
	cmd := makeInitCmd()

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドの設定を確認
	assert.Equal(t, "init", cmd.Use, "Command Use should be 'init'")
	assert.Contains(t, cmd.Short, "設定ファイル（config.yml）のテンプレートを生成します", "Short description should mention config.yml generation")
	assert.Contains(t, cmd.Short, "既存ファイルは上書きしません", "Short description should mention file protection")
}
