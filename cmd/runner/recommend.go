package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/cache"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra/message"
	"github.com/slack-go/slack"
)

// ErrNoArticlesFound は記事が見つからなかった場合のsentinel error
var ErrNoArticlesFound = errors.New("no articles found in the feed")

// RecommendParams はrecommendコマンドの実行パラメータを表す構造体
type RecommendParams struct {
	URLs []string
}

// RecommendRunner はrecommendコマンドのビジネスロジックを実行する構造体
type RecommendRunner struct {
	fetcher     *domain.Fetcher
	recommender domain.Recommender
	viewers     []domain.MessageSender
	cache       domain.RecommendCache
	stderr      io.Writer
	stdout      io.Writer
}

// NewRecommendRunner はRecommendRunnerの新しいインスタンスを作成する
func NewRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stderr io.Writer, stdout io.Writer, outputConfig *entity.OutputConfig, promptConfig *entity.PromptConfig, cacheConfig *entity.CacheConfig) (*RecommendRunner, error) {
	fetcher := domain.NewFetcher(
		fetchClient,
		func(url string, err error) error {
			fmt.Fprintf(stderr, "エラー: フィードの取得に失敗しました: %s\n", url)
			fmt.Fprintln(stderr, "フィードのURLが正しいか確認してください。サイトが一時的に利用できない可能性もあります。")
			slog.Error("Failed to fetch feed", "url", url, "error", err)
			return err
		},
	)
	var viewers []domain.MessageSender

	if outputConfig != nil {
		if outputConfig.SlackAPI != nil {
			// entity.SlackAPIConfigは既にToEntity()変換済みなので直接使用
			slackConfig := outputConfig.SlackAPI
			// enabledフラグのチェック
			if !slackConfig.Enabled {
				slog.Info("Slack API output is disabled (enabled: false)")
			} else {
				// Slackクライアントのオプションを設定
				options := []slack.Option{}
				if slackConfig.APIURL != nil && *slackConfig.APIURL != "" {
					// テスト用：カスタムAPIエンドポイントを設定
					options = append(options, slack.OptionAPIURL(*slackConfig.APIURL))
				}
				slackClient := slack.New(slackConfig.APIToken.Value(), options...)
				slackViewer := message.NewSlackSender(slackConfig, slackClient)
				viewers = append(viewers, slackViewer)
			}
		}
		if outputConfig.Misskey != nil {
			// entity.MisskeyConfigは既にToEntity()変換済みなので直接使用
			misskeyConfig := outputConfig.Misskey
			// enabledフラグのチェック
			if !misskeyConfig.Enabled {
				slog.Info("Misskey output is disabled (enabled: false)")
			} else {
				misskeyViewer, err := message.NewMisskeySender(misskeyConfig.APIURL, misskeyConfig.APIToken.Value(), misskeyConfig.MessageTemplate)
				if err != nil {
					return nil, fmt.Errorf("failed to create Misskey viewer: %w", err)
				}
				viewers = append(viewers, misskeyViewer)
			}
		}
	}

	// キャッシュインスタンスの作成と初期化
	cache, err := createCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	if err := cache.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return &RecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
		cache:       cache,
		stderr:      stderr,
		stdout:      stdout,
	}, nil
}

// createCache は設定に基づいてキャッシュインスタンスを作成する
func createCache(config *entity.CacheConfig) (domain.RecommendCache, error) {
	if config == nil || !config.Enabled {
		return cache.NewNopCache(), nil
	}

	return cache.NewFileRecommendCache(config), nil
}

// selectRandomFeed は利用可能なfeed URLからランダムに1つを選択する
func selectRandomFeed(urls []string, excludedURLs map[string]bool) (string, error) {
	var availableURLs []string
	for _, url := range urls {
		if !excludedURLs[url] {
			availableURLs = append(availableURLs, url)
		}
	}

	if len(availableURLs) == 0 {
		return "", errors.New("no available feeds")
	}

	return availableURLs[rand.IntN(len(availableURLs))], nil
}

// Run はrecommendコマンドのビジネスロジックを実行する
func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile *entity.Profile) error {
	slog.Debug("RecommendRunner.Run parameters", slog.Any("profile", profile))
	slog.Info("Starting recommend command execution")
	slog.Debug("Selecting feed from URLs", "url_count", len(params.URLs))

	// キャッシュのリソース管理
	defer func() {
		if err := r.cache.Close(); err != nil {
			slog.Error("Failed to close cache", "error", err)
		}
	}()

	// 複数feedからの2段階ランダム選択とリトライロジック
	excludedURLs := make(map[string]bool)
	var allArticles []entity.Article
	var selectedURL string

	for attempt := 1; attempt <= len(params.URLs); attempt++ {
		// 進行状況メッセージ: フィード選択
		if attempt == 1 {
			fmt.Fprintln(r.stderr, "フィードを選択しています...")
		} else {
			fmt.Fprintf(r.stderr, "別のフィードで再試行しています... (%d/%d)\n", attempt, len(params.URLs))
		}

		// Step 1: ランダムに1つのfeedを選択
		var err error
		selectedURL, err = selectRandomFeed(params.URLs, excludedURLs)
		if err != nil {
			slog.Error("Failed to select feed", "error", err, "attempt", attempt, "total_feeds", len(params.URLs))
			break
		}

		slog.Debug("Selected feed for articles fetch", "url", selectedURL, "attempt", attempt)

		// 進行状況メッセージ: フィード取得
		fmt.Fprintf(r.stderr, "フィードを取得しています... (%sから)\n", selectedURL)

		// Step 2: 選択されたfeedから記事を取得
		allArticles, err = r.fetcher.Fetch([]string{selectedURL}, 0)

		// エラーまたは記事0件の場合は失敗として次のフィードを試す
		var shouldRetry bool
		var logMessage string
		if err != nil {
			shouldRetry = true
			logMessage = "Failed to fetch from feed, retrying with another feed"
			slog.Warn(logMessage,
				"url", selectedURL,
				"error", err.Error(),
				"attempt", attempt,
				"total_feeds", len(params.URLs))
		} else if len(allArticles) == 0 {
			shouldRetry = true
			logMessage = "No articles found in feed, retrying with another feed"
			slog.Warn(logMessage,
				"url", selectedURL,
				"attempt", attempt,
				"total_feeds", len(params.URLs))
		}

		if shouldRetry {
			excludedURLs[selectedURL] = true
			continue
		}

		// 成功した場合
		// 進行状況メッセージ: 記事解析
		fmt.Fprintf(r.stderr, "記事を解析しています... (%d件の記事を発見)\n", len(allArticles))
		slog.Info("Successfully fetched articles from feed", "url", selectedURL, "article_count", len(allArticles))
		break
	}

	// 全feedが失敗した場合
	if len(allArticles) == 0 {
		var triedURLs []string
		for url := range excludedURLs {
			triedURLs = append(triedURLs, url)
		}
		slog.Error("Failed to fetch from all feeds", "tried_urls", triedURLs)
		return ErrNoArticlesFound
	}

	// 記事の重複チェックとフィルタリング
	fmt.Fprintln(r.stderr, "記事の重複をチェックしています...")
	var uniqueArticles []entity.Article
	duplicateCount := 0

	for _, article := range allArticles {
		if r.cache.IsCached(article.Link) {
			duplicateCount++
			slog.Debug("Article is already cached, skipping", "url", article.Link, "title", article.Title)
		} else {
			uniqueArticles = append(uniqueArticles, article)
		}
	}

	slog.Info("Cache duplicate check completed",
		"total_articles", len(allArticles),
		"duplicate_articles", duplicateCount,
		"unique_articles", len(uniqueArticles))

	// 全ての記事が重複している場合
	if len(uniqueArticles) == 0 {
		fmt.Fprintln(r.stderr, "すべての記事が既にキャッシュされています")
		slog.Info("All articles are already cached")
		fmt.Fprintln(r.stdout, "新しい記事が見つかりませんでした。すべて投稿済みの記事です。")
		return nil
	}

	fmt.Fprintf(r.stderr, "%d件の新しい記事が見つかりました\n", len(uniqueArticles))

	// 進行状況メッセージ: 記事選定とコメント生成の開始
	fmt.Fprintln(r.stderr, "記事選定とコメント生成を行なっています...")

	slog.Debug("Generating recommendation from unique articles", "unique_article_count", len(uniqueArticles))
	recommend, err := r.recommender.Recommend(ctx, uniqueArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	// 進行状況メッセージ: 記事選定とコメント生成の完了
	fmt.Fprintf(r.stderr, "記事選定とコメント生成が完了しました: %s\n", recommend.Article.Title)

	// AIが生成したコメントをユーザーに表示
	if recommend.Comment != nil && *recommend.Comment != "" {
		fmt.Fprintf(r.stdout, "\nAIコメント:\n%s\n", *recommend.Comment)
	}

	// 記事URLを表示
	fmt.Fprintf(r.stdout, "\n記事URL: %s\n", recommend.Article.Link)

	slog.Debug("Recommendation generated successfully", "article_title", recommend.Article.Title)

	var errs []error
	fixedMessage := ""
	if profile.Prompt != nil {
		fixedMessage = profile.Prompt.FixedMessage
	}

	// 推薦記事の詳細情報をログ出力
	var commentValue string
	if recommend.Comment != nil {
		commentValue = *recommend.Comment
	}
	slog.Info("Recommendation article selected",
		"title", recommend.Article.Title,
		"link", recommend.Article.Link,
		"comment", commentValue,
		"fixed_message", fixedMessage,
	)

	slog.Debug("Sending recommendation to viewers", "viewer_count", len(r.viewers))

	// 外部サービスへの投稿状況をメッセージ表示
	if len(r.viewers) > 0 {
		fmt.Fprintln(r.stdout, "\n外部サービスに投稿しています...")
	} else {
		fmt.Fprintln(r.stdout, "\n外部サービスへの投稿は設定されていません")
	}

	for _, viewer := range r.viewers {
		var serviceName string
		isKnownService := true
		switch viewer.(type) {
		case *message.SlackSender:
			serviceName = "Slack"
		case *message.MisskeySender:
			serviceName = "Misskey"
		default:
			isKnownService = false
		}

		if isKnownService {
			fmt.Fprintf(r.stdout, "%sに投稿中...\n", serviceName)
		}

		if viewErr := viewer.SendRecommend(recommend, fixedMessage); viewErr != nil {
			if isKnownService {
				fmt.Fprintf(r.stdout, "%s投稿でエラーが発生しました: %v\n", serviceName, viewErr)
			} else {
				fmt.Fprintf(r.stdout, "投稿でエラーが発生しました: %v\n", viewErr)
			}
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
		} else {
			if isKnownService {
				fmt.Fprintf(r.stdout, "%sに投稿しました\n", serviceName)
			}
		}
	}

	// 投稿エラーがある場合はキャッシュを更新せずに終了
	if len(errs) > 0 {
		slog.Warn("Some posts failed, not updating cache", "error_count", len(errs))
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	// 全ての投稿が成功した場合のみキャッシュを更新
	fmt.Fprintln(r.stderr, "投稿履歴をキャッシュに保存しています...")
	if err := r.cache.AddEntry(recommend.Article.Link, recommend.Article.Title); err != nil {
		slog.Error("Failed to update cache", "url", recommend.Article.Link, "title", recommend.Article.Title, "error", err)
		// キャッシュ更新の失敗は致命的エラーとしない（投稿は成功しているため）
		fmt.Fprintf(r.stderr, "警告: キャッシュの更新に失敗しましたが、投稿は完了しました\n")
	} else {
		slog.Info("Cache updated successfully", "url", recommend.Article.Link, "title", recommend.Article.Title)
		fmt.Fprintln(r.stderr, "キャッシュを更新しました")
	}

	slog.Info("Recommendation sent successfully to all viewers")
	return nil
}
