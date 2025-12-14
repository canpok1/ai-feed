//go:build integration

package app

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/canpok1/ai-feed/internal/app"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isRunningAsRoot はルート権限で実行されているかどうかを確認する
func isRunningAsRoot() bool {
	return os.Geteuid() == 0
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
			stderr := &bytes.Buffer{}
			runner, err := app.NewProfileInitRunner(yamlRepo, stderr)
			if err != nil {
				results <- err
				return
			}
			results <- runner.Run()
		}()
	}

	// 結果を収集
	successCount := 0
	var errs []error
	for i := 0; i < goroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			errs = append(errs, err)
		}
	}

	// 1つだけ成功、残りはファイル既存エラーであることを確認
	assert.Equal(t, 1, successCount, "Exactly one goroutine should succeed")
	require.Len(t, errs, goroutines-1, "Other goroutines should get file exists error")
	for _, err := range errs {
		assert.ErrorContains(t, err, "profile file already exists")
	}
}

// BenchmarkProfileInitRunner_Run ベンチマークテスト（統合）
func BenchmarkProfileInitRunner_Run(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		filePath := filepath.Join(tempDir, "bench_profile_"+strconv.Itoa(i)+".yml")
		yamlRepo := profile.NewYamlProfileRepositoryImpl(filePath)
		stderr := &bytes.Buffer{}
		runner, err := app.NewProfileInitRunner(yamlRepo, stderr)
		if err != nil {
			b.Fatalf("failed to create runner: %v", err)
		}
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
			stderr := &bytes.Buffer{}
			runner, err := app.NewProfileInitRunner(yamlRepo, stderr)
			if err != nil {
				b.Fatalf("failed to create runner: %v", err)
			}
			_ = runner.Run()
		}
	})
}

func TestProfileInitRunner_Run_WithRealRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setup            func(t *testing.T, tmpDir string) string
		wantErr          bool
		expectedErrorMsg string
		verify           func(t *testing.T, filePath string)
	}{
		{
			name: "正常系: 新規ファイル作成成功",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "test_profile.yml")
			},
			wantErr:          false,
			expectedErrorMsg: "",
			verify: func(t *testing.T, filePath string) {
				// ファイルが作成されたことを確認
				_, err := os.Stat(filePath)
				require.NoError(t, err, "プロファイルファイルが作成されているべき - これ以降の検証は無意味")

				// ファイルの内容を確認
				content, err := os.ReadFile(filePath)
				require.NoError(t, err, "ファイル読み込みに失敗した場合、以降の検証は無意味")
				assert.Contains(t, string(content), "# AI Feedのプロファイル設定ファイル", "テンプレートコメントが含まれているべき")
			},
		},
		{
			name: "異常系: 既存ファイルが存在する場合はエラー",
			setup: func(t *testing.T, tmpDir string) string {
				filePath := filepath.Join(tmpDir, "existing_profile.yml")
				err := os.WriteFile(filePath, []byte("existing content"), 0644)
				require.NoError(t, err)
				return filePath
			},
			wantErr:          true,
			expectedErrorMsg: "profile file already exists",
			verify: func(t *testing.T, filePath string) {
				// ファイルが変更されていないことを確認
				content, err := os.ReadFile(filePath)
				require.NoError(t, err, "ファイル読み込みに失敗した場合、以降の検証は無意味")
				assert.Contains(t, string(content), "existing content", "既存の内容は変更されないべき")
			},
		},
		{
			name: "異常系: ディレクトリが存在しない場合はエラー",
			setup: func(t *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "nonexistent", "profile.yml")
			},
			wantErr:          true,
			expectedErrorMsg: "no such file or directory",
			verify: func(t *testing.T, filePath string) {
				// ファイルが作成されていないことを確認
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.ErrorIs(t, err, fs.ErrNotExist)
			},
		},
		{
			name: "異常系: 書き込み権限がない場合はエラー",
			setup: func(t *testing.T, tmpDir string) string {
				// NOTE: ルート権限での実行時は権限制限が機能しないため、
				// このテストはスキップされる
				if isRunningAsRoot() {
					t.Skip("権限テストはルート権限では動作しないためスキップします")
				}

				// 読み取り専用ディレクトリを作成
				dir := filepath.Join(tmpDir, "readonly")
				err := os.MkdirAll(dir, 0755)
				require.NoError(t, err)

				// ディレクトリを読み取り専用に変更(書き込み不可)
				err = os.Chmod(dir, 0555)
				require.NoError(t, err)

				// NOTE: クリーンアップで権限を戻さないと、
				// t.TempDir()の自動削除が失敗する
				t.Cleanup(func() {
					os.Chmod(dir, 0755)
				})

				return filepath.Join(dir, "profile.yml")
			},
			wantErr:          true,
			expectedErrorMsg: "permission denied",
			verify: func(t *testing.T, filePath string) {
				// ファイルが作成されていないことを確認
				_, err := os.Stat(filePath)
				assert.Error(t, err)
				assert.ErrorIs(t, err, fs.ErrNotExist)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given: テスト環境のセットアップ
			tmpDir := t.TempDir()
			filePath := tt.setup(t, tmpDir)
			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			stderr := &bytes.Buffer{}
			runner, runnerErr := app.NewProfileInitRunner(profileRepo, stderr)
			require.NoError(t, runnerErr, "Runnerの作成に失敗してはいけない")

			// When: テスト対象の実行
			err := runner.Run()

			// Then: 結果の検証
			if tt.wantErr {
				assert.Error(t, err, "エラーが発生すべき")
				switch tt.name {
				case "異常系: ディレクトリが存在しない場合はエラー":
					assert.ErrorIs(t, err, fs.ErrNotExist)
				case "異常系: 書き込み権限がない場合はエラー":
					assert.ErrorIs(t, err, fs.ErrPermission)
				default:
					if tt.expectedErrorMsg != "" {
						assert.Contains(t, err.Error(), tt.expectedErrorMsg,
							"エラーメッセージに期待される文字列が含まれるべき")
					}
				}
			} else {
				assert.NoError(t, err, "エラーが発生してはいけない")
			}

			// Then: 追加の検証
			if tt.verify != nil {
				tt.verify(t, filePath)
			}
		})
	}
}
