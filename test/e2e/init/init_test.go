//go:build e2e

package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/test/e2e/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestInitCommand_CreateConfigFile はai-feed initコマンドが設定ファイルを正常に作成できることを確認するテスト
func TestInitCommand_CreateConfigFile(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	tests := []struct {
		name              string
		setupFunc         func(t *testing.T, tmpDir string)
		wantOutputContain string
		wantError         bool
	}{
		{
			name: "設定ファイルが存在しない場合、新規作成される",
			setupFunc: func(t *testing.T, tmpDir string) {
				// 何もしない（クリーンな状態）
			},
			wantOutputContain: "config.yml を生成しました",
			wantError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			// セットアップ処理を実行
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// 一時ディレクトリに移動
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "init")

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

			// エラーがない場合のみファイルの存在とYAMLの妥当性をチェック
			if !tt.wantError {
				// 設定ファイルが作成されたことを確認
				configPath := filepath.Join(tmpDir, "config.yml")
				_, err := os.Stat(configPath)
				assert.NoError(t, err, "config.ymlファイルが作成されているはずです")

				// ファイルの内容が有効なYAMLであることを確認
				content, err := os.ReadFile(configPath)
				require.NoError(t, err, "config.ymlの読み取りに成功するはずです")

				var config map[string]interface{}
				err = yaml.Unmarshal(content, &config)
				assert.NoError(t, err, "YAMLのパースに成功するはずです")

				// 必要なフィールドが含まれていることを確認
				assert.Contains(t, config, "default_profile", "default_profileフィールドが含まれているはずです")

				// default_profileの中身を確認
				defaultProfile, ok := config["default_profile"].(map[string]interface{})
				require.True(t, ok, "default_profileはマップであるはずです")
				assert.Contains(t, defaultProfile, "ai", "default_profile.aiフィールドが含まれているはずです")
				assert.Contains(t, defaultProfile, "output", "default_profile.outputフィールドが含まれているはずです")

				// ファイル内容が空でないことを確認
				assert.Greater(t, len(content), 0, "ファイル内容が空でないはずです")

				// コメントが含まれていることを確認（テンプレートの特徴）
				contentStr := string(content)
				assert.True(t, strings.Contains(contentStr, "#"), "YAMLコメントが含まれているはずです（テンプレート使用の証明）")
			}
		})
	}
}

// TestInitCommand_ExistingFile は既存のconfig.ymlファイルがある場合の動作を確認するテスト
func TestInitCommand_ExistingFile(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	tests := []struct {
		name              string
		existingContent   string
		wantErrorContain  string
		wantFilePreserved bool
	}{
		{
			name:              "既存ファイルがある場合、エラーが発生し既存ファイルは保護される",
			existingContent:   "existing config content",
			wantErrorContain:  "config file already exists",
			wantFilePreserved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			// 既存のconfig.ymlを作成
			configPath := filepath.Join(tmpDir, "config.yml")
			err := os.WriteFile(configPath, []byte(tt.existingContent), 0644)
			require.NoError(t, err, "テスト用の既存ファイル作成に成功するはずです")

			// 一時ディレクトリに移動
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "init")

			// エラーが発生することを確認
			assert.Error(t, err, "既存ファイルがある場合エラーが発生するはずです")

			// エラーメッセージの確認
			if tt.wantErrorContain != "" {
				assert.Contains(t, output, tt.wantErrorContain, "期待されるエラーメッセージが含まれているはずです")
			}

			// 既存ファイルが保護されているか確認
			if tt.wantFilePreserved {
				content, err := os.ReadFile(configPath)
				require.NoError(t, err, "既存ファイルの読み取りに成功するはずです")
				assert.Equal(t, tt.existingContent, string(content), "既存ファイルの内容が変更されていないはずです")
			}
		})
	}
}
