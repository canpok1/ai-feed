package runner

import (
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/stretchr/testify/assert"
)

func TestProfileInitRunner_Run_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setupFile     func(string) error
		expectedError string
	}{
		{
			name: "保存成功",
			setupFile: func(filePath string) error {
				// ファイルが存在しない状態にする
				return nil
			},
			expectedError: "",
		},
		{
			name: "ファイル既存エラー",
			setupFile: func(filePath string) error {
				// 既にファイルが存在する状態を作る
				return os.WriteFile(filePath, []byte("existing content"), 0644)
			},
			expectedError: "profile file already exists",
		},
		{
			name: "書き込み権限エラー",
			setupFile: func(filePath string) error {
				// ディレクトリを読み取り専用にして書き込み権限エラーを発生させる
				dir := filepath.Dir(filePath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				if err := os.Chmod(dir, 0555); err != nil {
					return err
				}
				return nil
			},
			expectedError: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "test_profile.yml")

			// テスト用のファイル状態をセットアップ
			if err := tt.setupFile(filePath); err != nil {
				t.Fatalf("Failed to setup test file: %v", err)
			}

			// ProfileInitRunnerを作成
			yamlRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			runner := NewProfileInitRunner(yamlRepo)

			// 実行
			err := runner.Run()

			// 権限テストの後にディレクトリの権限を戻す
			if tt.name == "書き込み権限エラー" {
				dir := filepath.Dir(filePath)
				os.Chmod(dir, 0755) // 権限を戻す
			}

			// 検証
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				// ファイルが作成されていることを確認
				_, statErr := os.Stat(filePath)
				assert.NoError(t, statErr, "Profile file should be created")
			}
		})
	}
}

func TestProfileInitRunner_ConcurrentIntegration(t *testing.T) {
	const goroutines = 5
	tempDir := t.TempDir()

	// 複数のゴルーチンが同じファイルに対して並行してprofile initを実行
	// 1つだけが成功し、他はファイル既存エラーになるはず
	results := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			filePath := filepath.Join(tempDir, "concurrent_profile.yml")
			yamlRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			runner := NewProfileInitRunner(yamlRepo)
			results <- runner.Run()
		}()
	}

	// 結果を収集
	successCount := 0
	existsErrorCount := 0

	for i := 0; i < goroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else if assert.Contains(t, err.Error(), "profile file already exists") {
			existsErrorCount++
		}
	}

	// 1つだけ成功、残りはファイル既存エラーであることを確認
	assert.Equal(t, 1, successCount, "Exactly one goroutine should succeed")
	assert.Equal(t, goroutines-1, existsErrorCount, "Other goroutines should get file exists error")
}

// BenchmarkProfileInitRunner_Run ベンチマークテスト（統合）
func BenchmarkProfileInitRunner_Run(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		filePath := filepath.Join(tempDir, "bench_profile_"+strconv.Itoa(i)+".yml")
		yamlRepo := profile.NewYamlProfileRepositoryImpl(filePath)
		runner := NewProfileInitRunner(yamlRepo)
		b.StartTimer()

		_ = runner.Run()
	}
}

// BenchmarkProfileInitRunner_ConcurrentRun 並行実行ベンチマークテスト
func BenchmarkProfileInitRunner_ConcurrentRun(b *testing.B) {
	tempDir := b.TempDir()
	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := counter.Add(1)
			filePath := filepath.Join(tempDir, "concurrent_bench_"+strconv.FormatInt(id, 10)+".yml")
			yamlRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			runner := NewProfileInitRunner(yamlRepo)
			_ = runner.Run()
		}
	})
}
