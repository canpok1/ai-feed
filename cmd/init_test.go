package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitCommand_Success は正常系のテストを実行する
func TestInitCommand_Success(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(originalWd))
	})

	cmd := makeInitCmd()

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドを実行
	err = cmd.Execute()

	// 成功することを確認
	assert.NoError(t, err, "Init command should succeed in clean directory")

	// config.ymlファイルが生成されることを確認
	_, err = os.Stat("./config.yml")
	assert.NoError(t, err, "config.yml should be generated")
}

// TestInitCommand_FileExists は既存ファイルがある場合のテストを実行する
func TestInitCommand_FileExists(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(originalWd))
	})

	// 既存のconfig.ymlを作成
	existingContent := "existing content"
	err = os.WriteFile("./config.yml", []byte(existingContent), 0644)
	assert.NoError(t, err)

	cmd := makeInitCmd()

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドを実行
	err = cmd.Execute()

	// エラーが返されることを確認（既存ファイル保護）
	assert.Error(t, err, "Init command should fail when config.yml already exists")

	// 既存ファイルの内容が変更されていないことを確認
	content, err := os.ReadFile("./config.yml")
	assert.NoError(t, err)
	assert.Equal(t, existingContent, string(content), "Existing file should not be overwritten")
}

// TestInitCommand_NoArguments は引数なしでのコマンド実行をテストする
func TestInitCommand_NoArguments(t *testing.T) {
	cmd := makeInitCmd()

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドの設定を確認
	assert.Equal(t, "init", cmd.Use, "Command Use should be 'init'")
	assert.Contains(t, cmd.Short, "設定ファイル（config.yml）のテンプレートを生成します", "Short description should mention config.yml generation")
	assert.Contains(t, cmd.Short, "既存ファイルは上書きしません", "Short description should mention file protection")
}
