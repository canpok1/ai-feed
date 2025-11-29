package it

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

var (
	// binaryPath はテストスイート全体で共有されるバイナリパス
	binaryPath string
	// buildOnce はバイナリのビルドを一度だけ行うための同期制御
	buildOnce sync.Once
	// buildErr はビルド中に発生したエラーを保持する
	buildErr error
)

// setupPackage は各テストパッケージのTestMainから呼び出され、バイナリをビルドする
func setupPackage() {
	buildOnce.Do(func() {
		projectRoot, err := findProjectRoot()
		if err != nil {
			buildErr = fmt.Errorf("プロジェクトルートの特定に失敗しました: %w", err)
			return
		}

		// 一時ディレクトリにバイナリを作成
		tmpDir, err := os.MkdirTemp("", "ai-feed-it-")
		if err != nil {
			buildErr = fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
			return
		}

		binaryPath = filepath.Join(tmpDir, "ai-feed")

		// go buildでバイナリをビルド
		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("バイナリのビルドに失敗しました: %w\n出力: %s", err, string(output))
			return
		}
	})

	if buildErr != nil {
		log.Fatal(buildErr)
	}
}

// cleanupPackage はテスト終了時のクリーンアップを行う
func cleanupPackage() {
	if binaryPath != "" {
		os.RemoveAll(filepath.Dir(binaryPath))
	}
}

// findProjectRoot はプロジェクトのルートディレクトリパスを取得する
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
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	return "", fmt.Errorf("プロジェクトルート(go.modファイル)が見つかりませんでした")
}

// getProjectRoot はプロジェクトのルートディレクトリパスを取得する
func getProjectRoot(t *testing.T) string {
	t.Helper()

	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("プロジェクトルートの取得に失敗しました: %v", err)
	}
	return root
}

// getBinaryPath はビルドされたバイナリのパスを返す
func getBinaryPath() string {
	return binaryPath
}

// executeCommand はバイナリを実行し、標準出力と標準エラー出力を結合して返す
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()

	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return outputStr, fmt.Errorf("コマンド実行に失敗しました: %w", err)
	}

	return outputStr, nil
}

// executeCommandInDir は指定されたディレクトリでバイナリを実行する
func executeCommandInDir(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return outputStr, fmt.Errorf("コマンド実行に失敗しました: %w", err)
	}

	return outputStr, nil
}

// changeToTempDir は一時ディレクトリに移動し、テスト終了時に元のディレクトリに戻す
func changeToTempDir(t *testing.T, tmpDir string) {
	t.Helper()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリの取得に失敗しました: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("一時ディレクトリへの移動に失敗しました: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("元のディレクトリへの復帰に失敗しました: %v", err)
		}
	})
}

// createConfigFile はテスト用のconfig.ymlファイルを作成する
func createConfigFile(t *testing.T, dir string, content map[string]interface{}) string {
	t.Helper()

	configPath := filepath.Join(dir, "config.yml")

	data, err := yaml.Marshal(content)
	if err != nil {
		t.Fatalf("設定データのYAMLマーシャルに失敗しました: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("設定ファイルの書き込みに失敗しました: %v", err)
	}

	return configPath
}

// createProfileFile はテスト用のprofile.ymlファイルを作成する
func createProfileFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("プロファイルファイルの書き込みに失敗しました: %v", err)
	}
}
