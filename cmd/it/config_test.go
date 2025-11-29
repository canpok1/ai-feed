package it

import (
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
)

// TestConfigInit_WithConfigCheck_EnvBased はconfig initで生成したファイル（環境変数ベース）をconfig checkで検証するテスト
func TestConfigInit_WithConfigCheck_EnvBased(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// configファイルを生成（initコマンドのテンプレートを使用）
	configRepo := infra.NewYamlConfigRepository(configPath)
	err := configRepo.SaveWithTemplate()
	assert.NoError(t, err)

	// 環境変数を設定（有効なAPIキーをシミュレート）
	t.Setenv("GEMINI_API_KEY", "test-valid-gemini-key")

	// config checkコマンドを実行
	output, err := executeCommandInDir(t, tmpDir, "config", "check")

	// 環境変数から読み込む設定なので、バリデーションは成功するはず
	assert.NoError(t, err, "環境変数が設定されていれば、バリデーションは成功するべき")
	assert.Contains(t, output, "設定に問題ありません", "成功メッセージが表示されるべき")
}

// TestConfigInit_WithConfigCheck_DummyValue はダミー値を含むconfigファイルをconfig checkで検証するテスト
func TestConfigInit_WithConfigCheck_DummyValue(t *testing.T) {
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
	assert.NoError(t, err)

	// config checkコマンドを実行
	output, err := executeCommandInDir(t, tmpDir, "config", "check")

	// ダミー値が含まれているため、エラーになるはず
	assert.Error(t, err, "ダミー値が含まれているため、バリデーションに失敗するべき")
	assert.Contains(t, output, "設定に以下の問題があります", "エラーメッセージが表示されるべき")
	assert.Contains(t, output, "ダミー値", "ダミー値に関するエラーが含まれるべき")
}
