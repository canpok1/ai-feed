//go:build integration

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// テスト用の一時ディレクトリ
	tmpDir := t.TempDir()

	t.Run("profile init → check の一連の流れ", func(t *testing.T) {
		// 1. profile init でファイル作成
		profilePath := filepath.Join(tmpDir, "test_profile.yml")
		initCmd := makeProfileInitCmd()
		initCmd.SetArgs([]string{profilePath})

		var initOut bytes.Buffer
		var initErr bytes.Buffer
		initCmd.SetOut(&initOut)
		initCmd.SetErr(&initErr)

		err := initCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, initOut.String(), "プロファイルファイルが正常に作成されました")

		// ファイルが実際に作成されていることを確認
		_, statErr := os.Stat(profilePath)
		require.NoError(t, statErr)

		// 2. 作成されたファイルを編集して有効なプロファイルにする
		content, err := os.ReadFile(profilePath)
		require.NoError(t, err)

		// 必要な設定を追加
		modifiedContent := updateProfileContent(string(content))
		err = os.WriteFile(profilePath, []byte(modifiedContent), 0644)
		require.NoError(t, err)

		// 3. profile check で検証
		checkCmd := makeProfileCheckCmd()
		checkCmd.SetArgs([]string{profilePath})

		var checkOut bytes.Buffer
		var checkErr bytes.Buffer
		checkCmd.SetOut(&checkOut)
		checkCmd.SetErr(&checkErr)

		err = checkCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, checkOut.String(), "プロファイルの検証が完了しました")
	})

	t.Run("config.ymlとprofile.ymlのマージテスト", func(t *testing.T) {
		// 作業ディレクトリを一時的に変更
		originalWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalWd)

		// config.ymlを作成（部分的な設定）
		configContent := `default_profile:
  ai:
    gemini:
      type: "gemini-1.5-flash"
      api_key: "config-api-key"
  system_prompt: "デフォルトシステムプロンプト"
  output:
    slack_api:
      api_token: "xoxb-config-token"
      channel: "#default"
      message_template: "{{COMMENT}} {{URL}}"
`
		err := os.WriteFile("./config.yml", []byte(configContent), 0644)
		require.NoError(t, err)

		// profile.ymlを作成（追加・上書き設定）
		profileContent := `system_prompt: "カスタムシステムプロンプト"
comment_prompt_template: "カスタムテンプレート {{TITLE}}"
output:
  misskey:
    api_token: "custom-misskey-token"
    api_url: "https://custom.misskey.social/api"
    message_template: "{{COMMENT}} {{URL}}"
`
		profilePath := "./test_merge_profile.yml"
		err = os.WriteFile(profilePath, []byte(profileContent), 0644)
		require.NoError(t, err)

		// profile check でマージされたプロファイルを検証
		checkCmd := makeProfileCheckCmd()
		checkCmd.SetArgs([]string{profilePath})

		var checkOut bytes.Buffer
		checkCmd.SetOut(&checkOut)

		err = checkCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, checkOut.String(), "プロファイルの検証が完了しました")
	})

	t.Run("エラーケースの統合テスト", func(t *testing.T) {
		scenarios := []struct {
			name        string
			setupFile   func() string
			expectedErr string
		}{
			{
				name: "存在しないファイル",
				setupFile: func() string {
					return filepath.Join(tmpDir, "nonexistent.yml")
				},
				expectedErr: "プロファイルファイルが見つかりません",
			},
			{
				name: "不正なYAML",
				setupFile: func() string {
					invalidPath := filepath.Join(tmpDir, "invalid.yml")
					invalidContent := `
invalid yaml content:
  - no proper
    broken: yaml syntax
`
					err := os.WriteFile(invalidPath, []byte(invalidContent), 0644)
					require.NoError(t, err)
					return invalidPath
				},
				expectedErr: "プロファイルの読み込みに失敗しました",
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				filePath := scenario.setupFile()

				checkCmd := makeProfileCheckCmd()
				checkCmd.SetArgs([]string{filePath})

				var checkOut bytes.Buffer
				var checkErr bytes.Buffer
				checkCmd.SetOut(&checkOut)
				checkCmd.SetErr(&checkErr)

				err := checkCmd.Execute()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), scenario.expectedErr)
			})
		}
	})
}

// updateProfileContent テンプレートファイルの内容を有効なプロファイルに更新する
func updateProfileContent(templateContent string) string {
	// テンプレートコメントを実際の設定に置き換える
	lines := strings.Split(templateContent, "\n")
	var result []string

	inExampleSection := false
	for _, line := range lines {
		if strings.Contains(line, "# 以下は設定例です") {
			inExampleSection = true
			result = append(result, line)
			// 有効な設定を追加
			result = append(result, "ai:")
			result = append(result, "  gemini:")
			result = append(result, "    type: \"gemini-1.5-flash\"")
			result = append(result, "    api_key: \"test-api-key\"")
			result = append(result, "system_prompt: \"テスト用システムプロンプト\"")
			result = append(result, "comment_prompt_template: \"テスト用テンプレート {{TITLE}}\"")
			result = append(result, "output:")
			result = append(result, "  slack_api:")
			result = append(result, "    api_token: \"xoxb-test-token\"")
			result = append(result, "    channel: \"#test\"")
			result = append(result, "    message_template: \"{{COMMENT}} {{URL}}\"")
			continue
		}

		if inExampleSection && strings.HasPrefix(line, "#") {
			continue // コメント化された例を削除
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func TestProfileCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test")
	}

	tmpDir := t.TempDir()

	t.Run("複数ファイルの並行処理", func(t *testing.T) {
		const fileCount = 10

		// 複数のプロファイルファイルを作成
		for i := 0; i < fileCount; i++ {
			profilePath := filepath.Join(tmpDir, fmt.Sprintf("profile_%d.yml", i))

			initCmd := makeProfileInitCmd()
			initCmd.SetArgs([]string{profilePath})

			var initOut bytes.Buffer
			initCmd.SetOut(&initOut)

			err := initCmd.Execute()
			require.NoError(t, err)

			// ファイルを有効なプロファイルに更新
			content, err := os.ReadFile(profilePath)
			require.NoError(t, err)

			modifiedContent := updateProfileContent(string(content))
			err = os.WriteFile(profilePath, []byte(modifiedContent), 0644)
			require.NoError(t, err)
		}

		// 並行して検証実行
		results := make(chan error, fileCount)
		for i := 0; i < fileCount; i++ {
			go func(index int) {
				profilePath := filepath.Join(tmpDir, fmt.Sprintf("profile_%d.yml", index))

				checkCmd := makeProfileCheckCmd()
				checkCmd.SetArgs([]string{profilePath})

				var checkOut bytes.Buffer
				checkCmd.SetOut(&checkOut)

				results <- checkCmd.Execute()
			}(i)
		}

		// 全ての処理が完了することを確認
		successCount := 0
		for i := 0; i < fileCount; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		assert.Equal(t, fileCount, successCount, "All profile validations should succeed")
	})
}
