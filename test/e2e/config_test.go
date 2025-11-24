//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigCommand_Check_Valid は有効な設定ファイルで検証が成功することを確認するテスト
func TestConfigCommand_Check_Valid(t *testing.T) {
	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := GetProjectRoot(t)

	tests := []struct {
		name              string
		configFileName    string
		wantOutputContain string
		wantError         bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			// テストデータファイルをコピー
			srcPath := filepath.Join(projectRoot, "test", "e2e", "testdata", "configs", tt.configFileName)
			dstPath := filepath.Join(tmpDir, "config.yml")

			srcData, err := os.ReadFile(srcPath)
			require.NoError(t, err, "テストデータファイルの読み込みに成功するはずです")

			err = os.WriteFile(dstPath, srcData, 0644)
			require.NoError(t, err, "設定ファイルのコピーに成功するはずです")

			// 一時ディレクトリに移動
			originalWd, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, os.Chdir(originalWd))
			})

			// コマンドを実行
			output, err := ExecuteCommand(t, binaryPath, "config", "check")

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
		})
	}
}
