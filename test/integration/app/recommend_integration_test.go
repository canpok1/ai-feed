//go:build integration

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/canpok1/ai-feed/internal/app"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/cache"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra/fetch"
	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRSSHandler は標準的なRSSフィードを返すモックハンドラを生成する
func mockRSSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS Feed</title>
    <link>https://example.com</link>
    <description>Test RSS Feed for Integration Testing</description>
    <item>
      <title>Test Article 1</title>
      <link>https://example.com/article1</link>
      <description>This is test article 1</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Test Article 2</title>
      <link>https://example.com/article2</link>
      <description>This is test article 2</description>
      <pubDate>Tue, 02 Jan 2024 00:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`))
	})
}

// mockEmptyFeedHandler は記事が0件の空フィードを返すモックハンドラを生成する
func mockEmptyFeedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>Empty Feed for Testing</description>
  </channel>
</rss>`))
	})
}

// mockErrorFeedHandler はHTTPエラーを返すモックハンドラを生成する
func mockErrorFeedHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	})
}

// mockRecommender はテスト用のRecommender実装
type mockRecommender struct {
	returnComment string
}

func newMockRecommender(comment string) *mockRecommender {
	return &mockRecommender{returnComment: comment}
}

func (r *mockRecommender) Recommend(ctx context.Context, articles []entity.Article) (*entity.Recommend, error) {
	if len(articles) == 0 {
		return nil, nil
	}
	comment := r.returnComment
	return &entity.Recommend{
		Article: articles[0],
		Comment: &comment,
	}, nil
}

// mockMessageSender はテスト用のMessageSender実装
type mockMessageSender struct {
	mu              sync.Mutex
	receivedCount   int
	receivedMessage string
	serviceName     string
	shouldError     bool
}

func newMockMessageSender(serviceName string) *mockMessageSender {
	return &mockMessageSender{serviceName: serviceName}
}

func (s *mockMessageSender) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receivedCount++
	if recommend.Comment != nil {
		s.receivedMessage = *recommend.Comment
	}
	if s.shouldError {
		return assert.AnError
	}
	return nil
}

func (s *mockMessageSender) ServiceName() string {
	return s.serviceName
}

func (s *mockMessageSender) GetReceivedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receivedCount
}

// testSenderFactory はテスト用のMessageSenderファクトリを返す
func testSenderFactory(senders []domain.MessageSender) app.MessageSenderFactory {
	return func(outputConfig *entity.OutputConfig) ([]domain.MessageSender, error) {
		return senders, nil
	}
}

// testCacheFactory はテスト用のRecommendCacheファクトリを返す
func testCacheFactory(c domain.RecommendCache) app.RecommendCacheFactory {
	return func(cacheConfig *entity.CacheConfig) (domain.RecommendCache, error) {
		return c, nil
	}
}

func TestRecommendRunner_Integration_HappyPath(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("This is a test comment")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	// NopCacheを使用（キャッシュなし）
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender.GetReceivedCount(), "メッセージが1回送信されるべき")
	assert.Contains(t, stdoutBuffer.String(), "記事URL:", "記事URLが出力に含まれるべき")
}

func TestRecommendRunner_Integration_WithRealFetcher(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Integration test comment")

	// 複数のMessageSenderを使用（Slack + Misskey模倣）
	slackSender := newMockMessageSender("Slack")
	misskeySender := newMockMessageSender("Misskey")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{slackSender, misskeySender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{
		Prompt: &entity.PromptConfig{
			FixedMessage: "Fixed test message",
		},
	}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender.GetReceivedCount(), "Slackにメッセージが送信されるべき")
	assert.Equal(t, 1, misskeySender.GetReceivedCount(), "Misskeyにメッセージが送信されるべき")
}

func TestRecommendRunner_Integration_EmptyFeed(t *testing.T) {
	t.Parallel()

	// 空のRSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockEmptyFeedHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Should not be called")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証: 空のフィードの場合はErrNoArticlesFoundが返される
	assert.ErrorIs(t, err, app.ErrNoArticlesFound)
	assert.Equal(t, 0, slackSender.GetReceivedCount(), "空のフィードではメッセージが送信されないべき")
}

func TestRecommendRunner_Integration_FeedFetchError(t *testing.T) {
	t.Parallel()

	// HTTPエラーを返すモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockErrorFeedHandler(http.StatusInternalServerError))
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Should not be called")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証: フェッチエラーの場合はErrNoArticlesFoundが返される
	assert.ErrorIs(t, err, app.ErrNoArticlesFound)
	assert.Equal(t, 0, slackSender.GetReceivedCount(), "フェッチエラー時はメッセージが送信されないべき")
}

func TestRecommendRunner_Integration_WithFileCache(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 一時ディレクトリにキャッシュファイルを作成
	tmpDir := t.TempDir()
	cacheFilePath := filepath.Join(tmpDir, "test_cache.jsonl")

	cacheConfig := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      cacheFilePath,
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// 実際のFileRecommendCacheを使用
	fileCache := cache.NewFileRecommendCache(cacheConfig)

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Cached test comment")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		cacheConfig,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(fileCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 1回目の実行
	err = runner.Run(context.Background(), params, profile)
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender.GetReceivedCount(), "1回目: メッセージが送信されるべき")

	// 2回目の実行: 同じ記事がキャッシュされているため、別の記事が選択されるか、全てキャッシュ済みならスキップされる
	// 新しいFileRecommendCacheを作成して2回目の実行をシミュレート
	fileCache2 := cache.NewFileRecommendCache(cacheConfig)
	slackSender2 := newMockMessageSender("Slack")

	stderrBuffer2 := new(bytes.Buffer)
	stdoutBuffer2 := new(bytes.Buffer)

	runner2, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer2,
		stdoutBuffer2,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		cacheConfig,
		testSenderFactory([]domain.MessageSender{slackSender2}),
		testCacheFactory(fileCache2),
	)
	require.NoError(t, err)

	err = runner2.Run(context.Background(), params, profile)
	// 2回目の実行では、1件目の記事がキャッシュ済みなので2件目が選択される、
	// または全てキャッシュ済みなら正常終了（エラーなし）
	// テストフィードには2記事あるため、2件目が選択される可能性がある
	if err == nil {
		// 2件目の記事が選択された場合
		assert.Equal(t, 1, slackSender2.GetReceivedCount(), "2回目: 別の記事でメッセージが送信されるべき")
	} else {
		// 全てキャッシュ済みの場合（ただしこのテストでは2件目があるのでこのケースは発生しない）
		t.Logf("2回目の実行でエラー: %v", err)
	}
}

func TestRecommendRunner_Integration_ConcurrentMultiSender(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Concurrent test comment")

	// 複数のMessageSenderを使用して並行送信をテスト
	sender1 := newMockMessageSender("Service1")
	sender2 := newMockMessageSender("Service2")
	sender3 := newMockMessageSender("Service3")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{sender1, sender2, sender3}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証: 全てのsenderにメッセージが送信されるべき
	assert.NoError(t, err)
	assert.Equal(t, 1, sender1.GetReceivedCount(), "Service1にメッセージが送信されるべき")
	assert.Equal(t, 1, sender2.GetReceivedCount(), "Service2にメッセージが送信されるべき")
	assert.Equal(t, 1, sender3.GetReceivedCount(), "Service3にメッセージが送信されるべき")
}

func TestRecommendRunner_Integration_SenderError(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Error test comment")

	// エラーを返すMessageSenderを使用
	errorSender := newMockMessageSender("ErrorService")
	errorSender.shouldError = true

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{errorSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証: senderエラーが発生した場合はエラーが返される
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send")
	assert.Equal(t, 1, errorSender.GetReceivedCount(), "エラーがあってもsenderは呼ばれるべき")
}

func TestRecommendRunner_Integration_MultipleFeedsWithRetry(t *testing.T) {
	t.Parallel()

	// 1つ目は失敗、2つ目は成功するフィードをセットアップ
	failingServer := httptest.NewServer(mockErrorFeedHandler(http.StatusInternalServerError))
	defer failingServer.Close()

	successServer := httptest.NewServer(mockRSSHandler())
	defer successServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Retry test comment")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	// 両方のフィードURLを渡す（ランダム選択のためどちらが先に選ばれるかは不確定）
	params := &app.RecommendParams{
		URLs: []string{failingServer.URL, successServer.URL},
	}
	profile := &entity.Profile{}

	// 実行: 失敗したフィードは除外され、成功したフィードから記事が取得される
	err = runner.Run(context.Background(), params, profile)

	// 検証: 少なくとも1つのフィードが成功すれば処理は続行される
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender.GetReceivedCount(), "成功したフィードからメッセージが送信されるべき")
}

func TestRecommendRunner_Integration_AllCachedArticles(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 一時ディレクトリにキャッシュファイルを作成
	tmpDir := t.TempDir()
	cacheFilePath := filepath.Join(tmpDir, "test_cache.jsonl")

	cacheConfig := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      cacheFilePath,
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Test comment")

	// モックMessageSenderを使用
	slackSender := newMockMessageSender("Slack")

	// 実際のFileCacheを使用
	fileCache := cache.NewFileRecommendCache(cacheConfig)

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		cacheConfig,
		testSenderFactory([]domain.MessageSender{slackSender}),
		testCacheFactory(fileCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 1回目: 1件目の記事を処理
	err = runner.Run(context.Background(), params, profile)
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender.GetReceivedCount())

	// 2回目: 2件目の記事を処理（新しいキャッシュインスタンスを使用）
	fileCache2 := cache.NewFileRecommendCache(cacheConfig)
	slackSender2 := newMockMessageSender("Slack")
	stderrBuffer2 := new(bytes.Buffer)
	stdoutBuffer2 := new(bytes.Buffer)

	runner2, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer2,
		stdoutBuffer2,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		cacheConfig,
		testSenderFactory([]domain.MessageSender{slackSender2}),
		testCacheFactory(fileCache2),
	)
	require.NoError(t, err)

	err = runner2.Run(context.Background(), params, profile)
	assert.NoError(t, err)
	assert.Equal(t, 1, slackSender2.GetReceivedCount())

	// 3回目: 全ての記事がキャッシュ済み
	fileCache3 := cache.NewFileRecommendCache(cacheConfig)
	slackSender3 := newMockMessageSender("Slack")
	stderrBuffer3 := new(bytes.Buffer)
	stdoutBuffer3 := new(bytes.Buffer)

	runner3, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer3,
		stdoutBuffer3,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		cacheConfig,
		testSenderFactory([]domain.MessageSender{slackSender3}),
		testCacheFactory(fileCache3),
	)
	require.NoError(t, err)

	err = runner3.Run(context.Background(), params, profile)
	// 全ての記事がキャッシュ済みの場合はエラーなしで正常終了
	assert.NoError(t, err)
	assert.Equal(t, 0, slackSender3.GetReceivedCount(), "全記事キャッシュ済みの場合はメッセージが送信されないべき")
	assert.Contains(t, stdoutBuffer3.String(), "新しい記事が見つかりませんでした", "キャッシュ済みメッセージが出力されるべき")
}

func TestRecommendRunner_Integration_NoSenders(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("No sender test comment")

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	// senderなしで実行
	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{}), // 空のsenders
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証: senderがなくてもエラーにならない
	assert.NoError(t, err)
	assert.Contains(t, stdoutBuffer.String(), "外部サービスへの投稿は設定されていません")
}

func TestRecommendRunner_Integration_WithFixedMessage(t *testing.T) {
	t.Parallel()

	// RSSフィードモックサーバーをセットアップ
	feedServer := httptest.NewServer(mockRSSHandler())
	defer feedServer.Close()

	// 実際のFetchClient（infra層）を使用
	fetchClient := fetch.NewFetchClient()

	// モックRecommenderを使用
	recommender := newMockRecommender("Test comment with fixed message")

	// 固定メッセージを受け取ったか確認するためのカスタムsender
	fixedMsgReceived := ""
	customSender := &mockMessageSenderWithFixedMsg{
		serviceName:    "CustomService",
		fixedMsgHolder: &fixedMsgReceived,
	}

	// NopCacheを使用
	nopCache := cache.NewNopCache()

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	runner, err := app.NewRecommendRunner(
		fetchClient,
		recommender,
		stderrBuffer,
		stdoutBuffer,
		&entity.OutputConfig{},
		&entity.PromptConfig{CommentPromptTemplate: "test-template"},
		nil,
		testSenderFactory([]domain.MessageSender{customSender}),
		testCacheFactory(nopCache),
	)
	require.NoError(t, err)

	params := &app.RecommendParams{
		URLs: []string{feedServer.URL},
	}
	profile := &entity.Profile{
		Prompt: &entity.PromptConfig{
			FixedMessage: "This is a fixed message",
		},
	}

	// 実行
	err = runner.Run(context.Background(), params, profile)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, "This is a fixed message", fixedMsgReceived, "固定メッセージが正しく渡されるべき")
}

// mockMessageSenderWithFixedMsg はfixedMessageを記録するMessageSender
type mockMessageSenderWithFixedMsg struct {
	serviceName    string
	fixedMsgHolder *string
}

func (s *mockMessageSenderWithFixedMsg) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	*s.fixedMsgHolder = fixedMessage
	return nil
}

func (s *mockMessageSenderWithFixedMsg) ServiceName() string {
	return s.serviceName
}

// Verify that mockMessageSender implements domain.MessageSender interface
var _ domain.MessageSender = (*mockMessageSender)(nil)
var _ domain.MessageSender = (*mockMessageSenderWithFixedMsg)(nil)

// Verify that mockRecommender implements domain.Recommender interface
var _ domain.Recommender = (*mockRecommender)(nil)

// testutilを使用してStringPtrをインポートする（未使用警告回避のため）
var _ = testutil.StringPtr
