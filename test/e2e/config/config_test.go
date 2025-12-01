//go:build e2e

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/test/e2e/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigCommand_Check は config check コマンドのテスト
func TestConfigCommand_Check(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	tests := []struct {
		name              string
		configFileName    string // 空文字列の場合は設定ファイルなし
		wantOutputContain string
		wantError         bool
		checkErrorList    bool
	}{
		{
			name:              "有効な設定ファイルで検証が成功する",
			configFileName:    "valid_config.yml",
			wantOutputContain: "設定に問題ありません。",
			wantError:         false,
		},
		{
			name:              "最小限の設定ファイルで検証が成功する",
			configFileName:    "minimal_config.yml",
			wantOutputContain: "設定に問題ありません。",
			wantError:         false,
		},
		{
			name:              "無効な設定ファイルでエラーが検出される",
			configFileName:    "invalid_config.yml",
			wantOutputContain: "設定に以下の問題があります：",
			wantError:         true,
			checkErrorList:    true,
		},
		{
			name:              "設定ファイルが存在しない場合、エラーが発生する",
			configFileName:    "",
			wantOutputContain: "設定ファイルの読み込みに失敗しました",
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			// configFileNameが指定されている場合、テストデータファイルをコピー
			if tt.configFileName != "" {
				srcPath := filepath.Join(projectRoot, "test", "e2e", "config", "testdata", tt.configFileName)
				dstPath := filepath.Join(tmpDir, "config.yml")

				srcData, err := os.ReadFile(srcPath)
				require.NoError(t, err, "テストデータファイルの読み込みに成功するはずです")

				err = os.WriteFile(dstPath, srcData, 0644)
				require.NoError(t, err, "設定ファイルのコピーに成功するはずです")
			}

			// 一時ディレクトリに移動
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "config", "check")

			// エラー確認
			if tt.wantError {
				assert.Error(t, err, "エラーが発生するはずです")
			} else {
				assert.NoError(t, err, "エラーは発生しないはずです")
			}

			// 出力メッセージの確認
			if tt.wantOutputContain != "" {
				assert.Contains(t, output, tt.wantOutputContain, "期待される出力メッセージが含まれているはずです")
			}

			// エラーリストの確認
			if tt.checkErrorList {
				assert.Contains(t, output, "-", "エラー項目がリスト表示されているはずです")
			}
		})
	}
}

// TestConfigCommand_CheckWithVerbose は config check --verbose コマンドのテスト
func TestConfigCommand_CheckWithVerbose(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 有効な設定ファイルをコピー
	srcPath := filepath.Join(projectRoot, "test", "e2e", "config", "testdata", "valid_config.yml")
	dstPath := filepath.Join(tmpDir, "config.yml")

	srcData, err := os.ReadFile(srcPath)
	require.NoError(t, err, "テストデータファイルの読み込みに成功するはずです")

	err = os.WriteFile(dstPath, srcData, 0644)
	require.NoError(t, err, "設定ファイルのコピーに成功するはずです")

	// 一時ディレクトリに移動
	common.ChangeToTempDir(t, tmpDir)

	// --verbose オプション付きでコマンドを実行
	output, err := common.ExecuteCommand(t, binaryPath, "config", "check", "--verbose")

	// エラーが発生しないことを確認
	assert.NoError(t, err, "エラーは発生しないはずです")

	// 出力メッセージの確認
	assert.Contains(t, output, "設定に問題ありません", "成功メッセージが含まれているはずです")
	assert.Contains(t, output, "【設定サマリー】", "設定サマリーが含まれているはずです")
	assert.Contains(t, output, "AI設定:", "AI設定セクションが含まれているはずです")
	assert.Contains(t, output, "Gemini API: 設定済み", "Gemini API設定状態が含まれているはずです")
	assert.Contains(t, output, "gemini-2.5-flash", "Geminiモデル名が含まれているはずです")
}

// TestConfigCommand_CheckWithProfile は config check --profile コマンドのテスト
func TestConfigCommand_CheckWithProfile(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// ダミー値を含む設定ファイルをコピー
	configSrcPath := filepath.Join(projectRoot, "test", "e2e", "config", "testdata", "invalid_config_for_profile_test.yml")
	configDstPath := filepath.Join(tmpDir, "config.yml")

	configData, err := os.ReadFile(configSrcPath)
	require.NoError(t, err, "設定ファイルの読み込みに成功するはずです")

	err = os.WriteFile(configDstPath, configData, 0644)
	require.NoError(t, err, "設定ファイルのコピーに成功するはずです")

	// プロファイルファイルをコピー
	profileSrcPath := filepath.Join(projectRoot, "test", "e2e", "config", "testdata", "profile_override.yml")
	profileDstPath := filepath.Join(tmpDir, "profile.yml")

	profileData, err := os.ReadFile(profileSrcPath)
	require.NoError(t, err, "プロファイルファイルの読み込みに成功するはずです")

	err = os.WriteFile(profileDstPath, profileData, 0644)
	require.NoError(t, err, "プロファイルファイルのコピーに成功するはずです")

	// 一時ディレクトリに移動
	common.ChangeToTempDir(t, tmpDir)

	// --profile オプション付きでコマンドを実行
	output, err := common.ExecuteCommand(t, binaryPath, "config", "check", "--profile", profileDstPath)

	// エラーが発生しないことを確認（プロファイルでダミー値が上書きされる）
	assert.NoError(t, err, "エラーは発生しないはずです（プロファイルでダミー値が上書きされるため）")

	// 出力メッセージの確認
	assert.Contains(t, output, "設定に問題ありません", "成功メッセージが含まれているはずです")
}

// TestConfigCommand_CheckWithProfileNotFound は存在しないプロファイルを指定した場合のテスト
func TestConfigCommand_CheckWithProfileNotFound(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 有効な設定ファイルをコピー
	srcPath := filepath.Join(projectRoot, "test", "e2e", "config", "testdata", "valid_config.yml")
	dstPath := filepath.Join(tmpDir, "config.yml")

	srcData, err := os.ReadFile(srcPath)
	require.NoError(t, err, "テストデータファイルの読み込みに成功するはずです")

	err = os.WriteFile(dstPath, srcData, 0644)
	require.NoError(t, err, "設定ファイルのコピーに成功するはずです")

	// 一時ディレクトリに移動
	common.ChangeToTempDir(t, tmpDir)

	// 存在しないプロファイルを指定してコマンドを実行
	output, err := common.ExecuteCommand(t, binaryPath, "config", "check", "--profile", "/nonexistent/profile.yml")

	// エラーが発生することを確認
	assert.Error(t, err, "エラーが発生するはずです")

	// 出力メッセージの確認
	assert.Contains(t, output, "プロファイルファイルの読み込みに失敗しました", "プロファイル読み込みエラーメッセージが含まれているはずです")
}
