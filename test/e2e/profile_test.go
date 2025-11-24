//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestProfileCommand_Init は profile init コマンドのe2eテスト
func TestProfileCommand_Init(t *testing.T) {
	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := GetProjectRoot(t)

	tests := []struct {
		name              string
		setupFile         func(string) error // プロファイルファイルのセットアップ
		wantOutputContain string
		wantError         bool
		validateFile      bool // ファイルの内容を検証するか
	}{
		{
			name: "プロファイルファイルが作成される",
			setupFile: func(profilePath string) error {
				// ファイルが存在しない状態にする
				return nil
			},
			wantOutputContain: "プロファイルファイルを作成しました:",
			wantError:         false,
			validateFile:      true,
		},
		{
			name: "既存ファイルがある場合はエラーが発生する",
			setupFile: func(profilePath string) error {
				// 既にファイルが存在する状態を作る
				return os.WriteFile(profilePath, []byte("existing content"), 0644)
			},
			wantOutputContain: "profile file already exists",
			wantError:         true,
			validateFile:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()
			profilePath := filepath.Join(tmpDir, "test_profile.yml")

			// テスト用のファイル状態をセットアップ
			if tt.setupFile != nil {
				err := tt.setupFile(profilePath)
				require.NoError(t, err, "テストのセットアップに成功するはずです")
			}

			// 一時ディレクトリに移動
			originalWd, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, os.Chdir(originalWd))
			})

			// コマンドを実行
			output, err := ExecuteCommand(t, binaryPath, "profile", "init", profilePath)

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

			// ファイルの内容を検証
			if tt.validateFile {
				// ファイルが作成されていることを確認
				_, statErr := os.Stat(profilePath)
				assert.NoError(t, statErr, "プロファイルファイルが作成されているはずです")

				// ファイルの内容を読み込む
				content, readErr := os.ReadFile(profilePath)
				require.NoError(t, readErr, "ファイルの読み込みに成功するはずです")

				// YAMLとしてパース可能か確認
				var parsedContent map[string]interface{}
				yamlErr := yaml.Unmarshal(content, &parsedContent)
				assert.NoError(t, yamlErr, "YAMLフォーマットが正しいはずです")

				// テンプレートに必要なセクションが含まれているか確認
				assert.Contains(t, string(content), "ai:", "AIセクションが含まれているはずです")
				assert.Contains(t, string(content), "system_prompt:", "system_promptが含まれているはずです")
				assert.Contains(t, string(content), "output:", "outputセクションが含まれているはずです")
			}
		})
	}
}

// TestProfileCommand_Check は profile check コマンドのe2eテスト
func TestProfileCommand_Check(t *testing.T) {
	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := GetProjectRoot(t)

	tests := []struct {
		name              string
		profileFileName   string // 空文字列の場合はプロファイルファイルなし
		wantOutputContain string
		wantError         bool
		checkErrorList    bool
	}{
		{
			name:              "有効なプロファイルで検証が成功する",
			profileFileName:   "valid_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
		{
			name:              "最小限のプロファイルで検証が成功する",
			profileFileName:   "minimal_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
		{
			name:              "無効なプロファイルでエラーが検出される",
			profileFileName:   "invalid_profile.yml",
			wantOutputContain: "プロファイルに以下の問題があります：",
			wantError:         true,
			checkErrorList:    true,
		},
		{
			name:              "プロファイルファイルが存在しない場合、エラーが発生する",
			profileFileName:   "",
			wantOutputContain: "プロファイルファイルが見つかりません",
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			var profilePath string
			// profileFileNameが指定されている場合、テストデータファイルをコピー
			if tt.profileFileName != "" {
				srcPath := filepath.Join(projectRoot, "test", "e2e", "testdata", "profiles", tt.profileFileName)
				profilePath = filepath.Join(tmpDir, "profile.yml")

				srcData, err := os.ReadFile(srcPath)
				require.NoError(t, err, "テストデータファイルの読み込みに成功するはずです")

				err = os.WriteFile(profilePath, srcData, 0644)
				require.NoError(t, err, "プロファイルファイルのコピーに成功するはずです")
			} else {
				// ファイルが存在しないパスを指定
				profilePath = filepath.Join(tmpDir, "nonexistent.yml")
			}

			// 一時ディレクトリに移動
			originalWd, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, os.Chdir(originalWd))
			})

			// コマンドを実行
			output, err := ExecuteCommand(t, binaryPath, "profile", "check", profilePath)

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
