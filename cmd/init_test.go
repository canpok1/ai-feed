package cmd

import (
	"bytes"
	"os"
	"strings"
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

// TestInitCommand_ProgressMessages は進行状況メッセージの出力をテストする
func TestInitCommand_ProgressMessages(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(originalWd))
	})

	cmd := makeInitCmd()

	// stderr と stdout をキャプチャ
	stderrBuf := &bytes.Buffer{}
	stdoutBuf := &bytes.Buffer{}
	cmd.SetErr(stderrBuf)
	cmd.SetOut(stdoutBuf)

	// コマンドを実行
	err = cmd.Execute()
	assert.NoError(t, err)

	// stderr に進行状況メッセージが出力されていることを確認
	stderrOutput := stderrBuf.String()
	assert.Contains(t, stderrOutput, "設定ファイルを初期化しています...", "Progress message should be output to stderr")
	assert.Contains(t, stderrOutput, "設定テンプレートを生成しています...", "Template generation message should be output to stderr")

	// stdout に完了メッセージが出力されていることを確認
	stdoutOutput := stdoutBuf.String()
	assert.Contains(t, stdoutOutput, "./config.yml を生成しました", "Completion message should be output to stdout")

	// メッセージの順序を確認
	stderrLines := strings.Split(strings.TrimSpace(stderrOutput), "\n")
	assert.Equal(t, "設定ファイルを初期化しています...", stderrLines[0], "First progress message should be initialization")
	assert.Equal(t, "設定テンプレートを生成しています...", stderrLines[1], "Second progress message should be template generation")
}
