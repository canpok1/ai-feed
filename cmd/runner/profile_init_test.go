package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileInitRunner_Run(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, tmpDir string) string
		wantErr bool
		verify  func(t *testing.T, filePath string)
	}{
		{
			name: "新規ファイル作成成功",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "test_profile.yml")
			},
			wantErr: false,
			verify: func(t *testing.T, filePath string) {
				// ファイルが作成されたことを確認
				_, err := os.Stat(filePath)
				assert.NoError(t, err, "プロファイルファイルが作成されているべき")

				// ファイルの内容を確認
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), "# AI Feedのプロファイル設定ファイル", "テンプレートコメントが含まれているべき")
			},
		},
		{
			name: "既存ファイルが存在する場合はエラー",
			setup: func(t *testing.T, tmpDir string) string {
				filePath := filepath.Join(tmpDir, "existing_profile.yml")
				err := os.WriteFile(filePath, []byte("existing content"), 0644)
				require.NoError(t, err)
				return filePath
			},
			wantErr: true,
			verify: func(t *testing.T, filePath string) {
				// ファイルが変更されていないことを確認
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), "existing content", "既存の内容は変更されないべき")
			},
		},
		{
			name: "ディレクトリが存在しない場合はエラー",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "nonexistent", "profile.yml")
			},
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の一時ディレクトリを作成
			tmpDir := t.TempDir()

			// テストのセットアップ
			filePath := tt.setup(t, tmpDir)

			// ProfileInitRunnerを作成して実行
			runner := NewProfileInitRunner(filePath)
			err := runner.Run()

			// エラーの確認
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 追加の検証
			if tt.verify != nil {
				tt.verify(t, filePath)
			}
		})
	}
}
