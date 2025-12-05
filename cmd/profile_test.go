package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testContext はテストのセットアップ情報を保持する構造体
type testContext struct {
	tempDir         string
	originalWd      string
	originalCfgFile string
	cleanup         func()
}

// setupTestWithTempDir は一時ディレクトリを作成し、作業ディレクトリを変更するヘルパー関数
// cfgFileのリセットも行う
func setupTestWithTempDir(t *testing.T) *testContext {
	t.Helper()

	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	tempDir := t.TempDir()

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	originalCfgFile := cfgFile
	cfgFile = ""

	ctx := &testContext{
		tempDir:         tempDir,
		originalWd:      originalWd,
		originalCfgFile: originalCfgFile,
	}
	ctx.cleanup = func() {
		cfgFile = ctx.originalCfgFile
		os.Chdir(ctx.originalWd)
		// t.TempDir()は自動クリーンアップされるため、os.RemoveAllは不要
	}

	return ctx
}

// setupTestTempDirOnly は一時ディレクトリのみを作成するヘルパー関数（作業ディレクトリは変更しない）
// t.TempDir()を使用するため、明示的なクリーンアップは不要
func setupTestTempDirOnly(t *testing.T) string {
	t.Helper()

	return t.TempDir()
}

// setupCfgFileOverride はcfgFileグローバル変数を一時的にオーバーライドするヘルパー関数
func setupCfgFileOverride(t *testing.T, newValue string) func() {
	t.Helper()

	originalCfgFile := cfgFile
	cfgFile = newValue

	return func() {
		cfgFile = originalCfgFile
	}
}

// createCmdWithBuffers はコマンドを作成し、標準出力と標準エラーをキャプチャするバッファを設定する
func createCmdWithBuffers() (*bytes.Buffer, *bytes.Buffer, *profileCheckCmd) {
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd := makeProfileCheckCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return stdout, stderr, cmd
}

// profileCheckCmd はcobra.Commandのエイリアス（型アサーション用）
type profileCheckCmd = cobra.Command

// validConfigContent は有効な設定ファイルの内容
const validConfigContent = `default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"
      api_key: "test-api-key"
  system_prompt: "テスト用システムプロンプト"
  comment_prompt_template: "テスト用テンプレート {{TITLE}}"
  selector_prompt: "テスト用記事選択プロンプト"
  output:
    slack_api:
      api_token: "xoxb-test-token"
      channel: "#test"
      message_template: "{{COMMENT}} {{URL}}"
`

// TestProfileCheckCommand_NoArguments は引数なしでのコマンド実行をテストする
func TestProfileCheckCommand_NoArguments(t *testing.T) {
	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{})

	_, err := cmd.ExecuteC()

	assert.Error(t, err, "Command should return validation error for empty profile")
	assert.Contains(t, stderr.String(), "以下の問題があります", "Error message should indicate validation failure")
	_ = stdout
}

// TestProfileCheckCommand_AcceptsOptionalArgs は引数がオプショナルになったことをテストする
func TestProfileCheckCommand_AcceptsOptionalArgs(t *testing.T) {
	cmd := makeProfileCheckCmd()

	// 引数なしの場合（エラーは発生するが引数エラーではない）
	cmd.SetArgs([]string{})
	_, err := cmd.ExecuteC()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "プロファイルの検証に失敗しました")

	// 引数ありの場合（存在しないファイル）
	cmd = makeProfileCheckCmd() // 新しいインスタンスを作成
	cmd.SetArgs([]string{"nonexistent.yml"})
	_, err = cmd.ExecuteC()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "プロファイルファイルが見つかりません")
}

// TestProfileCheckCommand_ConfigLoadingBehavior はconfig.yml読み込み動作をテストする
func TestProfileCheckCommand_ConfigLoadingBehavior(t *testing.T) {
	ctx := setupTestWithTempDir(t)
	defer ctx.cleanup()

	cmd := makeProfileCheckCmd()
	cmd.SetArgs([]string{})

	// config.ymlが存在しない場合でもパニックしない
	_, err := cmd.ExecuteC()
	assert.Error(t, err) // バリデーションエラーは発生する
	assert.Contains(t, err.Error(), "プロファイルの検証に失敗しました")
}

// osExit は os.Exit をモック可能にするための変数
var osExit = os.Exit

// TestProfileCheckCommand_WithConfigFlag は--configフラグが正しく参照されることをテストする
func TestProfileCheckCommand_WithConfigFlag(t *testing.T) {
	tempDir := setupTestTempDirOnly(t)

	// カスタム設定ファイルのパスを作成
	customConfigPath := filepath.Join(tempDir, "custom_config.yml")

	// 有効な設定ファイルを作成
	err := os.WriteFile(customConfigPath, []byte(validConfigContent), 0644)
	require.NoError(t, err)

	// グローバル変数cfgFileにカスタムパスを設定
	cleanupCfgFile := setupCfgFileOverride(t, customConfigPath)
	defer cleanupCfgFile()

	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{})

	_, err = cmd.ExecuteC()

	assert.NoError(t, err, "Command should succeed with custom config path via --config flag")
	assert.Contains(t, stdout.String(), "プロファイルの検証が完了しました", "Should show success message")
	assert.Contains(t, stderr.String(), customConfigPath, "Should display the custom config file path")
}
