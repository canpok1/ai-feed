//go:build integration

package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/canpok1/ai-feed/internal/app"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProfileInitRunner_EdgeCases エッジケースのテスト
func TestProfileInitRunner_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, tmpDir string) string
		wantErr bool
	}{
		{
			name: "特殊文字を含むパス",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "profile-test_特殊文字.yml")
			},
			wantErr: false,
		},
		{
			name: "スペースを含むパス",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "profile with spaces.yml")
			},
			wantErr: false,
		},
		{
			name: "長いパス",
			setup: func(t *testing.T, tmpDir string) string {
				// 長いディレクトリ名を作成
				longName := strings.Repeat("a", 100)
				longDir := filepath.Join(tmpDir, longName)
				err := os.MkdirAll(longDir, 0755)
				assert.NoError(t, err)
				return filepath.Join(longDir, "profile.yml")
			},
			wantErr: false,
		},
		{
			name: "Unicode文字を含むパス",
			setup: func(t *testing.T, tmpDir string) string {
				unicodeName := "プロファイル_测试_тест_🚀.yml"
				return filepath.Join(tmpDir, unicodeName)
			},
			wantErr: false,
		},
		{
			name: "存在しないディレクトリ内のファイル",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "nonexistent", "dir", "profile.yml")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := tt.setup(t, tmpDir)

			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			stderr := &bytes.Buffer{}
			runner, runnerErr := app.NewProfileInitRunner(profileRepo, stderr)
			require.NoError(t, runnerErr)
			err := runner.Run()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// ファイルが作成されたことを確認
				_, statErr := os.Stat(filePath)
				assert.NoError(t, statErr, "Profile file should be created")

				// ファイル内容にUnicodeが含まれていても適切に処理されることを確認
				content, readErr := os.ReadFile(filePath)
				assert.NoError(t, readErr)

				// 日本語コメントが含まれていることを確認
				contentStr := string(content)
				assert.Contains(t, contentStr, "AI Feedのプロファイル設定ファイル")

				// ファイルが有効なUTF-8であることを確認
				assert.True(t, isValidUTF8(contentStr), "File content should be valid UTF-8")
			}
		})
	}
}

// TestProfileInitRunner_ConcurrentExecution 並行実行のテスト
func TestProfileInitRunner_ConcurrentExecution(t *testing.T) {
	tmpDir := t.TempDir()
	const goroutines = 10

	// 複数のgoroutineで同時にプロファイルファイル作成
	results := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			filePath := filepath.Join(tmpDir, fmt.Sprintf("profile_%d.yml", index))
			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			stderr := &bytes.Buffer{}
			runner, err := app.NewProfileInitRunner(profileRepo, stderr)
			if err != nil {
				results <- err
				return
			}
			results <- runner.Run()
		}(i)
	}

	// 全てのgoroutineの完了を待機
	successCount := 0
	for i := 0; i < goroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	// 全てのファイルが正常に作成されたことを確認
	assert.Equal(t, goroutines, successCount, "All profiles should be created successfully")

	// 作成されたファイルの数を確認
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, goroutines, "All profile files should exist")
}

// isValidUTF8 文字列が有効なUTF-8かどうかをチェック
func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == unicode.ReplacementChar {
			return false
		}
	}
	return true
}
