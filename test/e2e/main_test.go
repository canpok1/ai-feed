//go:build e2e

package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	// binaryPath はテストスイート全体で共有されるバイナリパス
	binaryPath string
)

// TestMain はテストスイートの実行前に一度だけバイナリをビルドする
func TestMain(m *testing.M) {
	// プロジェクトルートを取得
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("プロジェクトルートの特定に失敗しました: %v", err)
	}

	// 一時ディレクトリにバイナリを作成
	tmpDir, err := os.MkdirTemp("", "ai-feed-e2e-")
	if err != nil {
		log.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}

	binaryPath = filepath.Join(tmpDir, "ai-feed")

	// go buildでバイナリをビルド
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("バイナリのビルドに失敗しました: %v\n出力: %s", err, string(output))
	}

	// テストを実行
	code := m.Run()

	// クリーンアップ
	os.RemoveAll(tmpDir)

	os.Exit(code)
}

// findProjectRoot はプロジェクトのルートディレクトリパスを取得する(TestMain用)
// *testing.Tに依存しないバージョン
func findProjectRoot() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
	}

	// go.mod ファイルを探索してプロジェクトルートを特定する
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("プロジェクトルートの絶対パス取得に失敗しました: %w", err)
			}
			return absPath, nil
		}
		// ルートディレクトリに達した場合
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	return "", fmt.Errorf("プロジェクトルート(go.modファイル)が見つかりませんでした")
}
