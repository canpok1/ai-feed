//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"gopkg.in/yaml.v3"
)

// GetProjectRoot はプロジェクトのルートディレクトリパスを取得する
// go.modファイルを探索してプロジェクトルートを特定する
func GetProjectRoot(t *testing.T) string {
	t.Helper()

	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリの取得に失敗しました: %v", err)
	}

	// go.mod ファイルを探索してプロジェクトルートを特定する
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				t.Fatalf("プロジェクトルートの絶対パス取得に失敗しました: %v", err)
			}
			return absPath
		}
		// ルートディレクトリに達した場合
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	t.Fatalf("プロジェクトルート(go.modファイル)が見つかりませんでした")
	return "" // 到達しない
}

// BuildBinary はTestMainでビルドされたバイナリのパスを返す
// この関数は下位互換性のために残されており、グローバル変数binaryPathを返す
func BuildBinary(t *testing.T) string {
	t.Helper()
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

// RecommendConfigParams はrecommendコマンドテスト用の設定パラメータを保持する構造体
type RecommendConfigParams struct {
	// FeedURLs はRSSフィードのURL一覧
	FeedURLs []string
	// GeminiAPIKey はGemini APIのキー
	GeminiAPIKey string
	// SlackWebhookURL はSlack WebhookのURL
	SlackWebhookURL string
	// MisskeyURL はMisskeyのURL
	MisskeyURL string
	// MisskeyToken はMisskeyのアクセストークン
	MisskeyToken string
}

// CreateRecommendTestConfig はrecommendコマンドのテスト用設定ファイルを作成する
// infra構造体を使用して型安全に設定を構築する
func CreateRecommendTestConfig(t *testing.T, tmpDir string, params RecommendConfigParams) string {
	t.Helper()

	configPath := filepath.Join(tmpDir, "config.yml")

	// テスト用の設定を型安全に構築
	// infra.Configとinfra.Profileを使用して構造を定義
	config := struct {
		DefaultProfile *struct {
			AI     *infra.AIConfig     `yaml:"ai,omitempty"`
			Output *infra.OutputConfig `yaml:"output,omitempty"`
			Feeds  []string            `yaml:"feeds,omitempty"` // テスト用のfeedsフィールド（実際のProfileには存在しない）
		} `yaml:"default_profile,omitempty"`
	}{
		DefaultProfile: &struct {
			AI     *infra.AIConfig     `yaml:"ai,omitempty"`
			Output *infra.OutputConfig `yaml:"output,omitempty"`
			Feeds  []string            `yaml:"feeds,omitempty"`
		}{
			// AI設定
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-2.5-flash",
					APIKey: params.GeminiAPIKey,
				},
			},
			// Feeds設定（テスト用）
			Feeds: params.FeedURLs,
		},
	}

	// Output設定を構築
	outputConfig := &infra.OutputConfig{}

	// Slack設定がある場合は追加
	if params.SlackWebhookURL != "" {
		// テストではWebhook URLをAPI Tokenとして扱う（モックサーバー用）
		enabled := true
		outputConfig.SlackAPI = &infra.SlackAPIConfig{
			Enabled:  &enabled,
			APIToken: params.SlackWebhookURL,
			Channel:  "#test-channel",
		}
	}

	// Misskey設定がある場合は追加
	if params.MisskeyURL != "" && params.MisskeyToken != "" {
		enabled := true
		outputConfig.Misskey = &infra.MisskeyConfig{
			Enabled:  &enabled,
			APIToken: params.MisskeyToken,
			APIURL:   params.MisskeyURL,
		}
	}

	config.DefaultProfile.Output = outputConfig

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
