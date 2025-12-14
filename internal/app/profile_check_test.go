package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileCheckRunner_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, tmpDir string) (configPath string, profilePath string)
		wantErr     bool
		wantResult  *ProfileCheckResult
		errContains string
	}{
		{
			name: "正常系: config.ymlのみで検証成功",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"
      api_key: "test-key"
  system_prompt: "test system prompt"
  comment_prompt_template: "test template {{TITLE}}"
  selector_prompt: "test selector prompt"
  output:
    slack_api:
      api_token: "xoxb-test-token"
      channel: "#general"
      message_template: "{{COMMENT}} {{URL}}"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)
				return configPath, ""
			},
			wantErr: false,
			wantResult: &ProfileCheckResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "正常系: プロファイルファイル指定で検証成功",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"
      api_key: "config-key"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				profilePath := filepath.Join(tmpDir, "profile.yml")
				profileContent := `
system_prompt: "test system prompt"
comment_prompt_template: "test template {{TITLE}}"
selector_prompt: "test selector prompt"
output:
  slack_api:
    api_token: "xoxb-test-token"
    channel: "#general"
    message_template: "{{COMMENT}} {{URL}}"
`
				err = os.WriteFile(profilePath, []byte(profileContent), 0644)
				require.NoError(t, err)

				return configPath, profilePath
			},
			wantErr: false,
			wantResult: &ProfileCheckResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "異常系: 必須フィールド不足でバリデーションエラー",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  fixed_message: "テストメッセージ"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)
				return configPath, ""
			},
			wantErr: false,
			wantResult: &ProfileCheckResult{
				IsValid: false,
				Errors: []string{
					"AI設定が設定されていません",
					"システムプロンプトが設定されていません",
					"コメントプロンプトテンプレートが設定されていません",
					"記事選択プロンプトが設定されていません",
					"出力設定が設定されていません",
				},
				Warnings: []string{},
			},
		},
		{
			name: "異常系: 存在しないプロファイルファイル",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  ai:
    gemini:
      api_key: "test-key"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				profilePath := filepath.Join(tmpDir, "nonexistent.yml")
				return configPath, profilePath
			},
			wantErr:     true,
			errContains: "プロファイルファイルが見つかりません",
		},
		{
			name: "正常系: config.ymlが存在しない場合でも検証成功",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "nonexistent_config.yml")
				profilePath := filepath.Join(tmpDir, "profile.yml")
				profileContent := `
ai:
  gemini:
    type: "gemini-2.5-flash"
    api_key: "test-key"
system_prompt: "test system prompt"
comment_prompt_template: "test template {{TITLE}}"
selector_prompt: "test selector prompt"
output:
  slack_api:
    api_token: "xoxb-test-token"
    channel: "#general"
    message_template: "{{COMMENT}} {{URL}}"
`
				err := os.WriteFile(profilePath, []byte(profileContent), 0644)
				require.NoError(t, err)
				return configPath, profilePath
			},
			wantErr: false,
			wantResult: &ProfileCheckResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "異常系: 不正なYAML形式のプロファイル",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  ai:
    gemini:
      api_key: "test-key"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				profilePath := filepath.Join(tmpDir, "invalid.yml")
				profileContent := `
invalid yaml content
  - no proper structure
    broken: yaml
`
				err = os.WriteFile(profilePath, []byte(profileContent), 0644)
				require.NoError(t, err)
				return configPath, profilePath
			},
			wantErr:     true,
			errContains: "プロファイルの読み込みに失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// テスト用の一時ディレクトリを作成
			tmpDir := t.TempDir()

			// テストのセットアップ
			configPath, profilePath := tt.setup(t, tmpDir)

			// stderrバッファを作成
			var stderr bytes.Buffer

			// ProfileCheckRunnerを作成して実行
			profileRepoFn := func(path string) domain.ProfileRepository {
				return profile.NewYamlProfileRepositoryImpl(path)
			}
			runner := NewProfileCheckRunner(configPath, &stderr, profileRepoFn)
			result, err := runner.Run(profilePath)

			// エラーの確認
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				// 結果の確認
				if tt.wantResult != nil {
					assert.Equal(t, tt.wantResult.IsValid, result.IsValid)
					assert.Equal(t, tt.wantResult.Errors, result.Errors)
					assert.Equal(t, tt.wantResult.Warnings, result.Warnings)
				}
			}
		})
	}
}
