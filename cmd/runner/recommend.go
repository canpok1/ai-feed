package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"
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
}

// NewRecommendRunner はRecommendRunnerの新しいインスタンスを作成する
func NewRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer, outputConfig *infra.OutputConfig, promptConfig *infra.PromptConfig) (*RecommendRunner, error) {
	fetcher := domain.NewFetcher(
		fetchClient,
		func(url string, err error) error {
			fmt.Fprintf(stderr, "Error fetching feed from %s: %v\n", url, err)
			return err
		},
	)
	viewer, err := message.NewStdSender(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create viewer: %w", err)
	}
	viewers := []domain.MessageSender{viewer}

	if outputConfig != nil {
		if outputConfig.SlackAPI != nil {
			slackConfig, err := outputConfig.SlackAPI.ToEntity()
			if err != nil {
				return nil, fmt.Errorf("failed to process Slack API config: %w", err)
			}
			slackViewer := message.NewSlackSender(slackConfig)
			viewers = append(viewers, slackViewer)
		}
		if outputConfig.Misskey != nil {
			misskeyConfig, err := outputConfig.Misskey.ToEntity()
			if err != nil {
				return nil, fmt.Errorf("failed to process Misskey config: %w", err)
			}
			misskeyViewer, err := message.NewMisskeySender(misskeyConfig.APIURL, misskeyConfig.APIToken, misskeyConfig.MessageTemplate)
			if err != nil {
				return nil, fmt.Errorf("failed to create Misskey viewer: %w", err)
			}
			viewers = append(viewers, misskeyViewer)
		}
	}

	return &RecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
	}, nil
}

// Run はrecommendコマンドのビジネスロジックを実行する
func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile infra.Profile) error {
	slog.Info("Starting recommend command execution")
	slog.Debug("Fetching articles from URLs", "url_count", len(params.URLs))

	allArticles, err := r.fetcher.Fetch(params.URLs, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch articles: %w", err)
	}

	if len(allArticles) == 0 {
		slog.Warn("No articles found in feeds")
		return ErrNoArticlesFound
	}

	slog.Debug("Articles fetched successfully", "article_count", len(allArticles))

	slog.Debug("Generating recommendation from articles")
	recommend, err := r.recommender.Recommend(ctx, allArticles)
	if err != nil {
		slog.Error("Failed to generate recommendation", "error", err)
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	slog.Debug("Recommendation generated successfully", "article_title", recommend.Article.Title)

	var errs []error
	fixedMessage := ""
	if profile.Prompt != nil {
		fixedMessage = profile.Prompt.FixedMessage
	}

	slog.Debug("Sending recommendation to viewers", "viewer_count", len(r.viewers))
	for _, viewer := range r.viewers {
		if viewErr := viewer.SendRecommend(recommend, fixedMessage); viewErr != nil {
			slog.Error("Failed to send recommendation to viewer", "error", viewErr)
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
		}
	}

	if len(errs) > 0 {
		slog.Error("Some viewers failed to send recommendation", "error_count", len(errs))
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	slog.Info("Recommendation sent successfully to all viewers")
	return nil
}
