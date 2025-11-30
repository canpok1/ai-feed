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

	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err, "Failed to create temp directory")

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
		os.RemoveAll(ctx.tempDir)
	}

	return ctx
}

// setupTestTempDirOnly は一時ディレクトリのみを作成するヘルパー関数（作業ディレクトリは変更しない）
func setupTestTempDirOnly(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err, "Failed to create temp directory")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
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

// TestProfileCheckCommand_Success は正常系のテストを実行する
func TestProfileCheckCommand_Success(t *testing.T) {
	ctx := setupTestWithTempDir(t)
	defer ctx.cleanup()

	// 有効なconfig.ymlを作成
	err := os.WriteFile("./config.yml", []byte(validConfigContent), 0644)
	require.NoError(t, err)

	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{})

	_, err = cmd.ExecuteC()

	assert.NoError(t, err, "Command should succeed with valid default profile")
	assert.Contains(t, stdout.String(), "プロファイルの検証が完了しました", "Should show success message")
	_ = stderr // stderrは進行状況メッセージ用
}

// TestProfileCheckCommand_FileAccessError はファイルアクセスエラーのテストを実行する
func TestProfileCheckCommand_FileAccessError(t *testing.T) {
	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{"nonexistent_profile.yml"})

	_, err := cmd.ExecuteC()

	assert.Error(t, err, "Command should return error for non-existent file")
	assert.Contains(t, err.Error(), "プロファイルファイルが見つかりません", "Error message should be in Japanese")
	_ = stdout
	_ = stderr
}

// TestProfileCheckCommand_NoArguments は引数なしでのコマンド実行をテストする
func TestProfileCheckCommand_NoArguments(t *testing.T) {
	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{})

	_, err := cmd.ExecuteC()

	assert.Error(t, err, "Command should return validation error for empty profile")
	assert.Contains(t, stderr.String(), "プロファイルの検証に失敗しました", "Error message should indicate validation failure")
	_ = stdout
}

// TestProfileCheckCommand_PathResolution はパス解決のテストを実行する
func TestProfileCheckCommand_PathResolution(t *testing.T) {
	tempDir, cleanup := setupTestTempDirOnly(t)
	defer cleanup()

	// 有効なプロファイルファイルを作成
	profilePath := filepath.Join(tempDir, "test_profile.yml")
	profileContent := `ai:
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
	err := os.WriteFile(profilePath, []byte(profileContent), 0644)
	require.NoError(t, err)

	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{profilePath})

	_, err = cmd.ExecuteC()

	assert.NoError(t, err, "Command should succeed with valid profile file path")
	assert.Contains(t, stdout.String(), "プロファイルの検証が完了しました", "Should show success message")
	_ = stderr
}

// TestProfileCheckCommand_WithProfileMerge はプロファイルマージのテストを実行する
func TestProfileCheckCommand_WithProfileMerge(t *testing.T) {
	ctx := setupTestWithTempDir(t)
	defer ctx.cleanup()

	// 部分的なconfig.ymlファイルを作成
	configContent := `default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"
      api_key: "default-api-key"
  system_prompt: "デフォルトシステムプロンプト"
  comment_prompt_template: "デフォルトテンプレート {{TITLE}}"
  selector_prompt: "デフォルト記事選択プロンプト"
  output:
    slack_api:
      api_token: "xoxb-default-token"
      channel: "#default"
      message_template: "{{COMMENT}} {{URL}}"
`
	err := os.WriteFile("./config.yml", []byte(configContent), 0644)
	require.NoError(t, err)

	// プロファイルファイルを作成（プロンプトをオーバーライド、Misskeyを追加）
	profileContent := `system_prompt: "カスタムシステムプロンプト"
comment_prompt_template: "カスタムテンプレート {{TITLE}}"
output:
  misskey:
    api_token: "custom-misskey-token"
    api_url: "https://custom.misskey.social/api"
    message_template: "{{COMMENT}} {{URL}}"
`
	profilePath := "./test_profile.yml"
	err = os.WriteFile(profilePath, []byte(profileContent), 0644)
	require.NoError(t, err)

	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{profilePath})

	_, err = cmd.ExecuteC()

	assert.NoError(t, err, "Command should succeed with merged profile")
	assert.Contains(t, stdout.String(), "プロファイルの検証が完了しました", "Should show success message for merged profile")
	_ = stderr
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
	tempDir, cleanup := setupTestTempDirOnly(t)
	defer cleanup()

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

// TestProfileCheckCommand_ConfigFlagIgnoredBugRegression は--configフラグが無視されるバグの再発を防ぐテスト
// issue #238: profile check コマンドで --config フラグが無視される問題の修正確認
func TestProfileCheckCommand_ConfigFlagIgnoredBugRegression(t *testing.T) {
	// テスト用の一時ディレクトリを2つ作成
	tempDir1, err := os.MkdirTemp("", "profile_test_dir1")
	require.NoError(t, err, "Failed to create temp directory 1")
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "profile_test_dir2")
	require.NoError(t, err, "Failed to create temp directory 2")
	defer os.RemoveAll(tempDir2)

	// tempDir1に有効な設定ファイルを作成
	validConfigPath := filepath.Join(tempDir1, "valid_config.yml")
	err = os.WriteFile(validConfigPath, []byte(validConfigContent), 0644)
	require.NoError(t, err)

	// tempDir2に空の設定ファイルを作成（バリデーション失敗するもの）
	err = os.WriteFile(filepath.Join(tempDir2, "invalid_config.yml"), []byte("# empty config\n"), 0644)
	require.NoError(t, err)

	// 作業ディレクトリをtempDir2に変更（./config.ymlがinvalid_config.ymlになるようにシミュレート）
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	err = os.Chdir(tempDir2)
	require.NoError(t, err, "Failed to change to temp directory 2")
	defer os.Chdir(originalWd)

	// ./config.ymlとしてinvalid設定をコピー
	err = os.WriteFile("./config.yml", []byte("# empty config\n"), 0644)
	require.NoError(t, err)

	// cfgFileに有効な設定ファイルのパスを設定
	cleanupCfgFile := setupCfgFileOverride(t, validConfigPath)
	defer cleanupCfgFile()

	stdout, stderr, cmd := createCmdWithBuffers()
	cmd.SetArgs([]string{})

	_, err = cmd.ExecuteC()

	// cfgFileで指定した有効な設定ファイルが使用されるため、成功するはず
	// バグがあった場合、./config.yml（無効な設定）が使用され、エラーになる
	assert.NoError(t, err, "Command should succeed because cfgFile (--config flag) points to valid config, not ./config.yml")
	assert.Contains(t, stderr.String(), validConfigPath, "Should use the config file specified by --config flag, not ./config.yml")
	_ = stdout
}
