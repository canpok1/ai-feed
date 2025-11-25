//go:build e2e

// Package common はe2eテスト用の共通ヘルパー関数とユーティリティを提供する
package common

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/test/e2e/common/mock"
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

// SetupPackage は各テストパッケージのTestMainから呼び出され、バイナリをビルドする
// この関数は同期されており、複数のパッケージから呼ばれても一度だけビルドを行う
func SetupPackage() {
	buildOnce.Do(func() {
		projectRoot, err := FindProjectRoot()
		if err != nil {
			buildErr = fmt.Errorf("プロジェクトルートの特定に失敗しました: %w", err)
			return
		}

		// 一時ディレクトリにバイナリを作成
		tmpDir, err := os.MkdirTemp("", "ai-feed-e2e-")
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

// CleanupPackage はテスト終了時のクリーンアップを行う
// 注: sync.Onceを使用しているため、実際のクリーンアップは最後に呼ばれた時のみ有効
func CleanupPackage() {
	if binaryPath != "" {
		os.RemoveAll(filepath.Dir(binaryPath))
	}
}

// FindProjectRoot はプロジェクトのルートディレクトリパスを取得する
// go.modファイルを探索してプロジェクトルートを特定する
func FindProjectRoot() (string, error) {
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

// GetProjectRoot はプロジェクトのルートディレクトリパスを取得する
// テスト用のヘルパー関数
func GetProjectRoot(t *testing.T) string {
	t.Helper()

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("プロジェクトルートの取得に失敗しました: %v", err)
	}
	return root
}

// GetBinaryPath はビルドされたバイナリのパスを返す
func GetBinaryPath() string {
	return binaryPath
}

// BuildBinary はビルドされたバイナリのパスを返す（下位互換性のため）
func BuildBinary(t *testing.T) string {
	t.Helper()
	return binaryPath
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
		DefaultProfile *infra.Profile `yaml:"default_profile,omitempty"`
	}{
		DefaultProfile: &infra.Profile{
			// AI設定
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-2.5-flash",
					APIKey: params.GeminiAPIKey,
				},
			},
			// プロンプト設定
			Prompt: &infra.PromptConfig{
				SystemPrompt:          "あなたはテスト用のアシスタントです。",
				CommentPromptTemplate: "以下の記事の紹介文を100字以内で作成してください。\n記事タイトル: {{TITLE}}\n記事URL: {{URL}}\n記事内容:\n{{CONTENT}}",
				SelectorPrompt:        "以下の記事一覧から、最も興味深い記事を1つ選択してください。",
			},
		},
	}

	// Output設定を構築
	outputConfig := &infra.OutputConfig{}

	// Slack設定がある場合は追加
	if params.SlackWebhookURL != "" {
		// テストではモックサーバーのURLをAPI URLとして設定
		// slack-goは base URL に /api/ が含まれることを期待しているため、末尾に /api/ を追加
		enabled := true
		slackTemplate := "{{if .Comment}}{{.Comment}}\n{{end}}<{{.Article.Link}}|{{.Article.Title}}>"
		apiURL := params.SlackWebhookURL + "/api/"
		outputConfig.SlackAPI = &infra.SlackAPIConfig{
			Enabled:         &enabled,
			APIToken:        "test-token", // モックサーバー用のダミートークン
			Channel:         "#test-channel",
			MessageTemplate: &slackTemplate,
			APIURL:          &apiURL, // モックサーバーのURL + /api/
		}
	}

	// Misskey設定がある場合は追加
	if params.MisskeyURL != "" && params.MisskeyToken != "" {
		enabled := true
		misskeyTemplate := "{{COMMENT}}\n[{{TITLE}}]({{URL}})"
		outputConfig.Misskey = &infra.MisskeyConfig{
			Enabled:         &enabled,
			APIToken:        params.MisskeyToken,
			APIURL:          params.MisskeyURL,
			MessageTemplate: &misskeyTemplate,
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

// RecommendTestEnv はrecommendコマンドテストの環境を保持する構造体
type RecommendTestEnv struct {
	TmpDir          string
	BinaryPath      string
	RSSServer       *httptest.Server
	SlackReceiver   *mock.MockSlackReceiver
	SlackServer     *httptest.Server
	MisskeyReceiver *mock.MockMisskeyReceiver
	MisskeyServer   *httptest.Server
}

// Cleanup はテスト環境のクリーンアップを実行する
func (e *RecommendTestEnv) Cleanup() {
	if e.RSSServer != nil {
		e.RSSServer.Close()
	}
	if e.SlackServer != nil {
		e.SlackServer.Close()
	}
	if e.MisskeyServer != nil {
		e.MisskeyServer.Close()
	}
}

// SetupRecommendTestOptions はセットアップのオプションを保持する構造体
type SetupRecommendTestOptions struct {
	// UseRSSServer はRSSモックサーバーを起動するかどうか
	UseRSSServer bool
	// RSSHandler はカスタムRSSハンドラ（nilの場合はデフォルトを使用）
	RSSHandler http.Handler
	// UseSlackServer はSlackモックサーバーを起動するかどうか
	UseSlackServer bool
	// UseMisskeyServer はMisskeyモックサーバーを起動するかどうか
	UseMisskeyServer bool
}

// SetupRecommendTest はrecommendコマンドテストの共通セットアップを実行する
// 必要なモックサーバーを起動し、テスト環境を構築する
func SetupRecommendTest(t *testing.T, opts SetupRecommendTestOptions) *RecommendTestEnv {
	t.Helper()

	env := &RecommendTestEnv{
		TmpDir:     t.TempDir(),
		BinaryPath: BuildBinary(t),
	}

	// RSSサーバーのセットアップ
	if opts.UseRSSServer {
		handler := opts.RSSHandler
		if handler == nil {
			handler = mock.NewMockRSSHandler()
		}
		env.RSSServer = httptest.NewServer(handler)
	}

	// Slackサーバーのセットアップ
	if opts.UseSlackServer {
		env.SlackReceiver = mock.NewMockSlackReceiver()
		env.SlackServer = httptest.NewServer(env.SlackReceiver)
	}

	// Misskeyサーバーのセットアップ
	if opts.UseMisskeyServer {
		env.MisskeyReceiver = mock.NewMockMisskeyReceiver()
		env.MisskeyServer = httptest.NewServer(env.MisskeyReceiver)
	}

	return env
}

// WaitForCondition は条件が満たされるまでポーリングで待機する
// タイムアウト時間内に条件が満たされればtrue、タイムアウトした場合はfalseを返す
func WaitForCondition(timeout time.Duration, condition func() bool) bool {
	timeoutCh := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCh:
			return false
		case <-ticker.C:
			if condition() {
				return true
			}
		}
	}
}

// SetupTestDataFile はテストデータファイルをtmpDirにコピーするヘルパー関数
func SetupTestDataFile(t *testing.T, projectRoot, testdataDir, fileName, dstFileName, tmpDir string) string {
	t.Helper()

	if fileName == "" {
		return ""
	}

	srcPath := filepath.Join(projectRoot, testdataDir, fileName)
	dstPath := filepath.Join(tmpDir, dstFileName)

	srcData, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("テストデータファイルの読み込みに失敗しました: %v (path: %s)", err, srcPath)
	}

	if err := os.WriteFile(dstPath, srcData, 0644); err != nil {
		t.Fatalf("テストデータファイルのコピーに失敗しました: %v", err)
	}

	return dstPath
}

// ChangeToTempDir は一時ディレクトリに移動し、テスト終了時に元のディレクトリに戻すヘルパー関数
func ChangeToTempDir(t *testing.T, tmpDir string) {
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
