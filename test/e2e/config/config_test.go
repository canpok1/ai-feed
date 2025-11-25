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
