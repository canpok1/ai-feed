package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
)

func TestMakeConfigCmd(t *testing.T) {
	cmd := makeConfigCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.True(t, cmd.HasSubCommands())
}

func TestMakeConfigCheckCmd(t *testing.T) {
	cmd := makeConfigCheckCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
	assert.True(t, cmd.Flags().HasFlags())
}

func TestConfigCheckCmd_Success(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 有効なconfig.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "valid-api-key-12345",
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "test prompt template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "設定に問題ありません")
	assert.Empty(t, stderr.String())
}

func TestConfigCheckCmd_WithVerbose(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 有効なconfig.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "valid-api-key-12345",
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "test prompt template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath
	cmd.SetArgs([]string{"--verbose"})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "設定に問題ありません")
	assert.Contains(t, output, "【設定サマリー】")
	assert.Contains(t, output, "AI設定:")
	assert.Contains(t, output, "Gemini API: 設定済み")
	assert.Contains(t, output, "gemini-1.5-flash")
	assert.Empty(t, stderr.String())
}

func TestConfigCheckCmd_ValidationError(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// ダミー値を含むconfig.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "xxxxxx", // ダミー値
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "test prompt template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "設定に以下の問題があります")
	assert.Contains(t, stderrOutput, "ai.gemini.api_key")
	assert.Contains(t, stderrOutput, "ダミー値")
}

func TestConfigCheckCmd_WithProfile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// ダミー値を含むconfig.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "xxxxxx", // ダミー値
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "test prompt template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// プロファイルファイルを作成
	profilePath := filepath.Join(tmpDir, "profile.yml")
	profileFile, err := os.Create(profilePath)
	assert.NoError(t, err)
	defer profileFile.Close()

	profileContent := `ai:
  gemini:
    api_key: valid-profile-api-key
`
	_, err = profileFile.WriteString(profileContent)
	assert.NoError(t, err)
	profileFile.Close()

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath
	cmd.SetArgs([]string{"--profile", profilePath})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "設定に問題ありません")
	assert.Empty(t, stderr.String())
}

func TestConfigCheckCmd_ConfigNotFound(t *testing.T) {
	// 存在しないファイルパス
	configPath := "/nonexistent/config.yml"

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "設定ファイルの読み込みに失敗しました")
}

func TestConfigCheckCmd_ProfileNotFound(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 有効なconfig.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "valid-api-key-12345",
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "test prompt template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// 存在しないプロファイルパス
	profilePath := "/nonexistent/profile.yml"

	// コマンドを実行
	cmd := makeConfigCheckCmd()
	cfgFile = configPath
	cmd.SetArgs([]string{"--profile", profilePath})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "プロファイルファイルの読み込みに失敗しました")
}

func TestConfigCheckCmd_Help(t *testing.T) {
	cmd := makeConfigCheckCmd()
	cmd.SetArgs([]string{"--help"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "check")
	assert.Contains(t, output, "Flags:")
	// --profile フラグの確認
	assert.True(t, strings.Contains(output, "--profile") || strings.Contains(output, "-p"))
	// --verbose フラグの確認
	assert.True(t, strings.Contains(output, "--verbose") || strings.Contains(output, "-v"))
}
