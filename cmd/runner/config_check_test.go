package runner

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

func TestNewConfigCheckRunner(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
	}{
		{
			name:       "正常系: 新しいインスタンスを作成できる",
			configPath: "config.yml",
		},
		{
			name:       "正常系: 空のconfigPathでも作成できる",
			configPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			runner := NewConfigCheckRunner(tt.configPath, stdout, stderr)

			require.NotNil(t, runner)
			assert.Equal(t, tt.configPath, runner.configPath)
			assert.Equal(t, stdout, runner.stdout)
			assert.Equal(t, stderr, runner.stderr)
		})
	}
}

func TestConfigCheckRunner_Run(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(t *testing.T, tmpDir string) (configPath string, profilePath string, verboseFlag bool)
		wantErr         bool
		errContains     string
		wantStdoutConta string
		wantStderrConta string
	}{
		{
			name: "正常系: config.ymlのみで検証成功",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
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
				return configPath, "", false
			},
			wantErr:         false,
			wantStdoutConta: "設定の検証が完了しました",
		},
		{
			name: "正常系: プロファイルファイル指定で検証成功",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
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

				return configPath, profilePath, false
			},
			wantErr:         false,
			wantStdoutConta: "設定の検証が完了しました",
		},
		{
			name: "正常系: verbose=trueでサマリー表示",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
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
				return configPath, "", true
			},
			wantErr:         false,
			wantStdoutConta: "【設定サマリー】",
		},
		{
			name: "異常系: 必須フィールド不足でエラー",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  fixed_message: "テストメッセージ"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)
				return configPath, "", false
			},
			wantErr:         true,
			wantStderrConta: "以下の問題があります",
			errContains:     "設定ファイルのバリデーションに失敗しました",
		},
		{
			name: "異常系: プロファイルファイルが存在しない",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
				configPath := filepath.Join(tmpDir, "config.yml")
				configContent := `
default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"
      api_key: "test-key"
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)
				return configPath, filepath.Join(tmpDir, "not-exist.yml"), false
			},
			wantErr:         true,
			wantStderrConta: "プロファイルファイルの読み込みに失敗しました",
			errContains:     "failed to load profile",
		},
		{
			name: "異常系: 設定ファイルが存在しない",
			setup: func(t *testing.T, tmpDir string) (string, string, bool) {
				return filepath.Join(tmpDir, "not-exist.yml"), "", false
			},
			wantErr:         true,
			wantStderrConta: "設定ファイルの読み込みに失敗しました",
			errContains:     "failed to load config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath, profilePath, verboseFlag := tt.setup(t, tmpDir)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// ProfileRepositoryのファクトリ関数
			profileRepoFn := func(path string) domain.ProfileRepository {
				return profile.NewYamlProfileRepositoryImpl(path)
			}

			runner := NewConfigCheckRunner(configPath, stdout, stderr)
			params := &ConfigCheckParams{
				ProfilePath: profilePath,
				VerboseFlag: verboseFlag,
			}

			err := runner.Run(params, profileRepoFn)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.wantStdoutConta != "" {
				assert.Contains(t, stdout.String(), tt.wantStdoutConta)
			}

			if tt.wantStderrConta != "" {
				assert.Contains(t, stderr.String(), tt.wantStderrConta)
			}
		})
	}
}
