package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra/message"
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
	stderr      io.Writer
	stdout      io.Writer
}

// NewRecommendRunner はRecommendRunnerの新しいインスタンスを作成する
func NewRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stderr io.Writer, stdout io.Writer, outputConfig *entity.OutputConfig, promptConfig *entity.PromptConfig) (*RecommendRunner, error) {
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
				slackViewer := message.NewSlackSender(slackConfig)
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
				misskeyViewer, err := message.NewMisskeySender(misskeyConfig.APIURL, misskeyConfig.APIToken, misskeyConfig.MessageTemplate)
				if err != nil {
					return nil, fmt.Errorf("failed to create Misskey viewer: %w", err)
				}
				viewers = append(viewers, misskeyViewer)
			}
		}
	}

	return &RecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
		stderr:      stderr,
		stdout:      stdout,
	}, nil
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
	slog.Info("Starting recommend command execution")
	slog.Debug("Selecting feed from URLs", "url_count", len(params.URLs))

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

	// 進行状況メッセージ: AI推薦生成
	fmt.Fprintln(r.stderr, "推薦記事を生成しています...")

	slog.Debug("Generating recommendation from articles")
	recommend, err := r.recommender.Recommend(ctx, allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	// 完了メッセージ: 推薦完了（stdout）
	fmt.Fprintf(r.stdout, "推薦記事を生成しました: %s\n", recommend.Article.Title)

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
	for _, viewer := range r.viewers {
		if viewErr := viewer.SendRecommend(recommend, fixedMessage); viewErr != nil {
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	slog.Info("Recommendation sent successfully to all viewers")
	return nil
}
