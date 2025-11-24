//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// GetProjectRoot はプロジェクトのルートディレクトリパスを取得する
func GetProjectRoot(t *testing.T) string {
	t.Helper()

	// テストファイルの位置からプロジェクトルートを推定
	// test/e2e からプロジェクトルートへ
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリの取得に失敗しました: %v", err)
	}

	// test/e2e から ../../ でルートへ
	projectRoot := filepath.Join(currentDir, "..", "..")
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		t.Fatalf("プロジェクトルートの絶対パス取得に失敗しました: %v", err)
	}

	return absPath
}

// BuildBinary はテスト用のバイナリをビルドし、ビルドしたバイナリのパスを返す
// バイナリは一時ディレクトリに配置される
func BuildBinary(t *testing.T) string {
	t.Helper()

	projectRoot := GetProjectRoot(t)

	// 一時ディレクトリにバイナリを作成
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "ai-feed")

	// go buildでバイナリをビルド
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("バイナリのビルドに失敗しました: %v\n出力: %s", err, string(output))
	}

	return binaryPath
}

// TestConfigParams はテスト用設定ファイルのパラメータを保持する構造体
type TestConfigParams struct {
	// DefaultProfile はデフォルトプロファイルの設定
	DefaultProfile map[string]interface{}
	// Cache はキャッシュの設定
	Cache map[string]interface{}
}

// CreateTestConfig はテスト用の設定ファイルを作成し、作成したファイルのパスを返す
func CreateTestConfig(t *testing.T, tmpDir string, params TestConfigParams) string {
	t.Helper()

	configPath := filepath.Join(tmpDir, "config.yml")

	// 設定データを構築
	config := make(map[string]interface{})

	if params.DefaultProfile != nil {
		config["default_profile"] = params.DefaultProfile
	}
	if params.Cache != nil {
		config["cache"] = params.Cache
	}

	// YAMLにマーシャル
	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("設定データのYAMLマーシャルに失敗しました: %v", err)
	}

	// ファイルに書き込み
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("設定ファイルの書き込みに失敗しました: %v", err)
	}

	return configPath
}

// ExecuteCommand はバイナリを実行し、標準出力と標準エラー出力を結合して返す
func ExecuteCommand(t *testing.T, binaryPath string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()

	// 標準出力と標準エラー出力を結合した出力を文字列として返す
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		// エラーが発生した場合も出力とともにエラーを返す
		return outputStr, fmt.Errorf("コマンド実行に失敗しました: %w", err)
	}

	return outputStr, nil
}
