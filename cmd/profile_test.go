package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProfileCheckCommand_Success は正常系のテストを実行する
func TestProfileCheckCommand_Success(t *testing.T) {
	// 作業ディレクトリを一時的に変更
	originalWd, _ := os.Getwd()
	tempDir, _ := os.MkdirTemp("", "profile_test")
	os.Chdir(tempDir)

	// グローバル変数cfgFileをリセット（テスト間の干渉を防ぐ）
	originalCfgFile := cfgFile
	cfgFile = ""
	defer func() {
		cfgFile = originalCfgFile
		os.Chdir(originalWd)
		os.RemoveAll(tempDir)
	}()

	// 有効なconfig.ymlを作成
	configContent := `default_profile:
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
	err := os.WriteFile("./config.yml", []byte(configContent), 0644)
	assert.NoError(t, err)

	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドを実行
	_, err = cmd.ExecuteC()

	// 成功することを確認
	assert.NoError(t, err, "Command should succeed with valid default profile")

	// 成功メッセージが出力されることを確認
	output := stdout.String()
	assert.Contains(t, output, "プロファイルの検証が完了しました", "Should show success message")
}

// TestProfileCheckCommand_FileAccessError はファイルアクセスエラーのテストを実行する
func TestProfileCheckCommand_FileAccessError(t *testing.T) {
	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// 存在しないプロファイルファイルを指定
	cmd.SetArgs([]string{"nonexistent_profile.yml"})

	// コマンドを実行
	_, err := cmd.ExecuteC()

	// エラーが返されることを確認
	assert.Error(t, err, "Command should return error for non-existent file")

	// 日本語のエラーメッセージが含まれることを確認
	assert.Contains(t, err.Error(), "プロファイルファイルが見つかりません", "Error message should be in Japanese")
}

// TestProfileCheckCommand_NoArguments は引数なしでのコマンド実行をテストする
func TestProfileCheckCommand_NoArguments(t *testing.T) {
	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドを実行
	_, err := cmd.ExecuteC()

	// エラーが返されることを確認（空のプロファイルはバリデーションエラーになる）
	assert.Error(t, err, "Command should return validation error for empty profile")

	// バリデーションエラーメッセージが含まれることを確認
	output := stderr.String()
	assert.Contains(t, output, "プロファイルの検証に失敗しました", "Error message should indicate validation failure")
}

// TestProfileCheckCommand_PathResolution はパス解決のテストを実行する
func TestProfileCheckCommand_PathResolution(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, _ := os.MkdirTemp("", "profile_test")
	defer os.RemoveAll(tempDir)

	// 有効なプロファイルファイルを作成
	profilePath := tempDir + "/test_profile.yml"
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
	assert.NoError(t, err)

	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// 絶対パスでプロファイルファイルを指定
	cmd.SetArgs([]string{profilePath})

	// コマンドを実行
	_, err = cmd.ExecuteC()

	// 成功することを確認
	assert.NoError(t, err, "Command should succeed with valid profile file path")

	// 成功メッセージが出力されることを確認
	output := stdout.String()
	assert.Contains(t, output, "プロファイルの検証が完了しました", "Should show success message")
}

// TestProfileCheckCommand_WithProfileMerge はプロファイルマージのテストを実行する
func TestProfileCheckCommand_WithProfileMerge(t *testing.T) {
	// 作業ディレクトリを一時的に変更
	originalWd, _ := os.Getwd()
	tempDir, _ := os.MkdirTemp("", "profile_test")
	os.Chdir(tempDir)

	// グローバル変数cfgFileをリセット（テスト間の干渉を防ぐ）
	originalCfgFile := cfgFile
	cfgFile = ""
	defer func() {
		cfgFile = originalCfgFile
		os.Chdir(originalWd)
		os.RemoveAll(tempDir)
	}()

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// コマンドライン引数を設定（プロファイルファイルを指定）
	cmd.SetArgs([]string{profilePath})

	// コマンドを実行
	_, err = cmd.ExecuteC()

	// 成功することを確認
	assert.NoError(t, err, "Command should succeed with merged profile")

	// 成功メッセージが出力されることを確認
	output := stdout.String()
	assert.Contains(t, output, "プロファイルの検証が完了しました", "Should show success message for merged profile")
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
	// 作業ディレクトリを一時的に変更
	originalWd, _ := os.Getwd()
	tempDir, _ := os.MkdirTemp("", "profile_test")
	os.Chdir(tempDir)
	defer func() {
		os.Chdir(originalWd)
		os.RemoveAll(tempDir)
	}()

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
	// テスト用の一時ディレクトリを作成
	tempDir, _ := os.MkdirTemp("", "profile_test_config_flag")
	defer os.RemoveAll(tempDir)

	// カスタム設定ファイルのパスを作成
	customConfigPath := tempDir + "/custom_config.yml"

	// 有効な設定ファイルを作成
	configContent := `default_profile:
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
	err := os.WriteFile(customConfigPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// グローバル変数cfgFileにカスタムパスを設定
	originalCfgFile := cfgFile
	cfgFile = customConfigPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// コマンドライン引数を設定（引数なし）
	cmd.SetArgs([]string{})

	// コマンドを実行
	_, err = cmd.ExecuteC()

	// 成功することを確認
	assert.NoError(t, err, "Command should succeed with custom config path via --config flag")

	// 成功メッセージが出力されることを確認
	output := stdout.String()
	assert.Contains(t, output, "プロファイルの検証が完了しました", "Should show success message")

	// 設定ファイルパスが正しく表示されることを確認
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, customConfigPath, "Should display the custom config file path")
}

// TestProfileCheckCommand_ConfigFlagIgnoredBugRegression は--configフラグが無視されるバグの再発を防ぐテスト
// issue #238: profile check コマンドで --config フラグが無視される問題の修正確認
func TestProfileCheckCommand_ConfigFlagIgnoredBugRegression(t *testing.T) {
	// テスト用の一時ディレクトリを2つ作成
	tempDir1, _ := os.MkdirTemp("", "profile_test_dir1")
	tempDir2, _ := os.MkdirTemp("", "profile_test_dir2")
	defer os.RemoveAll(tempDir1)
	defer os.RemoveAll(tempDir2)

	// tempDir1に有効な設定ファイルを作成
	validConfigPath := tempDir1 + "/valid_config.yml"
	validConfigContent := `default_profile:
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
	err := os.WriteFile(validConfigPath, []byte(validConfigContent), 0644)
	assert.NoError(t, err)

	// tempDir2に空の設定ファイルを作成（バリデーション失敗するもの）
	invalidConfigPath := tempDir2 + "/invalid_config.yml"
	err = os.WriteFile(invalidConfigPath, []byte("# empty config\n"), 0644)
	assert.NoError(t, err)

	// 作業ディレクトリをtempDir2に変更（./config.ymlがinvalid_config.ymlになるようにシミュレート）
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir2)
	// ./config.ymlとしてinvalid設定をコピー
	err = os.WriteFile("./config.yml", []byte("# empty config\n"), 0644)
	assert.NoError(t, err)
	defer func() {
		os.Chdir(originalWd)
	}()

	// cfgFileに有効な設定ファイルのパスを設定
	originalCfgFile := cfgFile
	cfgFile = validConfigPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	cmd := makeProfileCheckCmd()

	// 標準出力と標準エラーをキャプチャ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.SetArgs([]string{})

	// コマンドを実行
	_, err = cmd.ExecuteC()

	// cfgFileで指定した有効な設定ファイルが使用されるため、成功するはず
	// バグがあった場合、./config.yml（無効な設定）が使用され、エラーになる
	assert.NoError(t, err, "Command should succeed because cfgFile (--config flag) points to valid config, not ./config.yml")

	// 有効な設定ファイルパスが使用されていることを確認
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, validConfigPath, "Should use the config file specified by --config flag, not ./config.yml")
}
