//go:build e2e

package e2e

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/test/e2e/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecommendCommand_WithRealGeminiAPI は実際のGemini APIを使用してrecommendコマンドをテストする
func TestRecommendCommand_WithRealGeminiAPI(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:   true,
		UseSlackServer: true,
	})
	defer env.Cleanup()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:        []string{env.RSSServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: env.SlackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--url", env.RSSServer.URL)

	// コマンドが成功することを確認
	if !assert.NoError(t, err, "recommendコマンドは成功するはずです。出力: %s", output) {
		t.Logf("コマンド出力:\n%s", output)
		return
	}
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Slackにメッセージが送信されたことを確認
	if !waitForCondition(10*time.Second, env.SlackReceiver.ReceivedMessage) {
		t.Fatal("タイムアウト: Slackへのメッセージ送信が確認できませんでした")
	}

	// 受信したメッセージの確認
	messages := env.SlackReceiver.GetMessages()
	assert.Greater(t, len(messages), 0, "少なくとも1つのメッセージが送信されているはずです")

	lastMessage := env.SlackReceiver.GetLastMessage()
	assert.NotEmpty(t, lastMessage, "メッセージが空でないはずです")
}

// TestRecommendCommand_WithMisskey はMisskeyへの出力をテストする
func TestRecommendCommand_WithMisskey(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:     true,
		UseMisskeyServer: true,
	})
	defer env.Cleanup()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:     []string{env.RSSServer.URL},
		GeminiAPIKey: geminiAPIKey,
		MisskeyURL:   env.MisskeyServer.URL,
		MisskeyToken: "test-token",
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--url", env.RSSServer.URL)

	// コマンドが成功することを確認
	if !assert.NoError(t, err, "recommendコマンドは成功するはずです。出力: %s", output) {
		t.Logf("コマンド出力:\n%s", output)
		return
	}
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Misskeyにノートが送信されたことを確認
	if !waitForCondition(10*time.Second, env.MisskeyReceiver.ReceivedNote) {
		t.Fatal("タイムアウト: Misskeyへのノート送信が確認できませんでした")
	}

	// 受信したノートの確認
	notes := env.MisskeyReceiver.GetNotes()
	assert.Greater(t, len(notes), 0, "少なくとも1つのノートが送信されているはずです")

	lastNote := env.MisskeyReceiver.GetLastNote()
	assert.NotEmpty(t, lastNote, "ノートが空でないはずです")
}

// TestRecommendCommand_MultipleOutputs は複数出力先へのテストを実施する
func TestRecommendCommand_MultipleOutputs(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:     true,
		UseSlackServer:   true,
		UseMisskeyServer: true,
	})
	defer env.Cleanup()

	// テスト用の設定ファイルを作成（SlackとMisskey両方）
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:        []string{env.RSSServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: env.SlackServer.URL,
		MisskeyURL:      env.MisskeyServer.URL,
		MisskeyToken:    "test-token",
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--url", env.RSSServer.URL)

	// コマンドが成功することを確認
	if !assert.NoError(t, err, "recommendコマンドは成功するはずです。出力: %s", output) {
		t.Logf("コマンド出力:\n%s", output)
		return
	}
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Slackとミスキー両方にメッセージが送信されたことを確認
	slackReceived := waitForCondition(15*time.Second, env.SlackReceiver.ReceivedMessage)
	misskeyReceived := waitForCondition(15*time.Second, env.MisskeyReceiver.ReceivedNote)

	if !slackReceived {
		t.Error("タイムアウト: Slackへのメッセージ送信が確認できませんでした")
	}
	if !misskeyReceived {
		t.Error("タイムアウト: Misskeyへのノート送信が確認できませんでした")
	}
	if !slackReceived || !misskeyReceived {
		t.FailNow()
	}

	// 両方のメッセージが受信されていることを確認
	assert.True(t, env.SlackReceiver.ReceivedMessage(), "Slackにメッセージが送信されているはずです")
	assert.True(t, env.MisskeyReceiver.ReceivedNote(), "Misskeyにノートが送信されているはずです")
}

// TestRecommendCommand_EmptyFeed は空フィードの場合の動作をテストする
func TestRecommendCommand_EmptyFeed(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ（空フィードハンドラを使用）
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:   true,
		RSSHandler:     mock.NewMockEmptyFeedHandler(),
		UseSlackServer: true,
	})
	defer env.Cleanup()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:        []string{env.RSSServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: env.SlackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--url", env.RSSServer.URL)

	// 空のフィードの場合、エラーなく正常終了する
	require.NoError(t, err, "空フィードの場合、コマンドはエラーなく終了するはずです。出力: %s", output)
	assert.Contains(t, output, "記事が見つかりませんでした", "出力に記事がない旨のメッセージが含まれるはずです")

	// Slackにはメッセージが送信されないはず
	time.Sleep(2 * time.Second)
	assert.False(t, env.SlackReceiver.ReceivedMessage(), "空フィードの場合、Slackにメッセージは送信されないはずです")
}

// TestRecommendCommand_InvalidFeed は不正なフィードの場合の動作をテストする
func TestRecommendCommand_InvalidFeed(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ（不正なフィードハンドラを使用）
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:   true,
		RSSHandler:     mock.NewMockInvalidFeedHandler(),
		UseSlackServer: true,
	})
	defer env.Cleanup()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:        []string{env.RSSServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: env.SlackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--url", env.RSSServer.URL)

	// 不正なフィードの場合、エラーが発生するか、エラーメッセージが出力される
	if err != nil {
		// エラーの場合は、出力にパースエラーやXMLエラーがあることを確認
		outputLower := strings.ToLower(output)
		hasError := strings.Contains(outputLower, "error") ||
			strings.Contains(outputLower, "failed") ||
			strings.Contains(outputLower, "parse")
		assert.True(t, hasError, "エラーメッセージが含まれるはずです")
	}

	// Slackにはメッセージが送信されないはず
	time.Sleep(2 * time.Second)
	assert.False(t, env.SlackReceiver.ReceivedMessage(), "不正なフィードの場合、Slackにメッセージは送信されないはずです")
}

// TestRecommendCommand_WithProfile はプロファイルを使用したrecommendコマンドをテストする
func TestRecommendCommand_WithProfile(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// テスト環境をセットアップ
	env := SetupRecommendTest(t, SetupRecommendTestOptions{
		UseRSSServer:   true,
		UseSlackServer: true,
	})
	defer env.Cleanup()

	projectRoot := GetProjectRoot(t)

	// プロファイルファイルが存在するか確認
	// 存在しない場合はテストをスキップ（プロファイル機能のテストは別途実施されているため）
	profileTestDataPath := projectRoot + "/test/e2e/testdata/profiles/test_profile.yml"
	if _, err := os.Stat(profileTestDataPath); os.IsNotExist(err) {
		t.Skip("test_profile.ymlが存在しないためスキップします")
	}

	// プロファイルディレクトリを作成
	profilePath := setupTestDataFile(t, projectRoot, "profiles", "test_profile.yml", "test_profile.yml", env.TmpDir)
	require.NotEmpty(t, profilePath, "プロファイルファイルが作成されているはずです")

	// デフォルト設定ファイルを作成（プロファイルが優先される）
	// CreateRecommendTestConfigは失敗時にt.Fatalfで終了するため戻り値は無視
	_ = CreateRecommendTestConfig(t, env.TmpDir, RecommendConfigParams{
		FeedURLs:        []string{env.RSSServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: env.SlackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, env.TmpDir)

	// プロファイルを指定してrecommendコマンドを実行
	// 注: プロファイルファイルの内容によっては動作が変わるため、
	// 基本的な実行確認のみ行う
	output, err := ExecuteCommand(t, env.BinaryPath, "recommend", "--profile", "test_profile.yml")

	// プロファイル機能が正常に動作することを確認
	// エラーが発生した場合でも、プロファイルの読み込み自体は成功しているはず
	if err != nil {
		// エラーメッセージにプロファイル読み込みエラーが含まれていないことを確認
		assert.NotContains(t, strings.ToLower(output), "profile", "プロファイルの読み込みに失敗していないはずです")
	} else {
		assert.NotEmpty(t, output, "出力が空でないはずです")
	}
}
