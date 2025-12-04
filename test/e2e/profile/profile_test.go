//go:build e2e

package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/test/e2e/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// testdataDir はprofileテスト用のテストデータディレクトリパス
const testdataDir = "test/e2e/profile/testdata"

// assertProfileCheckOutput はprofile checkコマンドの出力を検証するヘルパー関数
func assertProfileCheckOutput(t *testing.T, output string, err error, wantError bool, wantOutputContain string, wantErrorContains []string) {
	t.Helper()

	// エラー確認
	if wantError {
		assert.Error(t, err, "エラーが発生するはずです")
	} else {
		assert.NoError(t, err, "エラーは発生しないはずです")
	}

	// 出力メッセージの確認
	if wantOutputContain != "" {
		assert.Contains(t, output, wantOutputContain, "期待される出力メッセージが含まれているはずです")
	}

	// 具体的なエラーメッセージの確認
	for _, expectedError := range wantErrorContains {
		assert.Contains(t, output, expectedError, "期待されるエラーメッセージが含まれているはずです")
	}
}

// TestProfileCommand_Init は profile init コマンドのe2eテスト
func TestProfileCommand_Init(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

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
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "profile", "init", profilePath)

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
				assert.Contains(t, parsedContent, "ai", "AIセクションが含まれているはずです")
				assert.Contains(t, parsedContent, "system_prompt", "system_promptが含まれているはずです")
				assert.Contains(t, parsedContent, "output", "outputセクションが含まれているはずです")
			}
		})
	}
}

// TestProfileCommand_Check は profile check コマンドのe2eテスト
func TestProfileCommand_Check(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	tests := []struct {
		name              string
		profileFileName   string // 空文字列の場合はプロファイルファイルなし
		wantOutputContain string
		wantErrorContains []string // 期待されるエラーメッセージのリスト
		wantError         bool
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
			wantOutputContain: "プロファイルの検証に失敗しました",
			wantErrorContains: []string{
				"AI設定が設定されていません",
				"出力設定が設定されていません",
			},
			wantError: true,
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

			// プロファイルファイルをセットアップ
			profilePath := common.SetupTestDataFile(t, projectRoot, testdataDir, tt.profileFileName, "profile.yml", tmpDir)
			if profilePath == "" {
				// ファイルが存在しないパスを指定
				profilePath = filepath.Join(tmpDir, "nonexistent.yml")
			}

			// 一時ディレクトリに移動
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "profile", "check", profilePath)

			// 出力を検証
			assertProfileCheckOutput(t, output, err, tt.wantError, tt.wantOutputContain, tt.wantErrorContains)
		})
	}
}

// TestProfileCommand_Check_WithConfig は config + profile の統合テストケース
func TestProfileCommand_Check_WithConfig(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	tests := []struct {
		name              string
		configFileName    string // 空文字列の場合は設定ファイルなし
		profileFileName   string
		wantOutputContain string
		wantError         bool
	}{
		{
			name:              "デフォルト設定 + プロファイルで検証が成功する",
			configFileName:    "valid_config.yml",
			profileFileName:   "valid_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
		{
			name:              "カスタム設定 + 最小限のプロファイルで検証が成功する",
			configFileName:    "valid_config.yml",
			profileFileName:   "minimal_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
		{
			name:              "設定ファイル + オーバーライドプロファイルで検証が成功する",
			configFileName:    "valid_config.yml",
			profileFileName:   "override_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
		{
			name:              "設定ファイルなし + 有効なプロファイルで検証が成功する",
			configFileName:    "",
			profileFileName:   "valid_profile.yml",
			wantOutputContain: "プロファイルの検証が完了しました",
			wantError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir := t.TempDir()

			// 設定ファイルをセットアップ（profile/testdata内のconfig用ファイルを使用）
			common.SetupTestDataFile(t, projectRoot, testdataDir, tt.configFileName, "config.yml", tmpDir)

			// プロファイルファイルをセットアップ
			profilePath := common.SetupTestDataFile(t, projectRoot, testdataDir, tt.profileFileName, "profile.yml", tmpDir)

			// 一時ディレクトリに移動
			common.ChangeToTempDir(t, tmpDir)

			// コマンドを実行
			output, err := common.ExecuteCommand(t, binaryPath, "profile", "check", profilePath)

			// 出力を検証（統合テストではwantErrorContainsは空なので、nilを渡す）
			assertProfileCheckOutput(t, output, err, tt.wantError, tt.wantOutputContain, nil)
		})
	}
}

// TestProfileCommand_Check_ConfigFlagIgnoredBugRegression は--configフラグが無視されるバグの再発を防ぐテスト
// issue #238: profile check コマンドで --config フラグが無視される問題の修正確認
func TestProfileCommand_Check_ConfigFlagIgnoredBugRegression(t *testing.T) {
	// バイナリをビルド
	binaryPath := common.BuildBinary(t)

	// プロジェクトルートを取得
	projectRoot := common.GetProjectRoot(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 作業ディレクトリに無効な設定ファイル（config.yml）を配置
	// これは --config フラグが無視された場合に使用される
	common.SetupTestDataFile(t, projectRoot, testdataDir, "invalid_config.yml", "config.yml", tmpDir)

	// 有効な設定ファイルを別の場所に配置（--config フラグで指定する用）
	validConfigPath := common.SetupTestDataFile(t, projectRoot, testdataDir, "valid_config.yml", "valid_config.yml", tmpDir)

	// 一時ディレクトリに移動
	common.ChangeToTempDir(t, tmpDir)

	// --config フラグで有効な設定ファイルを指定してコマンドを実行
	// バグがあった場合、./config.yml（無効な設定）が使用され、エラーになる
	output, err := common.ExecuteCommand(t, binaryPath, "--config", validConfigPath, "profile", "check")

	// cfgFileで指定した有効な設定ファイルが使用されるため、成功するはず
	assert.NoError(t, err, "--configフラグで指定した有効な設定ファイルが使用されるはずです")
	assert.Contains(t, output, "プロファイルの検証が完了しました", "検証成功メッセージが含まれているはずです")
	assert.Contains(t, output, validConfigPath, "--configフラグで指定したパスが出力に含まれているはずです")
}
