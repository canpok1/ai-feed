package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigInit_WithConfigCheck_EnvBased はconfig initで生成したファイル（環境変数ベース）をconfig checkで検証するテスト
func TestConfigInit_WithConfigCheck_EnvBased(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// DefaultConfigFilePathを一時的に変更するため、cfgFileを使用
	cfgFile = configPath

	// configファイルを生成（initコマンドのテンプレートを使用）
	configRepo := infra.NewYamlConfigRepository(configPath)
	err := configRepo.SaveWithTemplate()
	assert.NoError(t, err)

	// 環境変数を設定（有効なAPIキーをシミュレート）
	t.Setenv("GEMINI_API_KEY", "test-valid-gemini-key")

	// config checkコマンドを実行
	checkCmd := makeConfigCheckCmd()
	var checkStdout bytes.Buffer
	var checkStderr bytes.Buffer
	checkCmd.SetOut(&checkStdout)
	checkCmd.SetErr(&checkStderr)

	err = checkCmd.Execute()

	// 環境変数から読み込む設定なので、バリデーションは成功するはず
	assert.NoError(t, err, "環境変数が設定されていれば、バリデーションは成功するべき")
	stdoutOutput := checkStdout.String()
	assert.Contains(t, stdoutOutput, "設定に問題ありません", "成功メッセージが表示されるべき")
}

// TestConfigInit_WithConfigCheck_DummyValue はダミー値を含むconfigファイルをconfig checkで検証するテスト
func TestConfigInit_WithConfigCheck_DummyValue(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// ダミー値を含むconfigファイルを作成
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
	require.NoError(t, err)

	// config checkコマンドを実行
	checkCmd := makeConfigCheckCmd()
	cfgFile = configPath
	var checkStdout bytes.Buffer
	var checkStderr bytes.Buffer
	checkCmd.SetOut(&checkStdout)
	checkCmd.SetErr(&checkStderr)

	err = checkCmd.Execute()

	// ダミー値が含まれているため、エラーになるはず
	assert.Error(t, err, "ダミー値が含まれているため、バリデーションに失敗するべき")

	stderrOutput := checkStderr.String()
	assert.Contains(t, stderrOutput, "設定に以下の問題があります", "エラーメッセージが表示されるべき")
	assert.Contains(t, stderrOutput, "ダミー値", "ダミー値に関するエラーが含まれるべき")
}
