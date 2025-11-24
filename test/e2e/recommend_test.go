//go:build e2e

package e2e

import (
	"net/http/httptest"
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

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// モックRSSサーバーを起動
	rssServer := httptest.NewServer(mock.NewMockRSSHandler())
	defer rssServer.Close()

	// モックSlackサーバーを起動
	slackReceiver := mock.NewMockSlackReceiver()
	slackServer := httptest.NewServer(slackReceiver)
	defer slackServer.Close()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:        []string{rssServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: slackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, binaryPath, "recommend")

	// コマンドが成功することを確認
	assert.NoError(t, err, "recommendコマンドは成功するはずです")
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Slackにメッセージが送信されたことを確認（タイムアウト付き）
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	messageReceived := false
	for !messageReceived {
		select {
		case <-timeout:
			t.Fatal("タイムアウト: Slackへのメッセージ送信が確認できませんでした")
		case <-ticker.C:
			if slackReceiver.ReceivedMessage() {
				messageReceived = true
			}
		}
	}

	// 受信したメッセージの確認
	messages := slackReceiver.GetMessages()
	assert.Greater(t, len(messages), 0, "少なくとも1つのメッセージが送信されているはずです")

	lastMessage := slackReceiver.GetLastMessage()
	assert.NotEmpty(t, lastMessage, "メッセージが空でないはずです")
}

// TestRecommendCommand_WithMisskey はMisskeyへの出力をテストする
func TestRecommendCommand_WithMisskey(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// モックRSSサーバーを起動
	rssServer := httptest.NewServer(mock.NewMockRSSHandler())
	defer rssServer.Close()

	// モックMisskeyサーバーを起動
	misskeyReceiver := mock.NewMockMisskeyReceiver()
	misskeyServer := httptest.NewServer(misskeyReceiver)
	defer misskeyServer.Close()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:     []string{rssServer.URL},
		GeminiAPIKey: geminiAPIKey,
		MisskeyURL:   misskeyServer.URL,
		MisskeyToken: "test-token",
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, binaryPath, "recommend")

	// コマンドが成功することを確認
	assert.NoError(t, err, "recommendコマンドは成功するはずです")
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Misskeyにノートが送信されたことを確認（タイムアウト付き）
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	noteReceived := false
	for !noteReceived {
		select {
		case <-timeout:
			t.Fatal("タイムアウト: Misskeyへのノート送信が確認できませんでした")
		case <-ticker.C:
			if misskeyReceiver.ReceivedNote() {
				noteReceived = true
			}
		}
	}

	// 受信したノートの確認
	notes := misskeyReceiver.GetNotes()
	assert.Greater(t, len(notes), 0, "少なくとも1つのノートが送信されているはずです")

	lastNote := misskeyReceiver.GetLastNote()
	assert.NotEmpty(t, lastNote, "ノートが空でないはずです")
}

// TestRecommendCommand_MultipleOutputs は複数出力先へのテストを実施する
func TestRecommendCommand_MultipleOutputs(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// モックRSSサーバーを起動
	rssServer := httptest.NewServer(mock.NewMockRSSHandler())
	defer rssServer.Close()

	// モックSlackサーバーを起動
	slackReceiver := mock.NewMockSlackReceiver()
	slackServer := httptest.NewServer(slackReceiver)
	defer slackServer.Close()

	// モックMisskeyサーバーを起動
	misskeyReceiver := mock.NewMockMisskeyReceiver()
	misskeyServer := httptest.NewServer(misskeyReceiver)
	defer misskeyServer.Close()

	// テスト用の設定ファイルを作成（SlackとMisskey両方）
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:        []string{rssServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: slackServer.URL,
		MisskeyURL:      misskeyServer.URL,
		MisskeyToken:    "test-token",
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, binaryPath, "recommend")

	// コマンドが成功することを確認
	assert.NoError(t, err, "recommendコマンドは成功するはずです")
	assert.NotEmpty(t, output, "出力が空でないはずです")

	// Slackとミスキー両方にメッセージが送信されたことを確認
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	slackReceived := false
	misskeyReceived := false
	for !slackReceived || !misskeyReceived {
		select {
		case <-timeout:
			if !slackReceived {
				t.Error("タイムアウト: Slackへのメッセージ送信が確認できませんでした")
			}
			if !misskeyReceived {
				t.Error("タイムアウト: Misskeyへのノート送信が確認できませんでした")
			}
			t.FailNow()
		case <-ticker.C:
			if !slackReceived && slackReceiver.ReceivedMessage() {
				slackReceived = true
			}
			if !misskeyReceived && misskeyReceiver.ReceivedNote() {
				misskeyReceived = true
			}
		}
	}

	// 両方のメッセージが受信されていることを確認
	assert.True(t, slackReceiver.ReceivedMessage(), "Slackにメッセージが送信されているはずです")
	assert.True(t, misskeyReceiver.ReceivedNote(), "Misskeyにノートが送信されているはずです")
}

// TestRecommendCommand_EmptyFeed は空フィードの場合の動作をテストする
func TestRecommendCommand_EmptyFeed(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// モック空フィードサーバーを起動
	emptyFeedServer := httptest.NewServer(mock.NewMockEmptyFeedHandler())
	defer emptyFeedServer.Close()

	// モックSlackサーバーを起動
	slackReceiver := mock.NewMockSlackReceiver()
	slackServer := httptest.NewServer(slackReceiver)
	defer slackServer.Close()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:        []string{emptyFeedServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: slackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, binaryPath, "recommend")

	// 空のフィードの場合、エラーなく正常終了する
	require.NoError(t, err, "空フィードの場合、コマンドはエラーなく終了するはずです")
	assert.Contains(t, output, "記事が見つかりませんでした", "出力に記事がない旨のメッセージが含まれるはずです")

	// Slackにはメッセージが送信されないはず
	time.Sleep(2 * time.Second)
	assert.False(t, slackReceiver.ReceivedMessage(), "空フィードの場合、Slackにメッセージは送信されないはずです")
}

// TestRecommendCommand_InvalidFeed は不正なフィードの場合の動作をテストする
func TestRecommendCommand_InvalidFeed(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// モック不正フィードサーバーを起動
	invalidFeedServer := httptest.NewServer(mock.NewMockInvalidFeedHandler())
	defer invalidFeedServer.Close()

	// モックSlackサーバーを起動
	slackReceiver := mock.NewMockSlackReceiver()
	slackServer := httptest.NewServer(slackReceiver)
	defer slackServer.Close()

	// テスト用の設定ファイルを作成
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:        []string{invalidFeedServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: slackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// recommendコマンドを実行
	output, err := ExecuteCommand(t, binaryPath, "recommend")

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
	assert.False(t, slackReceiver.ReceivedMessage(), "不正なフィードの場合、Slackにメッセージは送信されないはずです")
}

// TestRecommendCommand_WithProfile はプロファイルを使用したrecommendコマンドをテストする
func TestRecommendCommand_WithProfile(t *testing.T) {
	// Gemini APIキーの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
	}

	// バイナリをビルド
	binaryPath := BuildBinary(t)

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	projectRoot := GetProjectRoot(t)

	// プロファイルファイルが存在するか確認
	// 存在しない場合はテストをスキップ（プロファイル機能のテストは別途実施されているため）
	profileTestDataPath := projectRoot + "/test/e2e/testdata/profiles/test_profile.yml"
	if _, err := os.Stat(profileTestDataPath); os.IsNotExist(err) {
		t.Skip("test_profile.ymlが存在しないためスキップします")
	}

	// プロファイルディレクトリを作成
	profilePath := setupTestDataFile(t, projectRoot, "profiles", "test_profile.yml", "test_profile.yml", tmpDir)
	require.NotEmpty(t, profilePath, "プロファイルファイルが作成されているはずです")

	// モックRSSサーバーを起動
	rssServer := httptest.NewServer(mock.NewMockRSSHandler())
	defer rssServer.Close()

	// モックSlackサーバーを起動
	slackReceiver := mock.NewMockSlackReceiver()
	slackServer := httptest.NewServer(slackReceiver)
	defer slackServer.Close()

	// デフォルト設定ファイルを作成（プロファイルが優先される）
	// CreateRecommendTestConfigは失敗時にt.Fatalfで終了するため戻り値は無視
	_ = CreateRecommendTestConfig(t, tmpDir, RecommendConfigParams{
		FeedURLs:        []string{rssServer.URL},
		GeminiAPIKey:    geminiAPIKey,
		SlackWebhookURL: slackServer.URL,
	})

	// 一時ディレクトリに移動
	changeToTempDir(t, tmpDir)

	// プロファイルを指定してrecommendコマンドを実行
	// 注: プロファイルファイルの内容によっては動作が変わるため、
	// 基本的な実行確認のみ行う
	output, err := ExecuteCommand(t, binaryPath, "recommend", "--profile", "test_profile.yml")

	// プロファイル機能が正常に動作することを確認
	// エラーが発生した場合でも、プロファイルの読み込み自体は成功しているはず
	if err != nil {
		// エラーメッセージにプロファイル読み込みエラーが含まれていないことを確認
		assert.NotContains(t, strings.ToLower(output), "profile", "プロファイルの読み込みに失敗していないはずです")
	} else {
		assert.NotEmpty(t, output, "出力が空でないはずです")
	}
}
